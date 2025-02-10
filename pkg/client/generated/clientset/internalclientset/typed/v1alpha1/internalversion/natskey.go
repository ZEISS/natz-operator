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

// NatsKeysGetter has a method to return a NatsKeyInterface.
// A group's client should implement this interface.
type NatsKeysGetter interface {
	NatsKeys() NatsKeyInterface
}

// NatsKeyInterface has methods to work with NatsKey resources.
type NatsKeyInterface interface {
	Create(ctx context.Context, natsKey *v1alpha1.NatsKey, opts v1.CreateOptions) (*v1alpha1.NatsKey, error)
	Update(ctx context.Context, natsKey *v1alpha1.NatsKey, opts v1.UpdateOptions) (*v1alpha1.NatsKey, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, natsKey *v1alpha1.NatsKey, opts v1.UpdateOptions) (*v1alpha1.NatsKey, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.NatsKey, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.NatsKeyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.NatsKey, err error)
	NatsKeyExpansion
}

// natsKeys implements NatsKeyInterface
type natsKeys struct {
	*gentype.ClientWithList[*v1alpha1.NatsKey, *v1alpha1.NatsKeyList]
}

// newNatsKeys returns a NatsKeys
func newNatsKeys(c *NatzClient) *natsKeys {
	return &natsKeys{
		gentype.NewClientWithList[*v1alpha1.NatsKey, *v1alpha1.NatsKeyList](
			"natskeys",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *v1alpha1.NatsKey { return &v1alpha1.NatsKey{} },
			func() *v1alpha1.NatsKeyList { return &v1alpha1.NatsKeyList{} },
		),
	}
}
