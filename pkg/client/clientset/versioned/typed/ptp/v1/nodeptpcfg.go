/*
Copyright The Kubernetes Authors.

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

package v1

import (
	"time"

	v1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
	scheme "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NodePTPCfgsGetter has a method to return a NodePTPCfgInterface.
// A group's client should implement this interface.
type NodePTPCfgsGetter interface {
	NodePTPCfgs(namespace string) NodePTPCfgInterface
}

// NodePTPCfgInterface has methods to work with NodePTPCfg resources.
type NodePTPCfgInterface interface {
	Create(*v1.NodePTPCfg) (*v1.NodePTPCfg, error)
	Update(*v1.NodePTPCfg) (*v1.NodePTPCfg, error)
	UpdateStatus(*v1.NodePTPCfg) (*v1.NodePTPCfg, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.NodePTPCfg, error)
	List(opts metav1.ListOptions) (*v1.NodePTPCfgList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.NodePTPCfg, err error)
	NodePTPCfgExpansion
}

// nodePTPCfgs implements NodePTPCfgInterface
type nodePTPCfgs struct {
	client rest.Interface
	ns     string
}

// newNodePTPCfgs returns a NodePTPCfgs
func newNodePTPCfgs(c *PtpV1Client, namespace string) *nodePTPCfgs {
	return &nodePTPCfgs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the nodePTPCfg, and returns the corresponding nodePTPCfg object, and an error if there is any.
func (c *nodePTPCfgs) Get(name string, options metav1.GetOptions) (result *v1.NodePTPCfg, err error) {
	result = &v1.NodePTPCfg{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NodePTPCfgs that match those selectors.
func (c *nodePTPCfgs) List(opts metav1.ListOptions) (result *v1.NodePTPCfgList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.NodePTPCfgList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nodePTPCfgs.
func (c *nodePTPCfgs) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a nodePTPCfg and creates it.  Returns the server's representation of the nodePTPCfg, and an error, if there is any.
func (c *nodePTPCfgs) Create(nodePTPCfg *v1.NodePTPCfg) (result *v1.NodePTPCfg, err error) {
	result = &v1.NodePTPCfg{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		Body(nodePTPCfg).
		Do().
		Into(result)
	return
}

// Update takes the representation of a nodePTPCfg and updates it. Returns the server's representation of the nodePTPCfg, and an error, if there is any.
func (c *nodePTPCfgs) Update(nodePTPCfg *v1.NodePTPCfg) (result *v1.NodePTPCfg, err error) {
	result = &v1.NodePTPCfg{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		Name(nodePTPCfg.Name).
		Body(nodePTPCfg).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *nodePTPCfgs) UpdateStatus(nodePTPCfg *v1.NodePTPCfg) (result *v1.NodePTPCfg, err error) {
	result = &v1.NodePTPCfg{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		Name(nodePTPCfg.Name).
		SubResource("status").
		Body(nodePTPCfg).
		Do().
		Into(result)
	return
}

// Delete takes name of the nodePTPCfg and deletes it. Returns an error if one occurs.
func (c *nodePTPCfgs) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nodePTPCfgs) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched nodePTPCfg.
func (c *nodePTPCfgs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.NodePTPCfg, err error) {
	result = &v1.NodePTPCfg{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("nodeptpcfgs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
