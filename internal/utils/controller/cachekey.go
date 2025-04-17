package controller

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"

	"github.com/Azure/operation-cache-controller/api/v1alpha1"
)

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

func NewCacheKeyFromApplications(apps []v1alpha1.ApplicationSpec) string {
	// sort the apps by name to ensure consistent hashing
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Name < apps[j].Name
	})

	srcCacheKeys := lo.Reduce(apps, func(acc []string, app v1alpha1.ApplicationSpec, index int) []string {
		return append(acc, AppCacheFieldFromApplicationProvision(app).NewCacheKey())
	}, []string{})

	// get the cache id for the source
	hasher := sha256.New()
	for _, id := range srcCacheKeys {
		hasher.Write([]byte(id))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func AppCacheFieldFromApplicationProvision(app v1alpha1.ApplicationSpec) *AppCacheField {
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

func AppCacheFieldFromApplicationTeardown(app v1alpha1.ApplicationSpec) *AppCacheField {
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
