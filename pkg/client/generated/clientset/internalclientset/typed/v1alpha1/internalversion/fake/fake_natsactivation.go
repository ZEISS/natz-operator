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

package fake

import (
	v1alpha1 "github.com/zeiss/natz-operator/api/v1alpha1"
	internalversion "github.com/zeiss/natz-operator/pkg/client/generated/clientset/internalclientset/typed/v1alpha1/internalversion"
	gentype "k8s.io/client-go/gentype"
)

// fakeNatsActivations implements NatsActivationInterface
type fakeNatsActivations struct {
	*gentype.FakeClientWithList[*v1alpha1.NatsActivation, *v1alpha1.NatsActivationList]
	Fake *FakeNatz
}

func newFakeNatsActivations(fake *FakeNatz, namespace string) internalversion.NatsActivationInterface {
	return &fakeNatsActivations{
		gentype.NewFakeClientWithList[*v1alpha1.NatsActivation, *v1alpha1.NatsActivationList](
			fake.Fake,
			namespace,
			v1alpha1.SchemeGroupVersion.WithResource("natsactivations"),
			v1alpha1.SchemeGroupVersion.WithKind("NatsActivation"),
			func() *v1alpha1.NatsActivation { return &v1alpha1.NatsActivation{} },
			func() *v1alpha1.NatsActivationList { return &v1alpha1.NatsActivationList{} },
			func(dst, src *v1alpha1.NatsActivationList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.NatsActivationList) []*v1alpha1.NatsActivation {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.NatsActivationList, items []*v1alpha1.NatsActivation) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
