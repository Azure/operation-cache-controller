package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sresources "sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	v1 "github.com/Azure/operation-cache-controller/api/v1"
	"github.com/Azure/operation-cache-controller/test/utils/resources"
)

type cacheKey struct{}

var CacheFeature = features.New("appsv1/deployment/cache").
	Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		// start a deployment
		cache := resources.SampleCache
		cache.Namespace = testNamespace
		if err := c.Client().Resources().Create(ctx, &cache); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		return ctx
	}).
	Assess("create cache", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var cache v1.Cache
		if err := cfg.Client().Resources().Get(ctx, resources.SampleCache.Name, testNamespace, &cache); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "cache-sample", cache.Name)
		if err := wait.PollUntilContextTimeout(ctx, 5*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			var ops v1.OperationList
			if err := cfg.Client().Resources().List(ctx, &ops, k8sresources.WithFieldSelector(fmt.Sprintf("metadata.namespace=%s", testNamespace))); err != nil {
				return false, err
			}
			var countOwnedOps int
			for _, op := range ops.Items {
				op := op
				if op.ObjectMeta.GetOwnerReferences()[0].Name == cache.Name {
					countOwnedOps++
				}
			}
			if countOwnedOps != int(cache.Status.KeepAliveCount) {
				return false, nil
			}
			return true, nil
		}); err != nil {
			t.Fatal(err, "operations not meet the expected count")
		}
		return context.WithValue(ctx, cacheKey{}, &cache)
	}).
	Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		cache := ctx.Value(cacheKey{}).(*v1.Cache)
		if err := cfg.Client().Resources().Delete(ctx, cache); err != nil {
			t.Fatal(err)
		}
		return ctx
	}).Feature()
