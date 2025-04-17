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

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	"github.com/Azure/operation-cache-controller/test/utils"
)

var testenv env.Environment

var kindClusterName = "integration-test-cluster"

func init() {
	log.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
}

func TestMain(m *testing.M) {
	// Create a new test environment configuration
	testenv = env.New()

	// Setup the test environment with Kind cluster and necessary resources
	testenv = testenv.Setup(
		BuildImage,
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.LoadDockerImageToCluster(kindClusterName, utils.ProjectImage),
		envfuncs.CreateNamespace(utils.TestNamespace),
		InstallCRD,
		DeployControllerManager,
	)

	// Teardown the test environment
	testenv = testenv.Finish(
		envfuncs.DeleteNamespace(utils.TestNamespace),
		envfuncs.DestroyCluster(kindClusterName),
	)

	// Run the tests
	os.Exit(testenv.Run(m))
}

func BuildImage(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	// Build the Docker image for the controller manager
	cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", utils.ProjectImage))
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
	cmd := exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", utils.ProjectImage))
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
	// Create a new test environment configuration
	// Run the integration tests against the Kind cluster
	testenv.Test(t, SimpleRequirementFeature, CachedRequirementFeature)
}
