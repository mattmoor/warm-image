/*
Copyright 2017 The Kubernetes Authors.

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

package v2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WarmImage is a specification for a WarmImage resource
type WarmImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WarmImageSpec   `json:"spec"`
	Status WarmImageStatus `json:"status"`
}

// WarmImageSpec is the spec for a WarmImage resource
type WarmImageSpec struct {
	Image            string                       `json:"image"`
	ImagePullSecrets *corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// WarmImageStatus is the status for a WarmImage resource
type WarmImageStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WarmImageList is a list of WarmImage resources
type WarmImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []WarmImage `json:"items"`
}
