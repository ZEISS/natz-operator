/*
Copyright 2024.

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

package internalversion

import (
	context "context"

	v1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	scheme "github.com/zeiss/natz-operator/pkg/client/generated/clientset/internalclientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// NatsActivationsGetter has a method to return a NatsActivationInterface.
// A group's client should implement this interface.
type NatsActivationsGetter interface {
	NatsActivations(namespace string) NatsActivationInterface
}

// NatsActivationInterface has methods to work with NatsActivation resources.
type NatsActivationInterface interface {
	Create(ctx context.Context, natsActivation *v1alpha1.NatsActivation, opts v1.CreateOptions) (*v1alpha1.NatsActivation, error)
	Update(ctx context.Context, natsActivation *v1alpha1.NatsActivation, opts v1.UpdateOptions) (*v1alpha1.NatsActivation, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, natsActivation *v1alpha1.NatsActivation, opts v1.UpdateOptions) (*v1alpha1.NatsActivation, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.NatsActivation, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.NatsActivationList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.NatsActivation, err error)
	NatsActivationExpansion
}

// natsActivations implements NatsActivationInterface
type natsActivations struct {
	*gentype.ClientWithList[*v1alpha1.NatsActivation, *v1alpha1.NatsActivationList]
}

// newNatsActivations returns a NatsActivations
func newNatsActivations(c *NatzClient, namespace string) *natsActivations {
	return &natsActivations{
		gentype.NewClientWithList[*v1alpha1.NatsActivation, *v1alpha1.NatsActivationList](
			"natsactivations",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1alpha1.NatsActivation { return &v1alpha1.NatsActivation{} },
			func() *v1alpha1.NatsActivationList { return &v1alpha1.NatsActivationList{} },
		),
	}
}
