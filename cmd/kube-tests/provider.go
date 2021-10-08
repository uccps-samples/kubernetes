package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"k8s.io/kubernetes/cmd/kube-tests/cluster"
	testsutil "k8s.io/kubernetes/cmd/kube-tests/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	// these are loading important global flags that we need to get and set
	_ "k8s.io/kubernetes/test/e2e"
	_ "k8s.io/kubernetes/test/e2e/lifecycle"
)

func initializeTestFramework(context *e2e.TestContextType, config *cluster.ClusterConfiguration, dryRun bool) error {
	// update context with loaded config
	context.Provider = config.ProviderName
	context.CloudConfig = e2e.CloudConfig{
		ProjectID:   config.ProjectID,
		Region:      config.Region,
		Zone:        config.Zone,
		Zones:       config.Zones,
		NumNodes:    config.NumNodes,
		MultiMaster: config.MultiMaster,
		MultiZone:   config.MultiZone,
		ConfigFile:  config.ConfigFile,
	}
	context.AllowedNotReadyNodes = -1
	context.MinStartupPods = -1
	context.MaxNodesToGather = 0

	if err := testsutil.InitTest(dryRun); err != nil {
		return err
	}
	gomega.RegisterFailHandler(ginkgo.Fail)

	e2e.AfterReadingAllFlags(context)
	context.DumpLogsOnFailure = true

	// these constants are taken from kube e2e and used by tests
	context.IPFamily = "ipv4"
	if config.HasIPv6 && !config.HasIPv4 {
		context.IPFamily = "ipv6"
	}
	return nil
}

func decodeProvider(provider string, dryRun, discover bool, clusterState *cluster.ClusterState) (*cluster.ClusterConfiguration, error) {
	switch provider {
	case "none":
		return &cluster.ClusterConfiguration{ProviderName: "skeleton"}, nil

	case "":
		if _, ok := os.LookupEnv("KUBE_SSH_USER"); ok {
			if _, ok := os.LookupEnv("LOCAL_SSH_KEY"); ok {
				return &cluster.ClusterConfiguration{ProviderName: "local"}, nil
			}
		}
		if dryRun {
			return &cluster.ClusterConfiguration{ProviderName: "skeleton"}, nil
		}
		fallthrough

	case "azure", "aws", "baremetal", "gce", "vsphere":
		if clusterState == nil {
			clientConfig, err := e2e.LoadConfig(true)
			if err != nil {
				return nil, err
			}
			clusterState, err = cluster.DiscoverClusterState(clientConfig)
			if err != nil {
				return nil, err
			}
		}
		config, err := cluster.LoadConfig(clusterState)
		if err != nil {
			return nil, err
		}
		if len(config.ProviderName) == 0 {
			config.ProviderName = "skeleton"
		}
		return config, nil

	default:
		var providerInfo struct {
			Type string
		}
		if err := json.Unmarshal([]byte(provider), &providerInfo); err != nil {
			return nil, fmt.Errorf("provider must be a JSON object with the 'type' key at a minimum: %v", err)
		}
		if len(providerInfo.Type) == 0 {
			return nil, fmt.Errorf("provider must be a JSON object with the 'type' key")
		}
		var config *cluster.ClusterConfiguration
		if discover {
			if clusterState == nil {
				if clientConfig, err := e2e.LoadConfig(true); err == nil {
					clusterState, _ = cluster.DiscoverClusterState(clientConfig)
				}
			}
			if clusterState != nil {
				config, _ = cluster.LoadConfig(clusterState)
			}
		}
		if config == nil {
			config = &cluster.ClusterConfiguration{}
		}

		if err := json.Unmarshal([]byte(provider), config); err != nil {
			return nil, fmt.Errorf("provider must decode into the ClusterConfig object: %v", err)
		}
		return config, nil
	}
}
