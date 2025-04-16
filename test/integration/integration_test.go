package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	v1 "github.com/Azure/operation-cache-controller/api/v1"
	rqutils "github.com/Azure/operation-cache-controller/internal/utils/controller/requirement"
	"github.com/Azure/operation-cache-controller/test/utils"
)

type requirementKey struct{}

const (
	testRequirementName = "test-requirement"
)

var CacheFeature = features.New("Simple Requirements").
	Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		// start a deployment
		requiremnt := utils.NewRequirement(testRequirementName, utils.TestNamespcae)
		requiremnt.Namespace = utils.TestNamespcae
		if err := c.Client().Resources().Create(ctx, requiremnt); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		return ctx
	}).
	Assess("create requirement", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var requirement v1.Requirement
		if err := cfg.Client().Resources().Get(ctx, testRequirementName, utils.TestNamespcae, &requirement); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, testRequirementName, requirement.Name)
		if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			requirement := &v1.Requirement{}
			if err := cfg.Client().Resources().Get(ctx, testRequirementName, utils.TestNamespcae, requirement); err != nil {
				return false, err
			}
			if requirement.Status.Phase != rqutils.PhaseReady {
				return false, nil
			}
			return true, nil
		}); err != nil {
			t.Fatal(err, "operations not ready")
		}
		return context.WithValue(ctx, requirementKey{}, &requirement)
	}).
	Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		requirement := ctx.Value(requirementKey{}).(*v1.Requirement)
		if err := cfg.Client().Resources().Delete(ctx, requirement); err != nil {
			t.Fatal(err)
		}
		return ctx
	}).Feature()
