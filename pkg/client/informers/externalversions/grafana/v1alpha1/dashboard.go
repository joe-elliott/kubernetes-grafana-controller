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

package v1alpha1

import (
	grafanav1alpha1 "github.com/joe-elliott/kubernetes-grafana-controller/pkg/apis/grafana/v1alpha1"
	versioned "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/clientset/versioned"
	internalinterfaces "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/joe-elliott/kubernetes-grafana-controller/pkg/client/listers/grafana/v1alpha1"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// DashboardInformer provides access to a shared informer and lister for
// Dashboards.
type DashboardInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.DashboardLister
}

type dashboardInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewDashboardInformer constructs a new informer for Dashboard type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewDashboardInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredDashboardInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredDashboardInformer constructs a new informer for Dashboard type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredDashboardInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.GrafanaV1alpha1().Dashboards(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.GrafanaV1alpha1().Dashboards(namespace).Watch(options)
			},
		},
		&grafanav1alpha1.Dashboard{},
		resyncPeriod,
		indexers,
	)
}

func (f *dashboardInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredDashboardInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *dashboardInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&grafanav1alpha1.Dashboard{}, f.defaultInformer)
}

func (f *dashboardInformer) Lister() v1alpha1.DashboardLister {
	return v1alpha1.NewDashboardLister(f.Informer().GetIndexer())
}
