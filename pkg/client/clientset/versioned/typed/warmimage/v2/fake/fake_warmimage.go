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
package fake

import (
	v2 "github.com/mattmoor/warm-image/pkg/apis/warmimage/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeWarmImages implements WarmImageInterface
type FakeWarmImages struct {
	Fake *FakeMattmoorV2
	ns   string
}

var warmimagesResource = schema.GroupVersionResource{Group: "mattmoor.io", Version: "v2", Resource: "warmimages"}

var warmimagesKind = schema.GroupVersionKind{Group: "mattmoor.io", Version: "v2", Kind: "WarmImage"}

// Get takes name of the warmImage, and returns the corresponding warmImage object, and an error if there is any.
func (c *FakeWarmImages) Get(name string, options v1.GetOptions) (result *v2.WarmImage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(warmimagesResource, c.ns, name), &v2.WarmImage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.WarmImage), err
}

// List takes label and field selectors, and returns the list of WarmImages that match those selectors.
func (c *FakeWarmImages) List(opts v1.ListOptions) (result *v2.WarmImageList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(warmimagesResource, warmimagesKind, c.ns, opts), &v2.WarmImageList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v2.WarmImageList{}
	for _, item := range obj.(*v2.WarmImageList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested warmImages.
func (c *FakeWarmImages) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(warmimagesResource, c.ns, opts))

}

// Create takes the representation of a warmImage and creates it.  Returns the server's representation of the warmImage, and an error, if there is any.
func (c *FakeWarmImages) Create(warmImage *v2.WarmImage) (result *v2.WarmImage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(warmimagesResource, c.ns, warmImage), &v2.WarmImage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.WarmImage), err
}

// Update takes the representation of a warmImage and updates it. Returns the server's representation of the warmImage, and an error, if there is any.
func (c *FakeWarmImages) Update(warmImage *v2.WarmImage) (result *v2.WarmImage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(warmimagesResource, c.ns, warmImage), &v2.WarmImage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.WarmImage), err
}

// Delete takes name of the warmImage and deletes it. Returns an error if one occurs.
func (c *FakeWarmImages) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(warmimagesResource, c.ns, name), &v2.WarmImage{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeWarmImages) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(warmimagesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v2.WarmImageList{})
	return err
}

// Patch applies the patch and returns the patched warmImage.
func (c *FakeWarmImages) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v2.WarmImage, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(warmimagesResource, c.ns, name, data, subresources...), &v2.WarmImage{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v2.WarmImage), err
}
