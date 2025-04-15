package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"

	v1 "github.com/Azure/operation-cache-controller/api/v1"
	"github.com/Azure/operation-cache-controller/test/utils"
)

var testenv env.Environment

// projectImage is the name of the image which will be build and loaded
// with the code source changes to be tested.
var projectImage = "example.com/operation-cache-controller:v0.0.1"
var kindClusterName = "integration-test-cluster"
var testNamespace = "operation-cache-controller-test"

func init() {
	log.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))
	utilruntime.Must(v1.AddToScheme(scheme.Scheme))
}

func TestMain(m *testing.M) {
	// Create a new test environment configuration
	testenv = env.New()

	// Setup the test environment with Kind cluster and necessary resources
	testenv = testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.LoadDockerImageToCluster(kindClusterName, projectImage),
		envfuncs.CreateNamespace(testNamespace),
		BuildImage,
		InstallCRD,
		DeployControllerManager,
	)

	// Teardown the test environment
	testenv = testenv.Finish(
		envfuncs.DeleteNamespace(testNamespace),
		UninstallCRD,
		envfuncs.DestroyCluster(kindClusterName),
	)

	// Run the tests
	os.Exit(testenv.Run(m))
}

func BuildImage(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	// Build the Docker image for the controller manager
	cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectImage))
	_, err := utils.Run(cmd)
	return ctx, err
}

func InstallCRD(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	// Install the CRD in the test environment
	cmd := exec.Command("make", "install")
	_, err := utils.Run(cmd)
	return ctx, err
}

func DeployControllerManager(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	// Deploy the controller manager in the test environment
	cmd := exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectImage))
	_, err := utils.Run(cmd)
	return ctx, err
}

func UninstallCRD(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	// Uninstall the CRD in the test environment
	cmd := exec.Command("make", "uninstall")
	_, err := utils.Run(cmd)
	return ctx, err
}
func TestRealCluster(t *testing.T) {
	// Run the integration tests against the Kind cluster
	testenv.Test(t, CacheFeature)
}
