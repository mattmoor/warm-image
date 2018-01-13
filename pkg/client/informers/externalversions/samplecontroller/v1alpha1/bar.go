/*
Copyright 2018 The Kubernetes Authors.

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

// This file was automatically generated by informer-gen

package v1alpha1

import (
	time "time"

	samplecontroller_v1alpha1 "github.com/mattmoor/warm-image/pkg/apis/samplecontroller/v1alpha1"
	versioned "github.com/mattmoor/warm-image/pkg/client/clientset/versioned"
	internalinterfaces "github.com/mattmoor/warm-image/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/mattmoor/warm-image/pkg/client/listers/samplecontroller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// BarInformer provides access to a shared informer and lister for
// Bars.
type BarInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.BarLister
}

type barInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewBarInformer constructs a new informer for Bar type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBarInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBarInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredBarInformer constructs a new informer for Bar type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBarInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecontrollerV1alpha1().Bars(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplecontrollerV1alpha1().Bars(namespace).Watch(options)
			},
		},
		&samplecontroller_v1alpha1.Bar{},
		resyncPeriod,
		indexers,
	)
}

func (f *barInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBarInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *barInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&samplecontroller_v1alpha1.Bar{}, f.defaultInformer)
}

func (f *barInformer) Lister() v1alpha1.BarLister {
	return v1alpha1.NewBarLister(f.Informer().GetIndexer())
}
