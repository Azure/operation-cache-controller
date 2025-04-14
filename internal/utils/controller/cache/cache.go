package cache

import (
	"math/rand"
	"time"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
)

func RandomSelectCachedOperation(cache *appsv1.Cache) string {
	if len(cache.Status.AvailableCaches) == 0 {
		return ""
	}
	// nolint:gosec, G404 // this is expected PRNG usage
	return cache.Status.AvailableCaches[rand.Intn(len(cache.Status.AvailableCaches))]
}

func DefaultCacheExpireTime() string {
	// cache expire after 2 hours
	return time.Now().Add(2 * time.Hour).Format(time.RFC3339)
}
