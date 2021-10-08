package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kube-tests/util/image"

	"k8s.io/kubernetes/cmd/kube-tests/cluster"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/version"
	"k8s.io/kubectl/pkg/util/templates"
	e2e "k8s.io/kubernetes/test/e2e/framework"
	k8simage "k8s.io/kubernetes/test/utils/image"

	"k8s.io/kubernetes/cmd/kube-tests/testginkgo"
)

type runOptions struct {
	testginkgo.Options

	FromRepository string
	Provider       string

	// Passed to the test process if set
	UpgradeSuite string
	ToImage      string
	TestOptions  []string

	// Shared by initialization code
	config *cluster.ClusterConfiguration
}

func (opt *runOptions) SelectSuite(suites testSuites, args []string) (*testSuite, error) {
	suite, err := opt.Options.SelectSuite(suites.TestSuites(), args)
	if err != nil {
		return nil, err
	}
	for i := range suites {
		if &suites[i].TestSuite == suite {
			return &suites[i], nil
		}
	}
	if len(opt.Provider) > 0 {
		return &testSuite{TestSuite: *suite, PreSuite: suiteWithProviderPreSuite}, nil
	}
	return &testSuite{TestSuite: *suite}, nil
}

func (opt *runOptions) AsEnv() []string {
	var args []string
	args = append(args, "KUBE_TEST_REPO_LIST=") // explicitly prevent selective override
	args = append(args, fmt.Sprintf("KUBE_TEST_REPO=%s", opt.FromRepository))
	args = append(args, fmt.Sprintf("TEST_PROVIDER=%s", opt.Provider))
	args = append(args, fmt.Sprintf("TEST_JUNIT_DIR=%s", opt.JUnitDir))
	for i := 10; i > 0; i-- {
		if klog.V(klog.Level(i)).Enabled() {
			args = append(args, fmt.Sprintf("TEST_LOG_LEVEL=%d", i))
			break
		}
	}

	if len(opt.UpgradeSuite) > 0 {
		data, err := json.Marshal(UpgradeOptions{
			Suite:       opt.UpgradeSuite,
			ToImage:     opt.ToImage,
			TestOptions: opt.TestOptions,
		})
		if err != nil {
			panic(err)
		}
		args = append(args, fmt.Sprintf("TEST_UPGRADE_OPTIONS=%s", string(data)))
	} else {
		args = append(args, "TEST_UPGRADE_OPTIONS=")
	}

	return args
}

type UpgradeOptions struct {
	Suite       string
	ToImage     string
	TestOptions []string
}

func (o *UpgradeOptions) ToEnv() string {
	out, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// defaultTestImageMirrorLocation is where all Kube test inputs are sourced.
const defaultTestImageMirrorLocation = "quay.io/openshift/community-e2e-images"

func main() {
	// KUBE_TEST_REPO_LIST is calculated during package initialization and prevents
	// proper mirroring of images referenced by tests. Clear the value and re-exec the
	// current process to ensure we can verify from a known state.
	if len(os.Getenv("KUBE_TEST_REPO_LIST")) > 0 {
		fmt.Fprintln(os.Stderr, "warning: KUBE_TEST_REPO_LIST may not be set when using openshift-tests and will be ignored")
		os.Setenv("KUBE_TEST_REPO_LIST", "")
		// resolve the call to execute since Exec() does not do PATH resolution
		if err := syscall.Exec(exec.Command(os.Args[0]).Path, os.Args, os.Environ()); err != nil {
			panic(fmt.Sprintf("%s: %v", os.Args[0], err))
		}
		return
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	rand.Seed(time.Now().UTC().UnixNano())

	pflag.CommandLine.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
	//pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	root := &cobra.Command{
		Long: templates.LongDesc(`
		Kubernetes Tests

		This command verifies behavior of an Kubernetes cluster by running remote tests against
		the cluster API that exercise functionality. In general these tests may be disruptive
		or require elevated privileges - see the descriptions of each test suite.
		`),
	}

	root.AddCommand(
		newRunCommand(),
	)

	f := flag.CommandLine.Lookup("v")
	root.PersistentFlags().AddGoFlag(f)
	pflag.CommandLine = pflag.NewFlagSet("empty", pflag.ExitOnError)
	flag.CommandLine = flag.NewFlagSet("empty", flag.ExitOnError)

	e2e.RegisterCommonFlags(flag.CommandLine)
	e2e.RegisterClusterFlags(flag.CommandLine)

	if err := func() error {
		// TODO: import library-go to support this?
		// defer serviceability.Profile(os.Getenv("OPENSHIFT_PROFILE")).Stop()
		return root.Execute()
	}(); err != nil {
		if ex, ok := err.(testginkgo.ExitError); ok {
			os.Exit(ex.Code)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// mirrorToFile ensures a copy of all output goes to the provided OutFile, including
// any error returned from fn. The function returns fn() or any error encountered while
// attempting to open the file.
func mirrorToFile(opt *testginkgo.Options, fn func() error) error {
	if opt.Out == nil {
		opt.Out = os.Stdout
	}
	if opt.ErrOut == nil {
		opt.ErrOut = os.Stderr
	}
	if len(opt.OutFile) == 0 {
		return fn()
	}

	f, err := os.OpenFile(opt.OutFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	opt.Out = io.MultiWriter(opt.Out, f)
	opt.ErrOut = io.MultiWriter(opt.ErrOut, f)
	exitErr := fn()
	if exitErr != nil {
		fmt.Fprintf(f, "error: %s", exitErr)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintf(opt.ErrOut, "error: Unable to close output file\n")
	}
	return exitErr
}

func verifyImages() error {
	if len(os.Getenv("KUBE_TEST_REPO")) > 0 {
		return fmt.Errorf("KUBE_TEST_REPO may not be specified when this command is run")
	}
	return verifyImagesWithoutEnv()
}

func verifyImagesWithoutEnv() error {
	defaults := k8simage.GetOriginalImageConfigs()

	for originalPullSpec, index := range image.OriginalImages() {
		if index == -1 {
			continue
		}
		existing, ok := defaults[index]
		if !ok {
			return fmt.Errorf("image %q not found in upstream images, must be moved to test/extended/util/image", originalPullSpec)
		}
		if existing.GetE2EImage() != originalPullSpec {
			return fmt.Errorf("image %q defines index %d but is defined upstream as %q, must be fixed in test/extended/util/image", originalPullSpec, index, existing.GetE2EImage())
		}
		mirror := image.LocationFor(originalPullSpec)
		upstreamMirror := k8simage.GetE2EImage(index)
		if mirror != upstreamMirror {
			return fmt.Errorf("image %q defines index %d and mirror %q but is mirrored upstream as %q, must be fixed in test/extended/util/image", originalPullSpec, index, mirror, upstreamMirror)
		}
	}

	return nil
}

func newRunCommand() *cobra.Command {
	opt := &runOptions{
		FromRepository: defaultTestImageMirrorLocation,
	}

	cmd := &cobra.Command{
		Use:   "run SUITE",
		Short: "Run a test suite",
		Long: templates.LongDesc(`
		Run a test suite against an OpenShift server

		This command will run one of the following suites against a cluster identified by the current
		KUBECONFIG file. See the suite description for more on what actions the suite will take.

		If you specify the --dry-run argument, the names of each individual test that is part of the
		suite will be printed, one per line. You may filter this list and pass it back to the run
		command with the --file argument. You may also pipe a list of test names, one per line, on
		standard input by passing "-f -".

		`) + testginkgo.SuitesString(staticSuites.TestSuites(), "\n\nAvailable test suites:\n\n"),

		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mirrorToFile(&opt.Options, func() error {
				if err := verifyImages(); err != nil {
					return err
				}

				// opt.SyntheticEventTests = pulledInvalidImages(opt.FromRepository)

				suite, err := opt.SelectSuite(staticSuites, args)
				if err != nil {
					return err
				}
				if suite.PreSuite != nil {
					if err := suite.PreSuite(opt); err != nil {
						return err
					}
				}
				opt.CommandEnv = opt.AsEnv()
				if !opt.DryRun {
					fmt.Fprintf(os.Stderr, "%s version: %s\n", filepath.Base(os.Args[0]), version.Get().String())
				}
				err = opt.Run(&suite.TestSuite)
				if suite.PostSuite != nil {
					suite.PostSuite(opt)
				}
				return err
			})
		},
	}
	bindOptions(opt, cmd.Flags())
	return cmd
}

func bindOptions(opt *runOptions, flags *pflag.FlagSet) {
	flags.StringVar(&opt.FromRepository, "from-repository", opt.FromRepository, "A container image repository to retrieve test images from.")
	flags.StringVar(&opt.Provider, "provider", opt.Provider, "The cluster infrastructure provider. Will automatically default to the correct value.")
	bindTestOptions(&opt.Options, flags)
}

func bindTestOptions(opt *testginkgo.Options, flags *pflag.FlagSet) {
	flags.BoolVar(&opt.DryRun, "dry-run", opt.DryRun, "Print the tests to run without executing them.")
	flags.BoolVar(&opt.PrintCommands, "print-commands", opt.PrintCommands, "Print the sub-commands that would be executed instead.")
	flags.StringVar(&opt.JUnitDir, "junit-dir", opt.JUnitDir, "The directory to write test reports to.")
	flags.StringVarP(&opt.TestFile, "file", "f", opt.TestFile, "Create a suite from the newline-delimited test names in this file.")
	flags.StringVar(&opt.Regex, "run", opt.Regex, "Regular expression of tests to run.")
	flags.StringVarP(&opt.OutFile, "output-file", "o", opt.OutFile, "Write all test output to this file.")
	flags.IntVar(&opt.Count, "count", opt.Count, "Run each test a specified number of times. Defaults to 1 or the suite's preferred value. -1 will run forever.")
	flags.BoolVar(&opt.FailFast, "fail-fast", opt.FailFast, "If a test fails, exit immediately.")
	flags.DurationVar(&opt.Timeout, "timeout", opt.Timeout, "Set the maximum time a test can run before being aborted. This is read from the suite by default, but will be 10 minutes otherwise.")
	flags.BoolVar(&opt.IncludeSuccessOutput, "include-success", opt.IncludeSuccessOutput, "Print output from successful tests.")
	flags.IntVar(&opt.Parallelism, "max-parallel-tests", opt.Parallelism, "Maximum number of tests running in parallel. 0 defaults to test suite recommended value, which is different in each suite.")
}
