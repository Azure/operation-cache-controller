package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

type CacheHelper struct{}

func NewCacheHelper() CacheHelper { return CacheHelper{} }

func (c CacheHelper) RandomSelectCachedOperation(cache *v1alpha1.Cache) string {
	if len(cache.Status.AvailableCaches) == 0 {
		return ""
	}
	// nolint:gosec, G404 // this is expected PRNG usage
	return cache.Status.AvailableCaches[rand.Intn(len(cache.Status.AvailableCaches))]
}
func (c CacheHelper) DefaultCacheExpireTime() string {
	// cache expire after 2 hours
	return time.Now().Add(2 * time.Hour).Format(time.RFC3339)
}

type AppCacheField struct {
	Name         string
	Image        string
	Command      []string
	Args         []string
	WorkingDir   string
	Env          []corev1.EnvVar
	Dependencies []string
}

func (c *AppCacheField) NewCacheKey() string {
	hasher := sha256.New()
	hasher.Write([]byte(c.Name))
	hasher.Write([]byte(c.Image))
	hasher.Write([]byte(strings.Join(c.Command, " ")))
	hasher.Write([]byte(strings.Join(c.Args, " ")))
	hasher.Write([]byte(c.WorkingDir))

	// Sort environment variables to ensure consistent hashing
	sort.Slice(c.Env, func(i, j int) bool {
		return c.Env[i].Name < c.Env[j].Name
	})
	for _, env := range c.Env {
		hasher.Write([]byte(env.Name + "=" + env.Value))
	}

	// Sort dependencies to ensure consistent hashing
	sort.Strings(c.Dependencies)
	for _, dep := range c.Dependencies {
		hasher.Write([]byte(dep))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

func (c CacheHelper) NewCacheKeyFromApplications(apps []v1alpha1.ApplicationSpec) string {
	// sort the apps by name to ensure consistent hashing
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Name < apps[j].Name
	})

	srcCacheKeys := lo.Reduce(apps, func(acc []string, app v1alpha1.ApplicationSpec, index int) []string {
		return append(acc, c.AppCacheFieldFromApplicationProvision(app).NewCacheKey())
	}, []string{})

	// get the cache id for the source
	hasher := sha256.New()
	for _, id := range srcCacheKeys {
		hasher.Write([]byte(id))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (c CacheHelper) AppCacheFieldFromApplicationProvision(app v1alpha1.ApplicationSpec) *AppCacheField {
	return &AppCacheField{
		Name:         app.Name,
		Image:        app.Provision.Template.Spec.Containers[0].Image,
		Command:      app.Provision.Template.Spec.Containers[0].Command,
		Args:         app.Provision.Template.Spec.Containers[0].Args,
		WorkingDir:   app.Provision.Template.Spec.Containers[0].WorkingDir,
		Env:          app.Provision.Template.Spec.Containers[0].Env,
		Dependencies: app.Dependencies,
	}
}

func (c CacheHelper) AppCacheFieldFromApplicationTeardown(app v1alpha1.ApplicationSpec) *AppCacheField {
	return &AppCacheField{
		Name:         app.Name,
		Image:        app.Teardown.Template.Spec.Containers[0].Image,
		Command:      app.Teardown.Template.Spec.Containers[0].Command,
		Args:         app.Teardown.Template.Spec.Containers[0].Args,
		WorkingDir:   app.Teardown.Template.Spec.Containers[0].WorkingDir,
		Env:          app.Teardown.Template.Spec.Containers[0].Env,
		Dependencies: app.Dependencies,
	}
}
