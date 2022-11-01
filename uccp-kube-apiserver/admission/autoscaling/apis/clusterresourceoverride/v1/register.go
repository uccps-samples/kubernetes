package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/kubernetes/uccp-kube-apiserver/admission/autoscaling/apis/clusterresourceoverride"
)

func (obj *ClusterResourceOverrideConfig) GetObjectKind() schema.ObjectKind { return &obj.TypeMeta }

var GroupVersion = schema.GroupVersion{Group: "autoscaling.uccp.io", Version: "v1"}

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		addKnownTypes,
		clusterresourceoverride.Install,
	)
	Install = localSchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&ClusterResourceOverrideConfig{},
	)
	return nil
}
