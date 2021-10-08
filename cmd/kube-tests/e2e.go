package main

import (
	"strings"

	e2e "k8s.io/kubernetes/test/e2e/framework"

	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubernetes/cmd/kube-tests/testginkgo"
)

var TestContext = &e2e.TestContext

func isDisabled(name string) bool {
	return strings.Contains(name, "[Disabled")
}

type testSuite struct {
	testginkgo.TestSuite

	PreSuite  func(opt *runOptions) error
	PostSuite func(opt *runOptions)

	PreTest func() error
}

type testSuites []testSuite

func (s testSuites) TestSuites() []*testginkgo.TestSuite {
	copied := make([]*testginkgo.TestSuite, 0, len(s))
	for i := range s {
		copied = append(copied, &s[i].TestSuite)
	}
	return copied
}

// staticSuites are all known test suites this binary should run
var staticSuites = testSuites{
	{
		TestSuite: testginkgo.TestSuite{
			Name: "kubernetes/conformance",
			Description: templates.LongDesc(`
		The default Kubernetes conformance suite.
		`),
			Matches: func(name string) bool {
				if isDisabled(name) {
					return false
				}
				return strings.Contains(name, "[Suite:k8s]") && strings.Contains(name, "[Conformance]")
			},
			Parallelism: 30,
			//SyntheticEventTests: testginkgo.JUnitForEventsFunc(synthetictests.StableSystemEventInvariants),
		},
		PreSuite: suiteWithProviderPreSuite,
	},
	{
		TestSuite: testginkgo.TestSuite{
			Name: "all",
			Description: templates.LongDesc(`
		Run all tests.
		`),
			Matches: func(name string) bool {
				return true
			},
		},
		PreSuite: suiteWithInitializedProviderPreSuite,
	},
}

// suiteWithInitializedProviderPreSuite loads the provider info, but does not
// exclude any tests specific to that provider.
func suiteWithInitializedProviderPreSuite(opt *runOptions) error {
	config, err := decodeProvider(opt.Provider, opt.DryRun, true, nil)
	if err != nil {
		return err
	}
	opt.config = config

	opt.Provider = config.ToJSONString()
	return nil
}

// suiteWithProviderPreSuite ensures that the suite filters out tests from providers
// that aren't relevant (see exutilcluster.ClusterConfig.MatchFn) by loading the
// provider info from the cluster or flags.
func suiteWithProviderPreSuite(opt *runOptions) error {
	if err := suiteWithInitializedProviderPreSuite(opt); err != nil {
		return err
	}
	opt.MatchFn = opt.config.MatchFn()
	return nil
}

// suiteWithNoProviderPreSuite blocks out provider settings from being passed to
// child tests. Used with suites that should not have cloud specific behavior.
func suiteWithNoProviderPreSuite(opt *runOptions) error {
	opt.Provider = `none`
	return suiteWithProviderPreSuite(opt)
}

// suiteWithKubeTestInitialization invokes the Kube suite in order to populate
// data from the environment for the CSI suite. Other suites should use
// suiteWithProviderPreSuite.
func suiteWithKubeTestInitializationPreSuite(opt *runOptions) error {
	if err := suiteWithProviderPreSuite(opt); err != nil {
		return err
	}
	return initializeTestFramework(TestContext, opt.config, opt.DryRun)
}
