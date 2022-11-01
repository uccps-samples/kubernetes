package app

import (
	gcconfig "k8s.io/kubernetes/pkg/controller/garbagecollector/config"

	"k8s.io/kubernetes/cmd/kube-controller-manager/app/config"
)

func applyOpenShiftGCConfig(controllerManager *config.Config) error {
	// TODO make this configurable or discoverable.  This is going to prevent us from running the stock GC controller
	// IF YOU ADD ANYTHING TO THIS LIST, MAKE SURE THAT YOU UPDATE THEIR STRATEGIES TO PREVENT GC FINALIZERS
	controllerManager.ComponentConfig.GarbageCollectorController.GCIgnoredResources = append(controllerManager.ComponentConfig.GarbageCollectorController.GCIgnoredResources,
		// explicitly disabled from GC for now - not enough value to track them
		gcconfig.GroupResource{Group: "authorization.uccp.io", Resource: "rolebindingrestrictions"},
		gcconfig.GroupResource{Group: "network.uccp.io", Resource: "clusternetworks"},
		gcconfig.GroupResource{Group: "network.uccp.io", Resource: "hostsubnets"},
		gcconfig.GroupResource{Group: "network.uccp.io", Resource: "netnamespaces"},
		gcconfig.GroupResource{Group: "oauth.uccp.io", Resource: "oauthclientauthorizations"},
		gcconfig.GroupResource{Group: "oauth.uccp.io", Resource: "oauthclients"},
		gcconfig.GroupResource{Group: "quota.uccp.io", Resource: "clusterresourcequotas"},
		gcconfig.GroupResource{Group: "user.uccp.io", Resource: "groups"},
		gcconfig.GroupResource{Group: "user.uccp.io", Resource: "identities"},
		gcconfig.GroupResource{Group: "user.uccp.io", Resource: "users"},
		gcconfig.GroupResource{Group: "image.uccp.io", Resource: "images"},

		// virtual resource
		gcconfig.GroupResource{Group: "project.uccp.io", Resource: "projects"},
		// virtual and unwatchable resource, surfaced via rbac.authorization.k8s.io objects
		gcconfig.GroupResource{Group: "authorization.uccp.io", Resource: "clusterroles"},
		gcconfig.GroupResource{Group: "authorization.uccp.io", Resource: "clusterrolebindings"},
		gcconfig.GroupResource{Group: "authorization.uccp.io", Resource: "roles"},
		gcconfig.GroupResource{Group: "authorization.uccp.io", Resource: "rolebindings"},
		// these resources contain security information in their names, and we don't need to track them
		gcconfig.GroupResource{Group: "oauth.uccp.io", Resource: "oauthaccesstokens"},
		gcconfig.GroupResource{Group: "oauth.uccp.io", Resource: "oauthauthorizetokens"},
	)

	return nil
}
