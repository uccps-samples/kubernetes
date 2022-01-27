/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dualstack

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	kubeapiservertesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/test/integration/framework"
)

func Test_ServiceDualStackIPFamilyPolicy(t *testing.T) {
	// Create an IPv4IPv6 dual stack control-plane
	serviceCIDR := "10.0.0.0/16"
	secondaryServiceCIDR := "2001:db8:1::/112"
	defer featuregatetesting.SetFeatureGateDuringTest(t, utilfeature.DefaultFeatureGate, features.IPv6DualStack, true)()

	s := kubeapiservertesting.StartTestServerOrDie(t, nil, []string{"--service-cluster-ip-range", serviceCIDR + "," + secondaryServiceCIDR}, framework.SharedEtcd())
	defer s.TearDownFn()

	b := &bytes.Buffer{}
	warningWriter := restclient.NewWarningWriter(b, restclient.WarningWriterOptions{})
	s.ClientConfig.WarningHandler = warningWriter
	client := clientset.NewForConfigOrDie(s.ClientConfig)

	singleStack := v1.IPFamilyPolicySingleStack
	preferDualStack := v1.IPFamilyPolicyPreferDualStack
	requireDualStack := v1.IPFamilyPolicyRequireDualStack
	var testcases = []struct {
		name           string
		clusterIP      string
		clusterIPs     []string
		ipFamilies     []v1.IPFamily
		ipFamilyPolicy *v1.IPFamilyPolicyType
		warnings       int
	}{
		// tests with hardcoded addresses first
		{
			name:           "Dual Stack - multiple cluster IPs - IPFamilyPolicy PreferDualStack - no warning",
			clusterIP:      "2001:db8:1::11",
			clusterIPs:     []string{"2001:db8:1::11", "10.0.0.11"},
			ipFamilyPolicy: &preferDualStack,
		},
		{
			name:       "Dual Stack - multiple clusterIPs - IPFamilyPolicy nil - warning",
			clusterIP:  "2001:db8:1::10",
			clusterIPs: []string{"2001:db8:1::10", "10.0.0.10"},
			warnings:   1,
		},
		// dynamic allocation
		{
			name:           "Single Stack - IPFamilyPolicy set - no warning",
			ipFamilies:     []v1.IPFamily{v1.IPv4Protocol},
			ipFamilyPolicy: &singleStack,
		},
		{
			name:           "Single Stack - IPFamilyPolicy nil - no warning",
			ipFamilies:     []v1.IPFamily{v1.IPv4Protocol},
			ipFamilyPolicy: nil,
		},
		{
			name:           "Dual Stack - multiple ipFamilies - IPFamilyPolicy RequireDualStack - no warning",
			ipFamilies:     []v1.IPFamily{v1.IPv4Protocol, v1.IPv6Protocol},
			ipFamilyPolicy: &requireDualStack,
		},
		{
			name:           "Dual Stack - multiple ipFamilies - IPFamilyPolicy nil - warning",
			clusterIPs:     []string{},
			ipFamilies:     []v1.IPFamily{v1.IPv4Protocol, v1.IPv6Protocol},
			ipFamilyPolicy: nil,
			warnings:       1,
		},
	}

	expectedWarningCount := 0
	for i, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("svc-test-%d", i), // use different services for each test
				},
				Spec: v1.ServiceSpec{
					Type:           v1.ServiceTypeClusterIP,
					ClusterIP:      tc.clusterIP,
					ClusterIPs:     tc.clusterIPs,
					IPFamilies:     tc.ipFamilies,
					IPFamilyPolicy: tc.ipFamilyPolicy,
					Ports: []v1.ServicePort{
						{
							Port:       443,
							TargetPort: intstr.FromInt(443),
						},
					},
				},
			}

			// create a service
			_, err := client.CoreV1().Services(metav1.NamespaceDefault).Create(context.TODO(), svc, metav1.CreateOptions{})
			if err != nil {
				t.Fatal(err)
			}

			if tc.warnings > 0 {
				expectedWarningCount += tc.warnings
				expectedPatchWarning := field.NewPath("service", "spec", "ipFamilyPolicy").String() + " must be RequireDualStack or PreferDualStack when multiple 'ipFamilies' are specified, this operation will fail starting with Red Hat OpenShift Platform 4.10."
				assertWarningMessage(t, b, expectedPatchWarning)
			}
			assertWarningCount(t, warningWriter, expectedWarningCount)
		})
	}
}

type warningCounter interface {
	WarningCount() int
}

func assertWarningCount(t *testing.T, counter warningCounter, expected int) {
	if counter.WarningCount() != expected {
		t.Errorf("unexpected warning count, expected: %v, got: %v", expected, counter.WarningCount())
	}
}

func assertWarningMessage(t *testing.T, b *bytes.Buffer, expected string) {
	defer b.Reset()
	actual := b.String()
	if len(expected) == 0 && len(actual) != 0 {
		t.Errorf("unexpected warning message, expected no warning, got: %v", actual)
	}
	if len(expected) == 0 {
		return
	}
	if !strings.Contains(actual, expected) {
		t.Errorf("unexpected warning message, expected: %v, got: %v", expected, actual)
	}
}
