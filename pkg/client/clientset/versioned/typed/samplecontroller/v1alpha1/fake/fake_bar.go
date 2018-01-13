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
	v1alpha1 "github.com/mattmoor/warm-image/pkg/apis/samplecontroller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBars implements BarInterface
type FakeBars struct {
	Fake *FakeSamplecontrollerV1alpha1
	ns   string
}

var barsResource = schema.GroupVersionResource{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Resource: "bars"}

var barsKind = schema.GroupVersionKind{Group: "samplecontroller.k8s.io", Version: "v1alpha1", Kind: "Bar"}

// Get takes name of the bar, and returns the corresponding bar object, and an error if there is any.
func (c *FakeBars) Get(name string, options v1.GetOptions) (result *v1alpha1.Bar, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(barsResource, c.ns, name), &v1alpha1.Bar{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Bar), err
}

// List takes label and field selectors, and returns the list of Bars that match those selectors.
func (c *FakeBars) List(opts v1.ListOptions) (result *v1alpha1.BarList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(barsResource, barsKind, c.ns, opts), &v1alpha1.BarList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.BarList{}
	for _, item := range obj.(*v1alpha1.BarList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bars.
func (c *FakeBars) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(barsResource, c.ns, opts))

}

// Create takes the representation of a bar and creates it.  Returns the server's representation of the bar, and an error, if there is any.
func (c *FakeBars) Create(bar *v1alpha1.Bar) (result *v1alpha1.Bar, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(barsResource, c.ns, bar), &v1alpha1.Bar{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Bar), err
}

// Update takes the representation of a bar and updates it. Returns the server's representation of the bar, and an error, if there is any.
func (c *FakeBars) Update(bar *v1alpha1.Bar) (result *v1alpha1.Bar, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(barsResource, c.ns, bar), &v1alpha1.Bar{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Bar), err
}

// Delete takes name of the bar and deletes it. Returns an error if one occurs.
func (c *FakeBars) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(barsResource, c.ns, name), &v1alpha1.Bar{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBars) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(barsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.BarList{})
	return err
}

// Patch applies the patch and returns the patched bar.
func (c *FakeBars) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Bar, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(barsResource, c.ns, name, data, subresources...), &v1alpha1.Bar{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Bar), err
}
