package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/uccp-kube-apiserver/admission/network/apis/restrictedendpoints"
)

var GroupVersion = schema.GroupVersion{Group: "network.uccp.io", Version: "v1"}

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		addKnownTypes,
		restrictedendpoints.Install,
	)
	Install = localSchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&RestrictedEndpointsAdmissionConfig{},
	)
	return nil
}
