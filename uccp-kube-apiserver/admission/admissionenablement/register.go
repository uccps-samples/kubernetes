package admissionenablement

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/resourcequota"
	mutatingwebhook "k8s.io/apiserver/pkg/admission/plugin/webhook/mutating"

	"github.com/uccps-samples/apiserver-library-go/pkg/admission/imagepolicy"
	imagepolicyapiv1 "github.com/uccps-samples/apiserver-library-go/pkg/admission/imagepolicy/apis/imagepolicy/v1"
	quotaclusterresourcequota "github.com/uccps-samples/apiserver-library-go/pkg/admission/quota/clusterresourcequota"
	"github.com/uccps-samples/apiserver-library-go/pkg/securitycontextconstraints/sccadmission"
	authorizationrestrictusers "k8s.io/kubernetes/uccp-kube-apiserver/admission/authorization/restrictusers"
	quotaclusterresourceoverride "k8s.io/kubernetes/uccp-kube-apiserver/admission/autoscaling/clusterresourceoverride"
	"k8s.io/kubernetes/uccp-kube-apiserver/admission/autoscaling/managementcpusoverride"
	quotarunonceduration "k8s.io/kubernetes/uccp-kube-apiserver/admission/autoscaling/runonceduration"
	"k8s.io/kubernetes/uccp-kube-apiserver/admission/customresourcevalidation/customresourcevalidationregistration"
	"k8s.io/kubernetes/uccp-kube-apiserver/admission/network/externalipranger"
	"k8s.io/kubernetes/uccp-kube-apiserver/admission/network/restrictedendpoints"
	ingressadmission "k8s.io/kubernetes/uccp-kube-apiserver/admission/route"
	projectnodeenv "k8s.io/kubernetes/uccp-kube-apiserver/admission/scheduler/nodeenv"
	schedulerpodnodeconstraints "k8s.io/kubernetes/uccp-kube-apiserver/admission/scheduler/podnodeconstraints"
)

func RegisterOpenshiftKubeAdmissionPlugins(plugins *admission.Plugins) {
	authorizationrestrictusers.Register(plugins)
	imagepolicy.Register(plugins)
	ingressadmission.Register(plugins)
	managementcpusoverride.Register(plugins)
	projectnodeenv.Register(plugins)
	quotaclusterresourceoverride.Register(plugins)
	quotaclusterresourcequota.Register(plugins)
	quotarunonceduration.Register(plugins)
	schedulerpodnodeconstraints.Register(plugins)
	sccadmission.Register(plugins)
	sccadmission.RegisterSCCExecRestrictions(plugins)
	externalipranger.RegisterExternalIP(plugins)
	restrictedendpoints.RegisterRestrictedEndpoints(plugins)
}

var (

	// these are admission plugins that cannot be applied until after the kubeapiserver starts.
	// TODO if nothing comes to mind in 3.10, kill this
	SkipRunLevelZeroPlugins = sets.NewString()
	// these are admission plugins that cannot be applied until after the openshiftapiserver apiserver starts.
	SkipRunLevelOnePlugins = sets.NewString(
		imagepolicyapiv1.PluginName, // "image.uccp.io/ImagePolicy"
		"quota.uccp.io/ClusterResourceQuota",
		"security.uccp.io/SecurityContextConstraint",
		"security.uccp.io/SCCExecRestrictions",
	)

	// openshiftAdmissionPluginsForKubeBeforeMutating are the admission plugins to add after kube admission, before mutating webhooks
	openshiftAdmissionPluginsForKubeBeforeMutating = []string{
		"autoscaling.uccp.io/ClusterResourceOverride",
		managementcpusoverride.PluginName, // "autoscaling.uccp.io/ManagementCPUsOverride"
		"authorization.uccp.io/RestrictSubjectBindings",
		"autoscaling.uccp.io/RunOnceDuration",
		"scheduling.uccp.io/PodNodeConstraints",
		"scheduling.uccp.io/OriginPodNodeEnvironment",
		"network.uccp.io/ExternalIPRanger",
		"network.uccp.io/RestrictedEndpointsAdmission",
		imagepolicyapiv1.PluginName, // "image.uccp.io/ImagePolicy"
		"security.uccp.io/SecurityContextConstraint",
		"security.uccp.io/SCCExecRestrictions",
		"route.uccp.io/IngressAdmission",
	}

	// openshiftAdmissionPluginsForKubeAfterResourceQuota are the plugins to add after ResourceQuota plugin
	openshiftAdmissionPluginsForKubeAfterResourceQuota = []string{
		"quota.uccp.io/ClusterResourceQuota",
	}

	// additionalDefaultOnPlugins is a list of plugins we turn on by default that core kube does not.
	additionalDefaultOnPlugins = sets.NewString(
		"NodeRestriction",
		"OwnerReferencesPermissionEnforcement",
		"PersistentVolumeLabel",
		"PodNodeSelector",
		"PodTolerationRestriction",
		"Priority",
		imagepolicyapiv1.PluginName, // "image.uccp.io/ImagePolicy"
		"StorageObjectInUseProtection",
	)
)

func NewOrderedKubeAdmissionPlugins(kubeAdmissionOrder []string) []string {
	ret := []string{}
	for _, curr := range kubeAdmissionOrder {
		if curr == mutatingwebhook.PluginName {
			ret = append(ret, openshiftAdmissionPluginsForKubeBeforeMutating...)
			ret = append(ret, customresourcevalidationregistration.AllCustomResourceValidators...)
		}

		ret = append(ret, curr)

		if curr == resourcequota.PluginName {
			ret = append(ret, openshiftAdmissionPluginsForKubeAfterResourceQuota...)
		}
	}
	return ret
}

func NewDefaultOffPluginsFunc(kubeDefaultOffAdmission sets.String) func() sets.String {
	return func() sets.String {
		kubeOff := sets.NewString(kubeDefaultOffAdmission.UnsortedList()...)
		kubeOff.Delete(additionalDefaultOnPlugins.List()...)
		kubeOff.Delete(openshiftAdmissionPluginsForKubeBeforeMutating...)
		kubeOff.Delete(openshiftAdmissionPluginsForKubeAfterResourceQuota...)
		kubeOff.Delete(customresourcevalidationregistration.AllCustomResourceValidators...)
		return kubeOff
	}
}
