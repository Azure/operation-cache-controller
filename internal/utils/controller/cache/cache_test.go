package cache

import (
	"testing"

	"github.com/stretchr/testify/require"

	appsv1 "github.com/Azure/operation-cache-controller/api/v1"
)

func TestRandomSelectCachedOperation(t *testing.T) {
	tests := []struct {
		name        string
		caches      []string
		expectEmpty bool
	}{
		{"empty caches", nil, true},
		{"non-empty caches", []string{"cache1", "cache2", "cache3"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ...existing setup code if any...
			cacheInstance := &appsv1.Cache{
				Status: appsv1.CacheStatus{
					AvailableCaches: tt.caches,
				},
			}

			result := RandomSelectCachedOperation(cacheInstance)

			if tt.expectEmpty {
				require.Equal(t, "", result)
			} else {
				require.Contains(t, tt.caches, result)
			}
			// ...existing teardown code if any...
		})
	}
}

func TestDefaultCacheExpireTime(t *testing.T) {
	result := DefaultCacheExpireTime()
	require.NotEmpty(t, result)
}
