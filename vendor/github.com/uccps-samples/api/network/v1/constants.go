package v1

const (
	// Pod annotations
	AssignMacvlanAnnotation = "pod.network.uccp.io/assign-macvlan"

	// HostSubnet annotations. (Note: should be "hostsubnet.network.uccp.io/", but the incorrect name is now part of the API.)
	AssignHostSubnetAnnotation = "pod.network.uccp.io/assign-subnet"
	FixedVNIDHostAnnotation    = "pod.network.uccp.io/fixed-vnid-host"
	NodeUIDAnnotation          = "pod.network.uccp.io/node-uid"

	// NetNamespace annotations
	MulticastEnabledAnnotation = "netnamespace.network.uccp.io/multicast-enabled"

	// ChangePodNetworkAnnotation is an annotation on NetNamespace to request change of pod network
	ChangePodNetworkAnnotation string = "pod.network.uccp.io/multitenant.change-network"
)
