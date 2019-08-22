// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeMatchList) DeepCopyInto(out *NodeMatchList) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeMatchList.
func (in *NodeMatchList) DeepCopy() *NodeMatchList {
	if in == nil {
		return nil
	}
	out := new(NodeMatchList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPConf) DeepCopyInto(out *NodePTPConf) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPConf.
func (in *NodePTPConf) DeepCopy() *NodePTPConf {
	if in == nil {
		return nil
	}
	out := new(NodePTPConf)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodePTPConf) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPConfList) DeepCopyInto(out *NodePTPConfList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodePTPConf, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPConfList.
func (in *NodePTPConfList) DeepCopy() *NodePTPConfList {
	if in == nil {
		return nil
	}
	out := new(NodePTPConfList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodePTPConfList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPConfSpec) DeepCopyInto(out *NodePTPConfSpec) {
	*out = *in
	if in.Profile != nil {
		in, out := &in.Profile, &out.Profile
		*out = make([]NodePTPProfile, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Recommend != nil {
		in, out := &in.Recommend, &out.Recommend
		*out = make([]NodePTPRecommend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPConfSpec.
func (in *NodePTPConfSpec) DeepCopy() *NodePTPConfSpec {
	if in == nil {
		return nil
	}
	out := new(NodePTPConfSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPConfStatus) DeepCopyInto(out *NodePTPConfStatus) {
	*out = *in
	if in.MatchList != nil {
		in, out := &in.MatchList, &out.MatchList
		*out = make([]NodeMatchList, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPConfStatus.
func (in *NodePTPConfStatus) DeepCopy() *NodePTPConfStatus {
	if in == nil {
		return nil
	}
	out := new(NodePTPConfStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPDev) DeepCopyInto(out *NodePTPDev) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPDev.
func (in *NodePTPDev) DeepCopy() *NodePTPDev {
	if in == nil {
		return nil
	}
	out := new(NodePTPDev)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodePTPDev) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPDevList) DeepCopyInto(out *NodePTPDevList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]NodePTPDev, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPDevList.
func (in *NodePTPDevList) DeepCopy() *NodePTPDevList {
	if in == nil {
		return nil
	}
	out := new(NodePTPDevList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *NodePTPDevList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPDevSpec) DeepCopyInto(out *NodePTPDevSpec) {
	*out = *in
	if in.PTPDevices != nil {
		in, out := &in.PTPDevices, &out.PTPDevices
		*out = make([]PTPDevice, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPDevSpec.
func (in *NodePTPDevSpec) DeepCopy() *NodePTPDevSpec {
	if in == nil {
		return nil
	}
	out := new(NodePTPDevSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPDevStatus) DeepCopyInto(out *NodePTPDevStatus) {
	*out = *in
	if in.PTPDevices != nil {
		in, out := &in.PTPDevices, &out.PTPDevices
		*out = make([]PTPDevice, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPDevStatus.
func (in *NodePTPDevStatus) DeepCopy() *NodePTPDevStatus {
	if in == nil {
		return nil
	}
	out := new(NodePTPDevStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPMatchRule) DeepCopyInto(out *NodePTPMatchRule) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPMatchRule.
func (in *NodePTPMatchRule) DeepCopy() *NodePTPMatchRule {
	if in == nil {
		return nil
	}
	out := new(NodePTPMatchRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPProfile) DeepCopyInto(out *NodePTPProfile) {
	*out = *in
	if in.Interfaces != nil {
		in, out := &in.Interfaces, &out.Interfaces
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Ptp4lConf != nil {
		in, out := &in.Ptp4lConf, &out.Ptp4lConf
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPProfile.
func (in *NodePTPProfile) DeepCopy() *NodePTPProfile {
	if in == nil {
		return nil
	}
	out := new(NodePTPProfile)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePTPRecommend) DeepCopyInto(out *NodePTPRecommend) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = make([]NodePTPMatchRule, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePTPRecommend.
func (in *NodePTPRecommend) DeepCopy() *NodePTPRecommend {
	if in == nil {
		return nil
	}
	out := new(NodePTPRecommend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PTPDevice) DeepCopyInto(out *PTPDevice) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PTPDevice.
func (in *PTPDevice) DeepCopy() *PTPDevice {
	if in == nil {
		return nil
	}
	out := new(PTPDevice)
	in.DeepCopyInto(out)
	return out
}
