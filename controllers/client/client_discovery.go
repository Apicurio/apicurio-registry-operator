package client

import (
	"errors"
	"go.uber.org/zap"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type APIGroupInfo struct {
	Group            string
	PreferredVersion string
	Versions         []string
}

type DiscoveryClient struct {
	log    *zap.Logger
	client *discovery.DiscoveryClient
}

func NewDiscoveryClient(log *zap.Logger, clientConfig *rest.Config) *DiscoveryClient {
	this := &DiscoveryClient{
		log:    log,
		client: discovery.NewDiscoveryClientForConfigOrDie(clientConfig),
	}
	return this
}

func (this *DiscoveryClient) IsOCP() (bool, error) {
	_, err := this.client.ServerResourcesForGroupVersion("route.openshift.io/v1")
	if err != nil {
		if api_errors.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (this *DiscoveryClient) IsMonitoringInstalled() (bool, error) {
	return this.resourceExists("monitoring.coreos.com/v1", "ServiceMonitor")
}

// Get information about the given API group.
// Returns an error if the API Group does not exist or the info could not be determined.
func (this *DiscoveryClient) GetVersionInfoForAPIGroup(apiGroup string) (*APIGroupInfo, error) {
	apiGroupList, err := this.client.ServerGroups()
	if err != nil {
		return nil, err
	}
	res := &APIGroupInfo{
		Group:            apiGroup,
		PreferredVersion: "",
		Versions:         make([]string, 0),
	}
	for _, g := range apiGroupList.Groups {
		if g.Name == apiGroup {
			for _, v := range g.Versions {
				res.Versions = append(res.Versions, v.Version)
			}
			res.PreferredVersion = g.PreferredVersion.Version
			return res, nil
		}
	}
	return nil, errors.New("API group '" + apiGroup + "' not found")
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
