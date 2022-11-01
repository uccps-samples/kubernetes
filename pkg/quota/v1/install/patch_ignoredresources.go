package install

import "k8s.io/apimachinery/pkg/runtime/schema"

func init() {
	newIgnoredResources := map[schema.GroupResource]struct{}{
		{Group: "extensions", Resource: "networkpolicies"}:                            {},
		{Group: "", Resource: "bindings"}:                                             {},
		{Group: "", Resource: "componentstatuses"}:                                    {},
		{Group: "", Resource: "events"}:                                               {},
		{Group: "authentication.k8s.io", Resource: "tokenreviews"}:                    {},
		{Group: "authorization.k8s.io", Resource: "subjectaccessreviews"}:             {},
		{Group: "authorization.k8s.io", Resource: "selfsubjectaccessreviews"}:         {},
		{Group: "authorization.k8s.io", Resource: "localsubjectaccessreviews"}:        {},
		{Group: "authorization.k8s.io", Resource: "selfsubjectrulesreviews"}:          {},
		{Group: "authorization.uccp.io", Resource: "selfsubjectaccessreviews"}:   {},
		{Group: "authorization.uccp.io", Resource: "subjectaccessreviews"}:       {},
		{Group: "authorization.uccp.io", Resource: "localsubjectaccessreviews"}:  {},
		{Group: "authorization.uccp.io", Resource: "resourceaccessreviews"}:      {},
		{Group: "authorization.uccp.io", Resource: "localresourceaccessreviews"}: {},
		{Group: "authorization.uccp.io", Resource: "selfsubjectrulesreviews"}:    {},
		{Group: "authorization.uccp.io", Resource: "subjectrulesreviews"}:        {},
		{Group: "authorization.uccp.io", Resource: "roles"}:                      {},
		{Group: "authorization.uccp.io", Resource: "rolebindings"}:               {},
		{Group: "authorization.uccp.io", Resource: "clusterroles"}:               {},
		{Group: "authorization.uccp.io", Resource: "clusterrolebindings"}:        {},
		{Group: "apiregistration.k8s.io", Resource: "apiservices"}:                    {},
		{Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"}:        {},
	}
	for k, v := range newIgnoredResources {
		ignoredResources[k] = v
	}
}
