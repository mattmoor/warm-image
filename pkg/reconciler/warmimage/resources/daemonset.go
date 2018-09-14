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
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	caching "github.com/knative/caching/pkg/apis/caching/v1alpha1"
	"github.com/knative/pkg/kmeta"
)

var (
	sleeperVolume = corev1.Volume{
		Name: "the-sleeper",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
	sleeperResources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1m"),
			corev1.ResourceMemory: resource.MustParse("20M"),
		},
	}
)

func sleeperContainer(sleeperImage string) corev1.Container {
	return corev1.Container{
		Name:  "the-sleeper",
		Image: sleeperImage,
		Args: []string{
			"-mode", "copy",
			"-to", "/drop/sleeper",
		},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      sleeperVolume.Name,
			MountPath: "/drop/",
		}},
		Resources: sleeperResources,
	}
}

func userContainer(image string) corev1.Container {
	return corev1.Container{
		Name:            "the-image",
		Image:           image,
		ImagePullPolicy: corev1.PullAlways,
		Command:         []string{"/drop/sleeper"},
		Args:            []string{"-mode", "sleep"},
		VolumeMounts: []corev1.VolumeMount{{
			Name:      sleeperVolume.Name,
			MountPath: "/drop/",
		}},
		Resources: sleeperResources,
	}
}

func MakeDaemonSet(wi *caching.Image, sleeperImage string) *extv1beta1.DaemonSet {
	ips := []corev1.LocalObjectReference{}
	if wi.Spec.ImagePullSecrets != nil {
		ips = append(ips, *wi.Spec.ImagePullSecrets)
	}
	return &extv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    wi.Name,
			Labels:          kmeta.MakeVersionLabels(wi),
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(wi)},
		},
		Spec: extv1beta1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: kmeta.MakeVersionLabels(wi),
				},
				Spec: corev1.PodSpec{
					InitContainers:   []corev1.Container{sleeperContainer(sleeperImage)},
					Containers:       []corev1.Container{userContainer(wi.Spec.Image)},
					ImagePullSecrets: ips,
					Volumes:          []corev1.Volume{sleeperVolume},
				},
			},
		},
	}
}
