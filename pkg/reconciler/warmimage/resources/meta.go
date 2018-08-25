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
	"fmt"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	warmimagev2 "github.com/mattmoor/warm-image/pkg/apis/warmimage/v2"
)

func MakeLabels(wi *warmimagev2.WarmImage) labels.Set {
	return map[string]string{
		"controller": string(wi.UID),
		"version":    wi.ResourceVersion,
	}
}

func MakeLabelSelector(wi *warmimagev2.WarmImage) labels.Selector {
	return labels.SelectorFromSet(MakeLabels(wi))
}

func MakeOldVersionLabelSelector(wi *warmimagev2.WarmImage) string {
	return fmt.Sprintf("controller=%s,version!=%s", wi.UID, wi.ResourceVersion)
}
