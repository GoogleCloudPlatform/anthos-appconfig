// +build !ignore_autogenerated

/* Copyright 2019 Google LLC

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

// autogenerated by controller-gen object, do not modify manually

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateAllowedEgress) DeepCopyInto(out *AppEnvConfigTemplateAllowedEgress) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateAllowedEgress.
func (in *AppEnvConfigTemplateAllowedEgress) DeepCopy() *AppEnvConfigTemplateAllowedEgress {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateAllowedEgress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateAuth) DeepCopyInto(out *AppEnvConfigTemplateAuth) {
	*out = *in
	if in.JWT != nil {
		in, out := &in.JWT, &out.JWT
		*out = new(AppEnvConfigTemplateJWT)
		(*in).DeepCopyInto(*out)
	}
	if in.GCPAccess != nil {
		in, out := &in.GCPAccess, &out.GCPAccess
		*out = new(AppEnvConfigTemplateGCPAccess)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateAuth.
func (in *AppEnvConfigTemplateAuth) DeepCopy() *AppEnvConfigTemplateAuth {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateGCPAccess) DeepCopyInto(out *AppEnvConfigTemplateGCPAccess) {
	*out = *in
	if in.SecretInfo != nil {
		in, out := &in.SecretInfo, &out.SecretInfo
		*out = new(AppEnvConfigTemplateGCPAccessSecretInfo)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateGCPAccess.
func (in *AppEnvConfigTemplateGCPAccess) DeepCopy() *AppEnvConfigTemplateGCPAccess {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateGCPAccess)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateGCPAccessSecretInfo) DeepCopyInto(out *AppEnvConfigTemplateGCPAccessSecretInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateGCPAccessSecretInfo.
func (in *AppEnvConfigTemplateGCPAccessSecretInfo) DeepCopy() *AppEnvConfigTemplateGCPAccessSecretInfo {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateGCPAccessSecretInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateJWT) DeepCopyInto(out *AppEnvConfigTemplateJWT) {
	*out = *in
	if in.Params != nil {
		in, out := &in.Params, &out.Params
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateJWT.
func (in *AppEnvConfigTemplateJWT) DeepCopy() *AppEnvConfigTemplateJWT {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateJWT)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateRelatedClientInfo) DeepCopyInto(out *AppEnvConfigTemplateRelatedClientInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateRelatedClientInfo.
func (in *AppEnvConfigTemplateRelatedClientInfo) DeepCopy() *AppEnvConfigTemplateRelatedClientInfo {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateRelatedClientInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateServiceInfo) DeepCopyInto(out *AppEnvConfigTemplateServiceInfo) {
	*out = *in
	if in.AllowedClients != nil {
		in, out := &in.AllowedClients, &out.AllowedClients
		*out = make([]AppEnvConfigTemplateRelatedClientInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateServiceInfo.
func (in *AppEnvConfigTemplateServiceInfo) DeepCopy() *AppEnvConfigTemplateServiceInfo {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateServiceInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateV2) DeepCopyInto(out *AppEnvConfigTemplateV2) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateV2.
func (in *AppEnvConfigTemplateV2) DeepCopy() *AppEnvConfigTemplateV2 {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateV2)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppEnvConfigTemplateV2) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateV2List) DeepCopyInto(out *AppEnvConfigTemplateV2List) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AppEnvConfigTemplateV2, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateV2List.
func (in *AppEnvConfigTemplateV2List) DeepCopy() *AppEnvConfigTemplateV2List {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateV2List)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AppEnvConfigTemplateV2List) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateV2Spec) DeepCopyInto(out *AppEnvConfigTemplateV2Spec) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]AppEnvConfigTemplateServiceInfo, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AllowedEgress != nil {
		in, out := &in.AllowedEgress, &out.AllowedEgress
		*out = make([]AppEnvConfigTemplateAllowedEgress, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = new(AppEnvConfigTemplateAuth)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateV2Spec.
func (in *AppEnvConfigTemplateV2Spec) DeepCopy() *AppEnvConfigTemplateV2Spec {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateV2Spec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppEnvConfigTemplateV2Status) DeepCopyInto(out *AppEnvConfigTemplateV2Status) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppEnvConfigTemplateV2Status.
func (in *AppEnvConfigTemplateV2Status) DeepCopy() *AppEnvConfigTemplateV2Status {
	if in == nil {
		return nil
	}
	out := new(AppEnvConfigTemplateV2Status)
	in.DeepCopyInto(out)
	return out
}
