/*
Copyright 2018 The Kubernetes Authors.

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
package v2

import (
	v2 "github.com/mattmoor/warm-image/pkg/apis/warmimage/v2"
	scheme "github.com/mattmoor/warm-image/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// WarmImagesGetter has a method to return a WarmImageInterface.
// A group's client should implement this interface.
type WarmImagesGetter interface {
	WarmImages(namespace string) WarmImageInterface
}

// WarmImageInterface has methods to work with WarmImage resources.
type WarmImageInterface interface {
	Create(*v2.WarmImage) (*v2.WarmImage, error)
	Update(*v2.WarmImage) (*v2.WarmImage, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v2.WarmImage, error)
	List(opts v1.ListOptions) (*v2.WarmImageList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.WarmImage, err error)
	WarmImageExpansion
}

// warmImages implements WarmImageInterface
type warmImages struct {
	client rest.Interface
	ns     string
}

// newWarmImages returns a WarmImages
func newWarmImages(c *MattmoorV2Client, namespace string) *warmImages {
	return &warmImages{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the warmImage, and returns the corresponding warmImage object, and an error if there is any.
func (c *warmImages) Get(name string, options v1.GetOptions) (result *v2.WarmImage, err error) {
	result = &v2.WarmImage{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("warmimages").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of WarmImages that match those selectors.
func (c *warmImages) List(opts v1.ListOptions) (result *v2.WarmImageList, err error) {
	result = &v2.WarmImageList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("warmimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested warmImages.
func (c *warmImages) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("warmimages").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a warmImage and creates it.  Returns the server's representation of the warmImage, and an error, if there is any.
func (c *warmImages) Create(warmImage *v2.WarmImage) (result *v2.WarmImage, err error) {
	result = &v2.WarmImage{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("warmimages").
		Body(warmImage).
		Do().
		Into(result)
	return
}

// Update takes the representation of a warmImage and updates it. Returns the server's representation of the warmImage, and an error, if there is any.
func (c *warmImages) Update(warmImage *v2.WarmImage) (result *v2.WarmImage, err error) {
	result = &v2.WarmImage{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("warmimages").
		Name(warmImage.Name).
		Body(warmImage).
		Do().
		Into(result)
	return
}

// Delete takes name of the warmImage and deletes it. Returns an error if one occurs.
func (c *warmImages) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("warmimages").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *warmImages) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("warmimages").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched warmImage.
func (c *warmImages) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.WarmImage, err error) {
	result = &v2.WarmImage{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("warmimages").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
