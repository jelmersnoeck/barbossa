// MIT License
//
// Copyright (c) 2018 Jelmer Snoeck
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/jelmersnoeck/barbossa/apis/barbossa/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeHighAvailabilityPolicies implements HighAvailabilityPolicyInterface
type FakeHighAvailabilityPolicies struct {
	Fake *FakeBarbossaV1alpha1
	ns   string
}

var highavailabilitypoliciesResource = schema.GroupVersionResource{Group: "barbossa.sphc.io", Version: "v1alpha1", Resource: "highavailabilitypolicies"}

var highavailabilitypoliciesKind = schema.GroupVersionKind{Group: "barbossa.sphc.io", Version: "v1alpha1", Kind: "HighAvailabilityPolicy"}

// Get takes name of the highAvailabilityPolicy, and returns the corresponding highAvailabilityPolicy object, and an error if there is any.
func (c *FakeHighAvailabilityPolicies) Get(name string, options v1.GetOptions) (result *v1alpha1.HighAvailabilityPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(highavailabilitypoliciesResource, c.ns, name), &v1alpha1.HighAvailabilityPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HighAvailabilityPolicy), err
}

// List takes label and field selectors, and returns the list of HighAvailabilityPolicies that match those selectors.
func (c *FakeHighAvailabilityPolicies) List(opts v1.ListOptions) (result *v1alpha1.HighAvailabilityPolicyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(highavailabilitypoliciesResource, highavailabilitypoliciesKind, c.ns, opts), &v1alpha1.HighAvailabilityPolicyList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.HighAvailabilityPolicyList{}
	for _, item := range obj.(*v1alpha1.HighAvailabilityPolicyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested highAvailabilityPolicies.
func (c *FakeHighAvailabilityPolicies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(highavailabilitypoliciesResource, c.ns, opts))

}

// Create takes the representation of a highAvailabilityPolicy and creates it.  Returns the server's representation of the highAvailabilityPolicy, and an error, if there is any.
func (c *FakeHighAvailabilityPolicies) Create(highAvailabilityPolicy *v1alpha1.HighAvailabilityPolicy) (result *v1alpha1.HighAvailabilityPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(highavailabilitypoliciesResource, c.ns, highAvailabilityPolicy), &v1alpha1.HighAvailabilityPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HighAvailabilityPolicy), err
}

// Update takes the representation of a highAvailabilityPolicy and updates it. Returns the server's representation of the highAvailabilityPolicy, and an error, if there is any.
func (c *FakeHighAvailabilityPolicies) Update(highAvailabilityPolicy *v1alpha1.HighAvailabilityPolicy) (result *v1alpha1.HighAvailabilityPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(highavailabilitypoliciesResource, c.ns, highAvailabilityPolicy), &v1alpha1.HighAvailabilityPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HighAvailabilityPolicy), err
}

// Delete takes name of the highAvailabilityPolicy and deletes it. Returns an error if one occurs.
func (c *FakeHighAvailabilityPolicies) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(highavailabilitypoliciesResource, c.ns, name), &v1alpha1.HighAvailabilityPolicy{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHighAvailabilityPolicies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(highavailabilitypoliciesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.HighAvailabilityPolicyList{})
	return err
}

// Patch applies the patch and returns the patched highAvailabilityPolicy.
func (c *FakeHighAvailabilityPolicies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.HighAvailabilityPolicy, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(highavailabilitypoliciesResource, c.ns, name, data, subresources...), &v1alpha1.HighAvailabilityPolicy{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HighAvailabilityPolicy), err
}