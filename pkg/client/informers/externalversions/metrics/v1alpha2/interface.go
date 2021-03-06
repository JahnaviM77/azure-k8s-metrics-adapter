// Code generated by informer-gen. DO NOT EDIT.

package v1alpha2

import (
	internalinterfaces "github.com/Azure/azure-k8s-metrics-adapter/pkg/client/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// CustomMetrics returns a CustomMetricInformer.
	CustomMetrics() CustomMetricInformer
	// ExternalMetrics returns a ExternalMetricInformer.
	ExternalMetrics() ExternalMetricInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// CustomMetrics returns a CustomMetricInformer.
func (v *version) CustomMetrics() CustomMetricInformer {
	return &customMetricInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ExternalMetrics returns a ExternalMetricInformer.
func (v *version) ExternalMetrics() ExternalMetricInformer {
	return &externalMetricInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
