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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	time "time"

	ptpv1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/apis/ptp/v1"
	versioned "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/clientset/versioned"
	internalinterfaces "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/informers/externalversions/internalinterfaces"
	v1 "github.com/zshi-redhat/kube-ptp-daemon/pkg/client/listers/ptp/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NodePTPConfInformer provides access to a shared informer and lister for
// NodePTPConves.
type NodePTPConfInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.NodePTPConfLister
}

type nodePTPConfInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewNodePTPConfInformer constructs a new informer for NodePTPConf type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNodePTPConfInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNodePTPConfInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredNodePTPConfInformer constructs a new informer for NodePTPConf type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNodePTPConfInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PtpV1().NodePTPConves(namespace).List(options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.PtpV1().NodePTPConves(namespace).Watch(options)
			},
		},
		&ptpv1.NodePTPConf{},
		resyncPeriod,
		indexers,
	)
}

func (f *nodePTPConfInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNodePTPConfInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *nodePTPConfInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&ptpv1.NodePTPConf{}, f.defaultInformer)
}

func (f *nodePTPConfInformer) Lister() v1.NodePTPConfLister {
	return v1.NewNodePTPConfLister(f.Informer().GetIndexer())
}
