/*
Copyright 2022 The open-podcasts Authors. All rights reserved.

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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/opensource-f2f/open-podcasts/api/osf2f.my.domain/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeShows implements ShowInterface
type FakeShows struct {
	Fake *FakeOsf2fV1alpha1
	ns   string
}

var showsResource = schema.GroupVersionResource{Group: "osf2f.my.domain", Version: "v1alpha1", Resource: "shows"}

var showsKind = schema.GroupVersionKind{Group: "osf2f.my.domain", Version: "v1alpha1", Kind: "Show"}

// Get takes name of the show, and returns the corresponding show object, and an error if there is any.
func (c *FakeShows) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Show, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(showsResource, c.ns, name), &v1alpha1.Show{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Show), err
}

// List takes label and field selectors, and returns the list of Shows that match those selectors.
func (c *FakeShows) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ShowList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(showsResource, showsKind, c.ns, opts), &v1alpha1.ShowList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ShowList{ListMeta: obj.(*v1alpha1.ShowList).ListMeta}
	for _, item := range obj.(*v1alpha1.ShowList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested shows.
func (c *FakeShows) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(showsResource, c.ns, opts))

}

// Create takes the representation of a show and creates it.  Returns the server's representation of the show, and an error, if there is any.
func (c *FakeShows) Create(ctx context.Context, show *v1alpha1.Show, opts v1.CreateOptions) (result *v1alpha1.Show, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(showsResource, c.ns, show), &v1alpha1.Show{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Show), err
}

// Update takes the representation of a show and updates it. Returns the server's representation of the show, and an error, if there is any.
func (c *FakeShows) Update(ctx context.Context, show *v1alpha1.Show, opts v1.UpdateOptions) (result *v1alpha1.Show, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(showsResource, c.ns, show), &v1alpha1.Show{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Show), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeShows) UpdateStatus(ctx context.Context, show *v1alpha1.Show, opts v1.UpdateOptions) (*v1alpha1.Show, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(showsResource, "status", c.ns, show), &v1alpha1.Show{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Show), err
}

// Delete takes name of the show and deletes it. Returns an error if one occurs.
func (c *FakeShows) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(showsResource, c.ns, name, opts), &v1alpha1.Show{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeShows) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(showsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ShowList{})
	return err
}

// Patch applies the patch and returns the patched show.
func (c *FakeShows) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Show, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(showsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Show{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Show), err
}
