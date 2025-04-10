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

// fakeNatsAccounts implements NatsAccountInterface
type fakeNatsAccounts struct {
	*gentype.FakeClientWithList[*v1alpha1.NatsAccount, *v1alpha1.NatsAccountList]
	Fake *FakeNatz
}

func newFakeNatsAccounts(fake *FakeNatz) internalversion.NatsAccountInterface {
	return &fakeNatsAccounts{
		gentype.NewFakeClientWithList[*v1alpha1.NatsAccount, *v1alpha1.NatsAccountList](
			fake.Fake,
			"",
			v1alpha1.SchemeGroupVersion.WithResource("natsaccounts"),
			v1alpha1.SchemeGroupVersion.WithKind("NatsAccount"),
			func() *v1alpha1.NatsAccount { return &v1alpha1.NatsAccount{} },
			func() *v1alpha1.NatsAccountList { return &v1alpha1.NatsAccountList{} },
			func(dst, src *v1alpha1.NatsAccountList) { dst.ListMeta = src.ListMeta },
			func(list *v1alpha1.NatsAccountList) []*v1alpha1.NatsAccount {
				return gentype.ToPointerSlice(list.Items)
			},
			func(list *v1alpha1.NatsAccountList, items []*v1alpha1.NatsAccount) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
