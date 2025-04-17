package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
	rqutils "github.com/Azure/operation-cache-controller/internal/utils/controller/requirement"
	"github.com/Azure/operation-cache-controller/test/utils"
)

type requirementKey struct{}

const (
	testRequirementName   = "test-requirement"
	cachedRequirementName = "cached-requirement"
)

var SimpleRequirementFeature = features.New("Simple Requirements").
	Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		// start a deployment
		testRequirement := utils.NewRequirement(testRequirementName, utils.TestNamespace)
		testRequirement.Namespace = utils.TestNamespace
		if err := c.Client().Resources().Create(ctx, testRequirement); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		return ctx
	}).
	Assess("requirement created successfully", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var requirement v1alpha1.Requirement
		if err := cfg.Client().Resources().Get(ctx, testRequirementName, utils.TestNamespace, &requirement); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, testRequirementName, requirement.Name)
		if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			requirement := &v1alpha1.Requirement{}
			if err := cfg.Client().Resources().Get(ctx, testRequirementName, utils.TestNamespace, requirement); err != nil {
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
		requirement := ctx.Value(requirementKey{}).(*v1alpha1.Requirement)
		if err := cfg.Client().Resources().Delete(ctx, requirement); err != nil {
			t.Fatal(err)
		}
		return ctx
	}).Feature()

func newRequirementWithCache(name string) *v1alpha1.Requirement {
	requirement := utils.NewRequirement(name, utils.TestNamespace)
	requirement.Spec.EnableCache = true
	return requirement
}

var CachedRequirementFeature = features.New("Cached Requirements").
	Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		// Create a requirement with cache enabled
		if err := c.Client().Resources().Create(ctx, newRequirementWithCache(cachedRequirementName)); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)
		return ctx
	}).
	Assess("cache requirement created and synced", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var requirement v1alpha1.Requirement
		if err := cfg.Client().Resources().Get(ctx, cachedRequirementName, utils.TestNamespace, &requirement); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, cachedRequirementName, requirement.Name)

		// Wait for requirement to be ready
		if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			requirement := &v1alpha1.Requirement{}
			if err := cfg.Client().Resources().Get(ctx, cachedRequirementName, utils.TestNamespace, requirement); err != nil {
				return false, err
			}
			if requirement.Status.Phase != rqutils.PhaseReady {
				return false, nil
			}
			return true, nil
		}); err != nil {
			t.Fatal(err, "requirement not ready")
		}
		cacheKey := requirement.Status.CacheKey
		// Get the associated Cache resource
		cache := &v1alpha1.Cache{}
		cacheName := "cache-" + cacheKey
		if err := cfg.Client().Resources().Get(ctx, cacheName, utils.TestNamespace, cache); err != nil {
			t.Fatal(err, "cache not found")
		}
		// Verify the cache key matches
		assert.Equal(t, cacheKey, cache.Status.CacheKey)

		// Get all Operations
		var operations v1alpha1.OperationList
		if err := cfg.Client().Resources().List(ctx, &operations); err != nil {
			t.Fatal(err, "failed to list operations")
		}

		// Verify one operation is owned by our requirement by checking owner references
		var (
			ownedByRequirement []v1alpha1.Operation
			ownedByCache       []v1alpha1.Operation
		)
		for _, op := range operations.Items {
			for _, ownerRef := range op.OwnerReferences {
				if ownerRef.APIVersion == v1alpha1.GroupVersion.String() &&
					ownerRef.Kind == "Requirement" &&
					ownerRef.Name == requirement.Name &&
					ownerRef.UID == requirement.UID {
					ownedByRequirement = append(ownedByRequirement, op)
				}
				if ownerRef.APIVersion == v1alpha1.GroupVersion.String() &&
					ownerRef.Kind == "Cache" &&
					ownerRef.Name == cache.Name &&
					ownerRef.UID == cache.UID {
					ownedByCache = append(ownedByCache, op)
				}
			}
		}
		// Verify one operation is owned by our requirement
		assert.Equal(t, 1, len(ownedByRequirement), "expected one operation owned by requirement")
		// Verify number of cache operations matches keepAlive count
		assert.Equal(t, int(cache.Status.KeepAliveCount), len(ownedByCache))

		// delete the requirement
		if err := cfg.Client().Resources().Delete(ctx, &requirement); err != nil {
			t.Fatal(err)
		}
		// wait for the requirement to be deleted
		if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			requirement := &v1alpha1.Requirement{}
			err := cfg.Client().Resources().Get(ctx, cachedRequirementName, utils.TestNamespace, requirement)
			// err is not found, so requirement is deleted
			if err != nil {
				if apierr.IsNotFound(err) {
					return true, nil
				}
				return false, err
			}
			return false, nil
		}); err != nil {
			t.Fatal(err, "requirement not deleted")
		}

		newCachedRequirementName := cachedRequirementName + "-new"
		// create a new requirement with the same name
		if err := cfg.Client().Resources().Create(ctx, newRequirementWithCache(newCachedRequirementName)); err != nil {
			t.Fatal(err)
		}

		newRequirement := &v1alpha1.Requirement{}
		// wait for the new requirement to be ready
		if err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			if err := cfg.Client().Resources().Get(ctx, newCachedRequirementName, utils.TestNamespace, newRequirement); err != nil {
				return false, err
			}
			if newRequirement.Status.Phase != rqutils.PhaseReady {
				return false, nil
			}
			return true, nil
		}); err != nil {
			t.Fatal(err, "new requirement not ready")
		}

		// cache should be hit
		cacheHit := false
		for _, condition := range newRequirement.Status.Conditions {
			if condition.Type == rqutils.ConditionOperationReady {
				cacheHit = true
				break
			}
		}
		assert.True(t, cacheHit, "expected cache hit condition")

		// list operations and verify the new requirement has operation with the same name from original cache
		var newOperations v1alpha1.OperationList
		if err := cfg.Client().Resources().List(ctx, &newOperations); err != nil {
			t.Fatal(err, "failed to list operations")
		}
		var (
			ownedByNewRequirement []v1alpha1.Operation
			ownedByCurrentCache   []v1alpha1.Operation
		)

		for _, op := range newOperations.Items {
			for _, ownerRef := range op.OwnerReferences {
				if ownerRef.APIVersion == v1alpha1.GroupVersion.String() &&
					ownerRef.Kind == "Requirement" &&
					ownerRef.Name == newCachedRequirementName &&
					ownerRef.UID == newRequirement.UID {
					ownedByNewRequirement = append(ownedByNewRequirement, op)
				}
				if ownerRef.APIVersion == v1alpha1.GroupVersion.String() &&
					ownerRef.Kind == "Cache" &&
					ownerRef.Name == cacheName &&
					ownerRef.UID == cache.UID {
					ownedByCurrentCache = append(ownedByCurrentCache, op)
				}
			}
		}
		// Verify one operation is owned by our requirement
		assert.Equal(t, 1, len(ownedByNewRequirement), "expected one operation owned by requirement")
		// Verify number of cache operations matches keepAlive count
		assert.Equal(t, int(cache.Status.KeepAliveCount), len(ownedByCurrentCache))

		// Verify the operation come from the original cache
		found := false
		for _, op := range ownedByCache {
			if op.Name == ownedByNewRequirement[0].Name {
				found = true
				break
			}
		}
		assert.True(t, found, "expected operation to be from original cache")

		// the operation should not be included in the new cache
		for _, op := range ownedByCurrentCache {
			assert.NotEqual(t, op.Name, ownedByNewRequirement[0].Name, "operation should not be included in the new cache")
		}
		return context.WithValue(ctx, requirementKey{}, newRequirement)
	}).
	Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		requirement := ctx.Value(requirementKey{}).(*v1alpha1.Requirement)
		if err := cfg.Client().Resources().Delete(ctx, requirement); err != nil {
			t.Fatal(err)
		}
		return ctx
	}).Feature()
