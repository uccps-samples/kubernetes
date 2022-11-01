package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/kubernetes/uccp-kube-apiserver/admission/route/apis/ingressadmission"
)

func (obj *IngressAdmissionConfig) GetObjectKind() schema.ObjectKind { return &obj.TypeMeta }

var GroupVersion = schema.GroupVersion{Group: "route.uccp.io", Version: "v1"}

var (
	localSchemeBuilder = runtime.NewSchemeBuilder(
		addKnownTypes,
		ingressadmission.Install,
	)
	Install = localSchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&IngressAdmissionConfig{},
	)
	return nil
}
