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

package resources

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	caching "github.com/knative/caching/pkg/apis/caching/v1alpha1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

func TestRevisions(t *testing.T) {
	sleeperImage := "gcr.io/sleeper/image:latest"
	boolTrue := true

	tests := []struct {
		name  string
		image *caching.Image
		want  *extv1beta1.DaemonSet
	}{{
		name: "just image",
		image: &caching.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "foo",
				Namespace:       "bar",
				ResourceVersion: "abcd",
				Generation:      1234,
			},
			Spec: caching.ImageSpec{
				Image: "busybox",
			},
		},
		want: &extv1beta1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "foo",
				Labels: map[string]string{
					"controller": "",
					"version":    "abcd",
				},
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion:         "caching.internal.knative.dev/v1alpha1",
					Kind:               "Image",
					Name:               "foo",
					Controller:         &boolTrue,
					BlockOwnerDeletion: &boolTrue,
				}},
			},
			Spec: extv1beta1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"controller": "",
							"version":    "abcd",
						},
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{sleeperContainer(sleeperImage)},
						Containers:     []corev1.Container{userContainer("busybox")},
						Volumes:        []corev1.Volume{sleeperVolume},
					},
				},
			},
		},
	}, {
		name: "image and service account",
		image: &caching.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "foo",
				Namespace:       "bar",
				ResourceVersion: "bleh",
				Generation:      1234,
			},
			Spec: caching.ImageSpec{
				ServiceAccountName: "james-bond",
				Image:              "redis",
			},
		},
		want: &extv1beta1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "foo",
				Labels: map[string]string{
					"controller": "",
					"version":    "bleh",
				},
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion:         "caching.internal.knative.dev/v1alpha1",
					Kind:               "Image",
					Name:               "foo",
					Controller:         &boolTrue,
					BlockOwnerDeletion: &boolTrue,
				}},
			},
			Spec: extv1beta1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"controller": "",
							"version":    "bleh",
						},
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: "james-bond",
						InitContainers:     []corev1.Container{sleeperContainer(sleeperImage)},
						Containers:         []corev1.Container{userContainer("redis")},
						Volumes:            []corev1.Volume{sleeperVolume},
					},
				},
			},
		},
	}, {
		name: "image with pull secrets",
		image: &caching.Image{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "foo",
				Namespace:       "bar",
				ResourceVersion: "yuck",
				Generation:      1234,
			},
			Spec: caching.ImageSpec{
				Image: "mysql",
				ImagePullSecrets: []corev1.LocalObjectReference{{
					Name: "im-batman",
				}},
			},
		},
		want: &extv1beta1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "foo",
				Labels: map[string]string{
					"controller": "",
					"version":    "yuck",
				},
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion:         "caching.internal.knative.dev/v1alpha1",
					Kind:               "Image",
					Name:               "foo",
					Controller:         &boolTrue,
					BlockOwnerDeletion: &boolTrue,
				}},
			},
			Spec: extv1beta1.DaemonSetSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"controller": "",
							"version":    "yuck",
						},
					},
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{sleeperContainer(sleeperImage)},
						Containers:     []corev1.Container{userContainer("mysql")},
						ImagePullSecrets: []corev1.LocalObjectReference{{
							Name: "im-batman",
						}},
						Volumes: []corev1.Volume{sleeperVolume},
					},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := MakeDaemonSet(test.image, sleeperImage)
			if diff := cmp.Diff(test.want, got, cmpopts.IgnoreUnexported(resource.Quantity{})); diff != "" {
				t.Errorf("MakeDaemonSet (-want, +got) = %v", diff)
			}
		})
	}
}
