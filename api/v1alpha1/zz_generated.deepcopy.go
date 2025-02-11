//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	v2 "github.com/nats-io/jwt/v2"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AuthCallout) DeepCopyInto(out *AuthCallout) {
	*out = *in
	if in.AuthUsers != nil {
		in, out := &in.AuthUsers, &out.AuthUsers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AuthCallout.
func (in *AuthCallout) DeepCopy() *AuthCallout {
	if in == nil {
		return nil
	}
	out := new(AuthCallout)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Authorization) DeepCopyInto(out *Authorization) {
	*out = *in
	in.AuthCallout.DeepCopyInto(&out.AuthCallout)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Authorization.
func (in *Authorization) DeepCopy() *Authorization {
	if in == nil {
		return nil
	}
	out := new(Authorization)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(Gateway)
		(*in).DeepCopyInto(*out)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLS)
		(*in).DeepCopyInto(*out)
	}
	if in.Authorization != nil {
		in, out := &in.Authorization, &out.Authorization
		*out = new(Authorization)
		(*in).DeepCopyInto(*out)
	}
	out.Resolver = in.Resolver
	if in.ResolverPreload != nil {
		in, out := &in.ResolverPreload, &out.ResolverPreload
		*out = make(ResolverPreload, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Export) DeepCopyInto(out *Export) {
	*out = *in
	if in.Revocations != nil {
		in, out := &in.Revocations, &out.Revocations
		*out = make(v2.RevocationList, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Latency != nil {
		in, out := &in.Latency, &out.Latency
		*out = new(v2.ServiceLatency)
		**out = **in
	}
	out.Info = in.Info
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Export.
func (in *Export) DeepCopy() *Export {
	if in == nil {
		return nil
	}
	out := new(Export)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gateway) DeepCopyInto(out *Gateway) {
	*out = *in
	in.Authorization.DeepCopyInto(&out.Authorization)
	if in.Gateways != nil {
		in, out := &in.Gateways, &out.Gateways
		*out = make([]GatewayEntry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gateway.
func (in *Gateway) DeepCopy() *Gateway {
	if in == nil {
		return nil
	}
	out := new(Gateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayEntry) DeepCopyInto(out *GatewayEntry) {
	*out = *in
	if in.URLS != nil {
		in, out := &in.URLS, &out.URLS
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.TLS.DeepCopyInto(&out.TLS)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayEntry.
func (in *GatewayEntry) DeepCopy() *GatewayEntry {
	if in == nil {
		return nil
	}
	out := new(GatewayEntry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JetStream) DeepCopyInto(out *JetStream) {
	*out = *in
	out.Limits = in.Limits
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JetStream.
func (in *JetStream) DeepCopy() *JetStream {
	if in == nil {
		return nil
	}
	out := new(JetStream)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JetStreamLimits) DeepCopyInto(out *JetStreamLimits) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JetStreamLimits.
func (in *JetStreamLimits) DeepCopy() *JetStreamLimits {
	if in == nil {
		return nil
	}
	out := new(JetStreamLimits)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Limits) DeepCopyInto(out *Limits) {
	*out = *in
	in.UserLimits.DeepCopyInto(&out.UserLimits)
	out.NatsLimits = in.NatsLimits
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Limits.
func (in *Limits) DeepCopy() *Limits {
	if in == nil {
		return nil
	}
	out := new(Limits)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsAccount) DeepCopyInto(out *NatsAccount) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsAccount.
func (in *NatsAccount) DeepCopy() *NatsAccount {
	if in == nil {
		return nil
	}
	out := new(NatsAccount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsAccount) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsAccountList) DeepCopyInto(out *NatsAccountList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsAccount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsAccountList.
func (in *NatsAccountList) DeepCopy() *NatsAccountList {
	if in == nil {
		return nil
	}
	out := new(NatsAccountList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsAccountList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsAccountReference) DeepCopyInto(out *NatsAccountReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsAccountReference.
func (in *NatsAccountReference) DeepCopy() *NatsAccountReference {
	if in == nil {
		return nil
	}
	out := new(NatsAccountReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsAccountSpec) DeepCopyInto(out *NatsAccountSpec) {
	*out = *in
	out.SignerKeyRef = in.SignerKeyRef
	out.PrivateKey = in.PrivateKey
	if in.SigningKeys != nil {
		in, out := &in.SigningKeys, &out.SigningKeys
		*out = make([]NatsKeyReference, len(*in))
		copy(*out, *in)
	}
	out.OperatorSigningKey = in.OperatorSigningKey
	if in.AllowUserNamespaces != nil {
		in, out := &in.AllowUserNamespaces, &out.AllowUserNamespaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Imports != nil {
		in, out := &in.Imports, &out.Imports
		*out = make([]*v2.Import, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(v2.Import)
				**out = **in
			}
		}
	}
	if in.Exports != nil {
		in, out := &in.Exports, &out.Exports
		*out = make([]Export, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Limits.DeepCopyInto(&out.Limits)
	if in.Revocations != nil {
		in, out := &in.Revocations, &out.Revocations
		*out = make(v2.RevocationList, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsAccountSpec.
func (in *NatsAccountSpec) DeepCopy() *NatsAccountSpec {
	if in == nil {
		return nil
	}
	out := new(NatsAccountSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsAccountStatus) DeepCopyInto(out *NatsAccountStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsAccountStatus.
func (in *NatsAccountStatus) DeepCopy() *NatsAccountStatus {
	if in == nil {
		return nil
	}
	out := new(NatsAccountStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsActivation) DeepCopyInto(out *NatsActivation) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsActivation.
func (in *NatsActivation) DeepCopy() *NatsActivation {
	if in == nil {
		return nil
	}
	out := new(NatsActivation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsActivation) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsActivationList) DeepCopyInto(out *NatsActivationList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsActivation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsActivationList.
func (in *NatsActivationList) DeepCopy() *NatsActivationList {
	if in == nil {
		return nil
	}
	out := new(NatsActivationList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsActivationList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsActivationReference) DeepCopyInto(out *NatsActivationReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsActivationReference.
func (in *NatsActivationReference) DeepCopy() *NatsActivationReference {
	if in == nil {
		return nil
	}
	out := new(NatsActivationReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsActivationSpec) DeepCopyInto(out *NatsActivationSpec) {
	*out = *in
	out.AccountRef = in.AccountRef
	out.SignerKeyRef = in.SignerKeyRef
	out.TargetAccountRef = in.TargetAccountRef
	in.Expiry.DeepCopyInto(&out.Expiry)
	in.Start.DeepCopyInto(&out.Start)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsActivationSpec.
func (in *NatsActivationSpec) DeepCopy() *NatsActivationSpec {
	if in == nil {
		return nil
	}
	out := new(NatsActivationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsActivationStatus) DeepCopyInto(out *NatsActivationStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsActivationStatus.
func (in *NatsActivationStatus) DeepCopy() *NatsActivationStatus {
	if in == nil {
		return nil
	}
	out := new(NatsActivationStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsConfig) DeepCopyInto(out *NatsConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsConfig.
func (in *NatsConfig) DeepCopy() *NatsConfig {
	if in == nil {
		return nil
	}
	out := new(NatsConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsConfigList) DeepCopyInto(out *NatsConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsConfigList.
func (in *NatsConfigList) DeepCopy() *NatsConfigList {
	if in == nil {
		return nil
	}
	out := new(NatsConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsConfigSpec) DeepCopyInto(out *NatsConfigSpec) {
	*out = *in
	out.OperatorRef = in.OperatorRef
	out.SystemAccountRef = in.SystemAccountRef
	if in.Gateways != nil {
		in, out := &in.Gateways, &out.Gateways
		*out = make([]NatsgatewayReference, len(*in))
		copy(*out, *in)
	}
	in.Config.DeepCopyInto(&out.Config)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsConfigSpec.
func (in *NatsConfigSpec) DeepCopy() *NatsConfigSpec {
	if in == nil {
		return nil
	}
	out := new(NatsConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsConfigStatus) DeepCopyInto(out *NatsConfigStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsConfigStatus.
func (in *NatsConfigStatus) DeepCopy() *NatsConfigStatus {
	if in == nil {
		return nil
	}
	out := new(NatsConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsGateway) DeepCopyInto(out *NatsGateway) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsGateway.
func (in *NatsGateway) DeepCopy() *NatsGateway {
	if in == nil {
		return nil
	}
	out := new(NatsGateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsGateway) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsGatewayList) DeepCopyInto(out *NatsGatewayList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsGateway, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsGatewayList.
func (in *NatsGatewayList) DeepCopy() *NatsGatewayList {
	if in == nil {
		return nil
	}
	out := new(NatsGatewayList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsGatewayList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsGatewaySpec) DeepCopyInto(out *NatsGatewaySpec) {
	*out = *in
	in.Username.DeepCopyInto(&out.Username)
	in.Password.DeepCopyInto(&out.Password)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsGatewaySpec.
func (in *NatsGatewaySpec) DeepCopy() *NatsGatewaySpec {
	if in == nil {
		return nil
	}
	out := new(NatsGatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsGatewayStatus) DeepCopyInto(out *NatsGatewayStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsGatewayStatus.
func (in *NatsGatewayStatus) DeepCopy() *NatsGatewayStatus {
	if in == nil {
		return nil
	}
	out := new(NatsGatewayStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsKey) DeepCopyInto(out *NatsKey) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsKey.
func (in *NatsKey) DeepCopy() *NatsKey {
	if in == nil {
		return nil
	}
	out := new(NatsKey)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsKey) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsKeyList) DeepCopyInto(out *NatsKeyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsKey, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsKeyList.
func (in *NatsKeyList) DeepCopy() *NatsKeyList {
	if in == nil {
		return nil
	}
	out := new(NatsKeyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsKeyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsKeyReference) DeepCopyInto(out *NatsKeyReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsKeyReference.
func (in *NatsKeyReference) DeepCopy() *NatsKeyReference {
	if in == nil {
		return nil
	}
	out := new(NatsKeyReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsKeySpec) DeepCopyInto(out *NatsKeySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsKeySpec.
func (in *NatsKeySpec) DeepCopy() *NatsKeySpec {
	if in == nil {
		return nil
	}
	out := new(NatsKeySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsKeyStatus) DeepCopyInto(out *NatsKeyStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsKeyStatus.
func (in *NatsKeyStatus) DeepCopy() *NatsKeyStatus {
	if in == nil {
		return nil
	}
	out := new(NatsKeyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsOperator) DeepCopyInto(out *NatsOperator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsOperator.
func (in *NatsOperator) DeepCopy() *NatsOperator {
	if in == nil {
		return nil
	}
	out := new(NatsOperator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsOperator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsOperatorList) DeepCopyInto(out *NatsOperatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsOperator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsOperatorList.
func (in *NatsOperatorList) DeepCopy() *NatsOperatorList {
	if in == nil {
		return nil
	}
	out := new(NatsOperatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsOperatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsOperatorReference) DeepCopyInto(out *NatsOperatorReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsOperatorReference.
func (in *NatsOperatorReference) DeepCopy() *NatsOperatorReference {
	if in == nil {
		return nil
	}
	out := new(NatsOperatorReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsOperatorSpec) DeepCopyInto(out *NatsOperatorSpec) {
	*out = *in
	out.PrivateKey = in.PrivateKey
	if in.SigningKeys != nil {
		in, out := &in.SigningKeys, &out.SigningKeys
		*out = make([]NatsKeyReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsOperatorSpec.
func (in *NatsOperatorSpec) DeepCopy() *NatsOperatorSpec {
	if in == nil {
		return nil
	}
	out := new(NatsOperatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsOperatorStatus) DeepCopyInto(out *NatsOperatorStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsOperatorStatus.
func (in *NatsOperatorStatus) DeepCopy() *NatsOperatorStatus {
	if in == nil {
		return nil
	}
	out := new(NatsOperatorStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsReference) DeepCopyInto(out *NatsReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsReference.
func (in *NatsReference) DeepCopy() *NatsReference {
	if in == nil {
		return nil
	}
	out := new(NatsReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsUser) DeepCopyInto(out *NatsUser) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsUser.
func (in *NatsUser) DeepCopy() *NatsUser {
	if in == nil {
		return nil
	}
	out := new(NatsUser)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsUser) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsUserList) DeepCopyInto(out *NatsUserList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NatsUser, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsUserList.
func (in *NatsUserList) DeepCopy() *NatsUserList {
	if in == nil {
		return nil
	}
	out := new(NatsUserList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NatsUserList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsUserSpec) DeepCopyInto(out *NatsUserSpec) {
	*out = *in
	out.PrivateKey = in.PrivateKey
	out.SignerKeyRef = in.SignerKeyRef
	out.AccountRef = in.AccountRef
	in.Permissions.DeepCopyInto(&out.Permissions)
	in.Limits.DeepCopyInto(&out.Limits)
	if in.AllowedConnectionTypes != nil {
		in, out := &in.AllowedConnectionTypes, &out.AllowedConnectionTypes
		*out = make(v2.StringList, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsUserSpec.
func (in *NatsUserSpec) DeepCopy() *NatsUserSpec {
	if in == nil {
		return nil
	}
	out := new(NatsUserSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsUserStatus) DeepCopyInto(out *NatsUserStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.LastUpdate.DeepCopyInto(&out.LastUpdate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsUserStatus.
func (in *NatsUserStatus) DeepCopy() *NatsUserStatus {
	if in == nil {
		return nil
	}
	out := new(NatsUserStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NatsgatewayReference) DeepCopyInto(out *NatsgatewayReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NatsgatewayReference.
func (in *NatsgatewayReference) DeepCopy() *NatsgatewayReference {
	if in == nil {
		return nil
	}
	out := new(NatsgatewayReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperatorLimits) DeepCopyInto(out *OperatorLimits) {
	*out = *in
	out.NatsLimits = in.NatsLimits
	out.AccountLimits = in.AccountLimits
	out.JetStreamLimits = in.JetStreamLimits
	if in.JetStreamTieredLimits != nil {
		in, out := &in.JetStreamTieredLimits, &out.JetStreamTieredLimits
		*out = make(v2.JetStreamTieredLimits, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperatorLimits.
func (in *OperatorLimits) DeepCopy() *OperatorLimits {
	if in == nil {
		return nil
	}
	out := new(OperatorLimits)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Permission) DeepCopyInto(out *Permission) {
	*out = *in
	if in.Allow != nil {
		in, out := &in.Allow, &out.Allow
		*out = make(v2.StringList, len(*in))
		copy(*out, *in)
	}
	if in.Deny != nil {
		in, out := &in.Deny, &out.Deny
		*out = make(v2.StringList, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Permission.
func (in *Permission) DeepCopy() *Permission {
	if in == nil {
		return nil
	}
	out := new(Permission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Permissions) DeepCopyInto(out *Permissions) {
	*out = *in
	in.Pub.DeepCopyInto(&out.Pub)
	in.Sub.DeepCopyInto(&out.Sub)
	if in.Resp != nil {
		in, out := &in.Resp, &out.Resp
		*out = new(v2.ResponsePermission)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Permissions.
func (in *Permissions) DeepCopy() *Permissions {
	if in == nil {
		return nil
	}
	out := new(Permissions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Resolver) DeepCopyInto(out *Resolver) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Resolver.
func (in *Resolver) DeepCopy() *Resolver {
	if in == nil {
		return nil
	}
	out := new(Resolver)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResolverPreload) DeepCopyInto(out *ResolverPreload) {
	{
		in := &in
		*out = make(ResolverPreload, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResolverPreload.
func (in ResolverPreload) DeepCopy() ResolverPreload {
	if in == nil {
		return nil
	}
	out := new(ResolverPreload)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretValueFromSource) DeepCopyInto(out *SecretValueFromSource) {
	*out = *in
	if in.SecretKeyRef != nil {
		in, out := &in.SecretKeyRef, &out.SecretKeyRef
		*out = new(v1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretValueFromSource.
func (in *SecretValueFromSource) DeepCopy() *SecretValueFromSource {
	if in == nil {
		return nil
	}
	out := new(SecretValueFromSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLS) DeepCopyInto(out *TLS) {
	*out = *in
	if in.PinnedCerts != nil {
		in, out := &in.PinnedCerts, &out.PinnedCerts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLS.
func (in *TLS) DeepCopy() *TLS {
	if in == nil {
		return nil
	}
	out := new(TLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *UserLimits) DeepCopyInto(out *UserLimits) {
	*out = *in
	if in.Src != nil {
		in, out := &in.Src, &out.Src
		*out = make(v2.CIDRList, len(*in))
		copy(*out, *in)
	}
	if in.Times != nil {
		in, out := &in.Times, &out.Times
		*out = make([]v2.TimeRange, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new UserLimits.
func (in *UserLimits) DeepCopy() *UserLimits {
	if in == nil {
		return nil
	}
	out := new(UserLimits)
	in.DeepCopyInto(out)
	return out
}
