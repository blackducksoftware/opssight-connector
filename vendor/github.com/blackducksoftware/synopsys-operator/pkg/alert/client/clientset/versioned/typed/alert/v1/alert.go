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

	scheme "github.com/blackducksoftware/synopsys-operator/pkg/alert/client/clientset/versioned/scheme"
	v1 "github.com/blackducksoftware/synopsys-operator/pkg/api/alert/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AlertsGetter has a method to return a AlertInterface.
// A group's client should implement this interface.
type AlertsGetter interface {
	Alerts(namespace string) AlertInterface
}

// AlertInterface has methods to work with Alert resources.
type AlertInterface interface {
	Create(*v1.Alert) (*v1.Alert, error)
	Update(*v1.Alert) (*v1.Alert, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Alert, error)
	List(opts metav1.ListOptions) (*v1.AlertList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Alert, err error)
	AlertExpansion
}

// alerts implements AlertInterface
type alerts struct {
	client rest.Interface
	ns     string
}

// newAlerts returns a Alerts
func newAlerts(c *SynopsysV1Client, namespace string) *alerts {
	return &alerts{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the alert, and returns the corresponding alert object, and an error if there is any.
func (c *alerts) Get(name string, options metav1.GetOptions) (result *v1.Alert, err error) {
	result = &v1.Alert{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("alerts").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Alerts that match those selectors.
func (c *alerts) List(opts metav1.ListOptions) (result *v1.AlertList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.AlertList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("alerts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested alerts.
func (c *alerts) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("alerts").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a alert and creates it.  Returns the server's representation of the alert, and an error, if there is any.
func (c *alerts) Create(alert *v1.Alert) (result *v1.Alert, err error) {
	result = &v1.Alert{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("alerts").
		Body(alert).
		Do().
		Into(result)
	return
}

// Update takes the representation of a alert and updates it. Returns the server's representation of the alert, and an error, if there is any.
func (c *alerts) Update(alert *v1.Alert) (result *v1.Alert, err error) {
	result = &v1.Alert{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("alerts").
		Name(alert.Name).
		Body(alert).
		Do().
		Into(result)
	return
}

// Delete takes name of the alert and deletes it. Returns an error if one occurs.
func (c *alerts) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("alerts").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *alerts) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("alerts").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched alert.
func (c *alerts) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Alert, err error) {
	result = &v1.Alert{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("alerts").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
