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
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/name"
	"github.com/google/go-containerregistry/v1/remote"
	"github.com/knative/pkg/controller"
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
	extv1beta1informers "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"

	extlisters "k8s.io/client-go/listers/extensions/v1beta1"

	warmimagev2 "github.com/mattmoor/warm-image/pkg/apis/warmimage/v2"
	clientset "github.com/mattmoor/warm-image/pkg/client/clientset/versioned"
	warmimagescheme "github.com/mattmoor/warm-image/pkg/client/clientset/versioned/scheme"
	informers "github.com/mattmoor/warm-image/pkg/client/informers/externalversions/warmimage/v2"
	listers "github.com/mattmoor/warm-image/pkg/client/listers/warmimage/v2"
)

const controllerAgentName = "warmimage-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a WarmImage is synced
	SuccessSynced = "Synced"

	// MessageResourceSynced is the message used for an Event fired when a WarmImage
	// is synced successfully
	MessageResourceSynced = "WarmImage synced successfully"
)

// Reconciler is the controller implementation for WarmImage resources
type Reconciler struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// warmimageclientset is a clientset for our own API group
	warmimageclientset clientset.Interface

	daemonsetsLister extlisters.DaemonSetLister
	warmimagesLister listers.WarmImageLister

	sleeperImage string

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

// Check that we implement the controller.Reconciler interface.
var _ controller.Reconciler = (*Reconciler)(nil)

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
	daemonsetInformer extv1beta1informers.DaemonSetInformer,
	warmimageInformer informers.WarmImageInformer,
	sleeperImage string,
) *controller.Impl {

	// Enrich the logs with controller name
	logger = logger.Named(controllerAgentName).With(zap.String(logkey.ControllerType, controllerAgentName))

	logger.Debug("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.Named("event-broadcaster").Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{
		Interface: kubeclientset.CoreV1().Events(""),
	})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	r := &Reconciler{
		kubeclientset:      kubeclientset,
		warmimageclientset: warmimageclientset,
		daemonsetsLister:   daemonsetInformer.Lister(),
		warmimagesLister:   warmimageInformer.Lister(),
		sleeperImage:       sleeperImage,
		recorder:           recorder,
		Logger:             logger,
	}
	impl := controller.NewImpl(r, logger, "WarmImages")

	logger.Info("Setting up event handlers")
	// Set up an event handler for when WarmImage resources change
	warmimageInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    impl.Enqueue,
		UpdateFunc: controller.PassNew(impl.Enqueue),
	})

	return impl
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

// TODO(mattmoor): Move to resources subpackage.
func newDaemonSet(client kubernetes.Interface, wi *warmimagev2.WarmImage, sleeperImage string) (*extv1beta1.DaemonSet, error) {
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
			// TODO(mattmoor): Use shared utility for this.
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
						Image:           sleeperImage,
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
								corev1.ResourceMemory: resource.MustParse("20M"),
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
								corev1.ResourceMemory: resource.MustParse("20M"),
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

// Reconcile implements controller.Reconciler
func (c *Reconciler) Reconcile(ctx context.Context, key string) error {
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
		ds, err := newDaemonSet(c.kubeclientset, warmimage, c.sleeperImage)
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
