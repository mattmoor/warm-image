package controller

import (
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	clientset "github.com/mattmoor/warm-image/pkg/client/clientset/versioned"
	informers "github.com/mattmoor/warm-image/pkg/client/informers/externalversions"
)

type Interface interface {
	Run(threadiness int, stopCh <-chan struct{}) error
}

type Constructor func(kubernetes.Interface, clientset.Interface, kubeinformers.SharedInformerFactory, informers.SharedInformerFactory) Interface
