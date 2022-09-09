package client

import (
	"github.com/go-logr/logr"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type DiscoveryClient struct {
	log    logr.Logger
	client *discovery.DiscoveryClient
}

func NewDiscoveryClient(log logr.Logger, clientConfig *rest.Config) *DiscoveryClient {
	this := &DiscoveryClient{
		log:    log,
		client: discovery.NewDiscoveryClientForConfigOrDie(clientConfig),
	}
	return this
}

var isOpenshift *bool
var isMonitoringInstalled *bool

func (this *DiscoveryClient) IsOCP() (bool, error) {
	if isOpenshift == nil {
		is, err := this.detectOpenshift()
		if err != nil {
			return is, err
		}
		isOpenshift = &is
	}
	return *isOpenshift, nil
}

func (this *DiscoveryClient) IsMonitoringInstalled() (bool, error) {
	if isMonitoringInstalled == nil {
		is, err := this.detectServiceMonitoring()
		if err != nil {
			return is, err
		}
		isMonitoringInstalled = &is
	}
	return *isMonitoringInstalled, nil
}

func (this *DiscoveryClient) detectServiceMonitoring() (bool, error) {

	_, err := this.client.ServerResourcesForGroupVersion("monitoring.coreos.com/v1")

	if err != nil && api_errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	serviceMonitorRegistered, err := this.resourceExists("monitoring.coreos.com/v1", "ServiceMonitor")

	if err != nil && api_errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return serviceMonitorRegistered, nil
}

func (this *DiscoveryClient) resourceExists(apiGroupVersion, kind string) (bool, error) {

	_, apiLists, err := this.client.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == apiGroupVersion {
			for _, r := range apiList.APIResources {
				if r.Kind == kind {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (this *DiscoveryClient) detectOpenshift() (bool, error) {
	_, err := this.client.ServerResourcesForGroupVersion("route.openshift.io/v1")

	if err != nil && api_errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
