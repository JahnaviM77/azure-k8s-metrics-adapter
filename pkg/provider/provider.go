package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/jsturtevant/azure-k8-metrics-adapter/pkg/aim"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
	"k8s.io/metrics/pkg/apis/external_metrics"

	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/servicebus/mgmt/2017-04-01/servicebus"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type externalMetric struct {
	info  provider.ExternalMetricInfo
	value external_metrics.ExternalMetricValue
}

type AzureProvider struct {
	client      dynamic.Interface
	mapper      apimeta.RESTMapper
	azureConfig *aim.AzureConfig

	values          map[provider.CustomMetricInfo]int64
	externalMetrics []externalMetric
}

func NewAzureProvider(client dynamic.Interface, mapper apimeta.RESTMapper) provider.MetricsProvider {
	azureConfig, err := aim.GetAzureConfig()
	if err != nil {
		glog.Errorf("unable to get azure config: %v", err)
	}

	return &AzureProvider{
		client:      client,
		mapper:      mapper,
		azureConfig: azureConfig,
		values:      make(map[provider.CustomMetricInfo]int64),
	}
}

/* Custom metric interface methods */
// not implemented
func (p *AzureProvider) GetRootScopedMetricByName(groupResource schema.GroupResource, name string, metricName string) (*custom_metrics.MetricValue, error) {
	//not implemented yet
	return nil, nil
}

// not implemented
func (p *AzureProvider) GetRootScopedMetricBySelector(groupResource schema.GroupResource, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	// not implemented yet
	return nil, nil
}

// not implemented
func (p *AzureProvider) GetNamespacedMetricByName(groupResource schema.GroupResource, namespace string, name string, metricName string) (*custom_metrics.MetricValue, error) {
	// not implemented yet
	return nil, nil
}

// not implemented
func (p *AzureProvider) GetNamespacedMetricBySelector(groupResource schema.GroupResource, namespace string, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	// not implemented yet
	return nil, nil
}

func (p *AzureProvider) ListAllMetrics() []provider.CustomMetricInfo {
	// not implemented yet
	return []provider.CustomMetricInfo{}
}

func (p *AzureProvider) GetExternalMetric(namespace string, metricName string, metricSelector labels.Selector) (*external_metrics.ExternalMetricValueList, error) {
	matchingMetrics := []external_metrics.ExternalMetricValue{}

	metricsClient := insights.NewMetricsClient(p.azureConfig.SubscriptionID)

	// create an authorizer from env vars or Azure Managed Service Idenity
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		metricsClient.Authorizer = authorizer
	}

	metricName = "Messages"
	metricResourceUri := metricResourceUri(p.azureConfig.SubscriptionID, "k8metrics", "k8custom")

	endtime := time.Now().UTC().Format(time.RFC3339)
	starttime := time.Now().Add(-(5 * time.Minute)).UTC().Format(time.RFC3339)
	timespan := fmt.Sprintf("%s/%s", starttime, endtime)

	metricResult, err := metricsClient.List(context.Background(), metricResourceUri, timespan, nil, metricName, "Total", nil, "", "", "", "")
	if err != nil {
		return nil, err
	}

	metricVals := *metricResult.Value
	Timeseries := *metricVals[0].Timeseries
	data := *Timeseries[0].Data
	total := *data[len(data)-1].Total

	metricValue := external_metrics.ExternalMetricValue{
		MetricName: metricName,
		Value:      *resource.NewQuantity(int64(total), resource.DecimalSI),
		Timestamp:  metav1.Now(),
	}
	matchingMetrics = append(matchingMetrics, metricValue)

	return &external_metrics.ExternalMetricValueList{
		Items: matchingMetrics,
	}, nil
}

func metricResourceUri(subId string, resourceGroup string, sbNameSpace string) string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ServiceBus/namespaces/%s", subId, resourceGroup, sbNameSpace)
}

func (p *AzureProvider) ListAllExternalMetrics() []provider.ExternalMetricInfo {
	externalMetricsInfo := []provider.ExternalMetricInfo{}

	namespaceClient := servicebus.NewNamespacesClient(p.azureConfig.SubscriptionID)

	// create an authorizer from env vars or Azure Managed Service Idenity
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err == nil {
		namespaceClient.Authorizer = authorizer
	}

	// TODO iterate over result set
	result, err := namespaceClient.List(context.Background())
	if err != nil {
		glog.Errorf("unable to get service bus namespaces: %v", err)
		return externalMetricsInfo
	}

	for _, namespace := range result.Values() {
		glog.V(2).Infoln("found namespace", *namespace.Name)
	}

	for _, metric := range p.externalMetrics {
		externalMetricsInfo = append(externalMetricsInfo, metric.info)
	}
	return externalMetricsInfo
}
