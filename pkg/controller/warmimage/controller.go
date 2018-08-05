/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package warmimage

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-containerregistry/name"
	"github.com/google/go-containerregistry/v1/remote"
	"github.com/knative/pkg/logging/logkey"
	"github.com/mattmoor/k8schain"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	"github.com/mattmoor/warm-image/pkg/controller"

	extlisters "k8s.io/client-go/listers/extensions/v1beta1"

	warmimagev2 "github.com/mattmoor/warm-image/pkg/apis/warmimage/v2"
	clientset "github.com/mattmoor/warm-image/pkg/client/clientset/versioned"
	warmimagescheme "github.com/mattmoor/warm-image/pkg/client/clientset/versioned/scheme"
	informers "github.com/mattmoor/warm-image/pkg/client/informers/externalversions"
	listers "github.com/mattmoor/warm-image/pkg/client/listers/warmimage/v2"
)

const controllerAgentName = "warmimage-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a WarmImage is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a WarmImage fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceSynced is the message used for an Event fired when a WarmImage
	// is synced successfully
	MessageResourceSynced = "WarmImage synced successfully"
)

var (
	sleeper = flag.String("sleeper", "", "The name of the sleeper image, see //cmd/sleeper")
)

// Controller is the controller implementation for WarmImage resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// warmimageclientset is a clientset for our own API group
	warmimageclientset clientset.Interface

	daemonsetsLister extlisters.DaemonSetLister
	daemonsetsSynced cache.InformerSynced

	warmimagesLister listers.WarmImageLister
	warmimagesSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	// Sugared logger is easier to use but is not as performant as the
	// raw logger. In performance critical paths, call logger.Desugar()
	// and use the returned raw logger instead. In addition to the
	// performance benefits, raw logger also preserves type-safety at
	// the expense of slightly greater verbosity.
	Logger *zap.SugaredLogger
}

func init() {
	// Create event broadcaster
	// Add warmimage-controller types to the default Kubernetes Scheme so Events can be
	// logged for warmimage-controller types.
	warmimagescheme.AddToScheme(scheme.Scheme)
}

// NewController returns a new warmimage controller
func NewController(
	logger *zap.SugaredLogger,
	kubeclientset kubernetes.Interface,
	warmimageclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	warmimageInformerFactory informers.SharedInformerFactory) controller.Interface {

	// Enrich the logs with controller name
	logger = logger.Named(controllerAgentName).With(zap.String(logkey.ControllerType, controllerAgentName))

	// obtain a reference to a shared index informer for the WarmImage type.
	daemonsetInformer := kubeInformerFactory.Extensions().V1beta1().DaemonSets()
	warmimageInformer := warmimageInformerFactory.Mattmoor().V2().WarmImages()

	logger.Debug("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Named("event-broadcaster").Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeclientset.CoreV1().Events(""),
	})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:      kubeclientset,
		warmimageclientset: warmimageclientset,
		daemonsetsLister:   daemonsetInformer.Lister(),
		daemonsetsSynced:   daemonsetInformer.Informer().HasSynced,
		warmimagesLister:   warmimageInformer.Lister(),
		warmimagesSynced:   warmimageInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "WarmImages"),
		recorder:           recorder,
		Logger:             logger,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when WarmImage resources change
	warmimageInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueWarmImage,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueWarmImage(new)
		},
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	c.Logger.Info("Starting WarmImage controller")

	// Wait for the caches to be synced before starting workers
	c.Logger.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.warmimagesSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.Logger.Info("Starting workers")
	// Launch two workers to process WarmImage resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.Logger.Info("Started workers")
	<-stopCh
	c.Logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// WarmImage resource to be synced.
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		c.Logger.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// enqueueWarmImage takes a WarmImage resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than WarmImage.
func (c *Controller) enqueueWarmImage(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func labelsForDaemonSet(wi *warmimagev2.WarmImage) map[string]string {
	return map[string]string{
		"controller": string(wi.UID),
		"version":    wi.ResourceVersion,
	}
}

func oldVersionLabelSelector(wi *warmimagev2.WarmImage) string {
	return fmt.Sprintf("controller=%s,version!=%s", wi.UID, wi.ResourceVersion)
}

func resolveTag(client kubernetes.Interface, t string, opt k8schain.Options) (string, error) {
	tag, err := name.NewTag(t, name.WeakValidation)
	if err != nil {
		// If we fail to parse it as a tag, simply return the input string (it is likely a digest)
		return t, nil
	}

	kc, err := k8schain.New(client, opt)
	if err != nil {
		return "", err
	}

	auth, err := kc.Resolve(tag.Registry)
	if err != nil {
		return "", err
	}

	img, err := remote.Image(tag, auth, http.DefaultTransport)
	if err != nil {
		return "", err
	}
	digest, err := img.Digest()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s@%s", tag.Repository.String(), digest), nil
}

func newDaemonSet(client kubernetes.Interface, wi *warmimagev2.WarmImage) (*extv1beta1.DaemonSet, error) {
	opt := k8schain.Options{
		Namespace: wi.Namespace,
	}
	ips := []corev1.LocalObjectReference{}
	if wi.Spec.ImagePullSecrets != nil {
		ips = append(ips, *wi.Spec.ImagePullSecrets)
		opt.ImagePullSecrets = append(opt.ImagePullSecrets, wi.Spec.ImagePullSecrets.Name)
	}
	img, err := resolveTag(client, wi.Spec.Image, opt)
	if err != nil {
		return nil, err
	}
	return &extv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: wi.Name,
			Labels:       labelsForDaemonSet(wi),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(wi, schema.GroupVersionKind{
					Group:   warmimagev2.SchemeGroupVersion.Group,
					Version: warmimagev2.SchemeGroupVersion.Version,
					Kind:    "WarmImage",
				}),
			},
		},
		Spec: extv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsForDaemonSet(wi),
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{{
						Name:            "the-sleeper",
						Image:           *sleeper,
						ImagePullPolicy: corev1.PullAlways,
						Args: []string{
							"-mode", "copy",
							"-to", "/drop/sleeper",
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "the-sleeper",
							MountPath: "/drop/",
						}},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1m"),
								corev1.ResourceMemory: resource.MustParse("10M"),
							},
						},
					}},
					Containers: []corev1.Container{{
						Name:            "the-image",
						Image:           img,
						ImagePullPolicy: corev1.PullAlways,
						Command:         []string{"/drop/sleeper"},
						Args:            []string{"-mode", "sleep"},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "the-sleeper",
							MountPath: "/drop/",
						}},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1m"),
								corev1.ResourceMemory: resource.MustParse("10M"),
							},
						},
					}},
					ImagePullSecrets: ips,
					Volumes: []corev1.Volume{{
						Name: "the-sleeper",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
				},
			},
		},
	}, nil
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the WarmImage resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the WarmImage resource with this namespace/name
	warmimage, err := c.warmimagesLister.WarmImages(namespace).Get(name)
	if err != nil {
		// The WarmImage resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("warmimage '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	// Make sure the desired image is warmed up ASAP.
	l := labelsForDaemonSet(warmimage)
	dss, err := c.kubeclientset.ExtensionsV1beta1().DaemonSets(namespace).List(metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: l,
		}),
	})
	if err != nil {
		return err
	}
	switch {
	// If none exist, create one.
	case len(dss.Items) == 0:
		ds, err := newDaemonSet(c.kubeclientset, warmimage)
		if err != nil {
			return err
		}
		ds, err = c.kubeclientset.ExtensionsV1beta1().DaemonSets(namespace).Create(ds)
		if err != nil {
			return err
		}
		c.Logger.Infof("Warming up: %q, with %q", warmimage.Spec.Image, ds.Name)

	// If multiple exist, delete all but one.
	case len(dss.Items) > 1:
		c.Logger.Error("NYI: cleaning up multiple daemonsets for a single WarmImage.")
	}

	// Delete any older versions of this WarmImage.
	propPolicy := metav1.DeletePropagationForeground
	err = c.kubeclientset.ExtensionsV1beta1().DaemonSets(namespace).DeleteCollection(
		&metav1.DeleteOptions{PropagationPolicy: &propPolicy},
		metav1.ListOptions{LabelSelector: oldVersionLabelSelector(warmimage)},
	)
	if err != nil {
		return err
	}

	c.recorder.Event(warmimage, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}
