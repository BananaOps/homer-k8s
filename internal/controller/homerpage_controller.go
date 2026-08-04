/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	homerv1alpha1 "github.com/BananaOps/homer-k8s/api/v1alpha1"
	homerconfig "github.com/BananaOps/homer-k8s/pkg/config"
)

// HomerPageReconciler reconciles a HomerPage object
type HomerPageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homer.bananaops.io,resources=homerpages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homer.bananaops.io,resources=homerpages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homer.bananaops.io,resources=homerpages/finalizers,verbs=update

// Reconcile watches HomerPage resources and writes the corresponding page YAML file
// into the Homer assets directory. When autoLink is enabled, it also keeps the link
// entry in config.yml in sync so that the page appears in the Homer navbar.
func (r *HomerPageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	pageLogCtx := []interface{}{
		"controller", "homerpage",
		"controllerGroup", "homer.bananaops.io",
		"controllerKind", "HomerPage",
		"name", req.Name,
		"namespace", req.Namespace,
	}

	// Resolve and sanitize the config directory (path-traversal guard, CWE-22).
	configDir := os.Getenv("HOMER_CONFIG_DIR")
	if configDir == "" {
		configDir = "/assets"
	}
	configDir = filepath.Clean(configDir)

	// --- Fetch the HomerPage resource ---
	var homerPage homerv1alpha1.HomerPage
	if err := r.Get(ctx, req.NamespacedName, &homerPage); err != nil {
		// If the resource no longer exists there is nothing to reconcile.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pageName := homerPage.Spec.PageName

	// Validate pageName before using it in a file path (defence-in-depth, the
	// CRD pattern validation is the primary guard but we double-check here).
	if strings.ContainsAny(pageName, "/\\..") || pageName == "" {
		return ctrl.Result{}, fmt.Errorf("invalid pageName %q: must not contain path separators or dots", pageName)
	}

	pageFile := filepath.Join(configDir, pageName+".yml")

	// Extra guard: ensure the resolved path stays within configDir.
	if !strings.HasPrefix(pageFile, configDir+string(filepath.Separator)) {
		return ctrl.Result{}, fmt.Errorf("page file path %q is outside config directory %q", pageFile, configDir)
	}

	// --- Build the page config ---
	pageConfig := homerconfig.HomerPageConfig{
		Services: homerPage.Spec.Groups,
	}

	pageBytes, err := yaml.Marshal(pageConfig)
	if err != nil {
		logger.Error(err, "failed to marshal page config", pageLogCtx...)
		return ctrl.Result{}, err
	}

	// Read the existing page file to detect changes.
	existingPageBytes, _ := os.ReadFile(pageFile)
	if !reflect.DeepEqual(existingPageBytes, pageBytes) {
		if err := os.WriteFile(pageFile, pageBytes, 0600); err != nil {
			logger.Error(err, "failed to write page file", pageLogCtx...)
			return ctrl.Result{}, err
		}
		logger.Info("Homer Page file updated", append(pageLogCtx, "file", pageFile)...)
	}

	// --- autoLink: keep config.yml links in sync ---
	if err := r.reconcileAutoLink(ctx, configDir, pageLogCtx); err != nil {
		return ctrl.Result{}, err
	}

	// --- Update status ---
	homerPage.Status.Ready = true
	if err := r.Status().Update(ctx, &homerPage); err != nil {
		// Non-fatal: log and continue.
		logger.Error(err, "failed to update HomerPage status", pageLogCtx...)
	}

	return ctrl.Result{}, nil
}

// reconcileAutoLink reads all HomerPage resources that have AutoLink=true and
// rewrites the `links` section of config.yml accordingly. Links that were
// added by other means (i.e., present in global_config.yml) are preserved.
func (r *HomerPageReconciler) reconcileAutoLink(ctx context.Context, configDir string, logCtx []interface{}) error {
	// Collect all HomerPages with autoLink enabled.
	var pageList homerv1alpha1.HomerPageList
	if err := r.List(ctx, &pageList); err != nil {
		return err
	}

	// Build the desired set of auto-managed links (keyed by url "#<pageName>").
	desiredLinks := map[string]homerconfig.Link{}
	for _, p := range pageList.Items {
		if !p.Spec.AutoLink {
			continue
		}
		url := "#" + p.Spec.PageName
		name := p.Spec.LinkName
		if name == "" {
			name = p.Spec.PageName
		}
		desiredLinks[url] = homerconfig.Link{
			Name: name,
			Icon: p.Spec.LinkIcon,
			URL:  url,
		}
	}

	// Read the current config.yml.
	configPath := filepath.Join(configDir, "config.yml")
	if !strings.HasPrefix(configPath, configDir+string(filepath.Separator)) {
		return fmt.Errorf("invalid config path: %s", configPath)
	}

	var globalConfig homerconfig.HomerConfig
	rawConfig, _ := os.ReadFile("/config/global_config.yml")
	if err := yaml.Unmarshal(rawConfig, &globalConfig); err != nil {
		logger.Error(err, "failed to unmarshal global_config.yml", logCtx...)
	}

	var currentConfig homerconfig.HomerConfig
	rawCurrent, _ := os.ReadFile(configPath)
	if err := yaml.Unmarshal(rawCurrent, &currentConfig); err != nil {
		logger.Error(err, "failed to unmarshal config.yml", logCtx...)
	}

	// Build the new links list: keep non-auto links, then add/update auto links.
	// A link is considered "non-auto" if its url does not match any current or
	// desired auto-managed url (i.e., does not start with "#").
	autoURLs := map[string]bool{}
	for url := range desiredLinks {
		autoURLs[url] = true
	}
	// Also consider urls of pages that no longer have autoLink (so we can remove them).
	for _, l := range currentConfig.Links {
		if strings.HasPrefix(l.URL, "#") {
			autoURLs[l.URL] = true
		}
	}

	// Start from global_config links as the base, then preserve manually-added
	// links from config.yml that are not in the auto-managed set.
	newLinks := make([]homerconfig.Link, 0, len(globalConfig.Links)+len(desiredLinks))
	newLinks = append(newLinks, globalConfig.Links...)
	for _, l := range currentConfig.Links {
		if !autoURLs[l.URL] {
			newLinks = append(newLinks, l)
		}
	}
	// Append the desired auto-managed links.
	for _, l := range desiredLinks {
		newLinks = append(newLinks, l)
	}

	// Merge services from global_config into currentConfig.
	currentConfig.Links = newLinks
	// Preserve non-links fields from global_config where currentConfig is empty.
	mergeGlobalIntoConfig(&currentConfig, &globalConfig)

	newBytes, err := yaml.Marshal(currentConfig)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(rawCurrent, newBytes) {
		if err := os.WriteFile(configPath, newBytes, 0600); err != nil {
			return err
		}
		logger.Info("Homer config.yml links updated", logCtx...)
	}

	return nil
}

// mergeGlobalIntoConfig copies non-zero fields from src into dst only when dst
// has a zero value. This ensures global_config.yml settings are the fallback.
func mergeGlobalIntoConfig(dst, src *homerconfig.HomerConfig) {
	if dst.Title == "" {
		dst.Title = src.Title
	}
	if dst.Subtitle == "" {
		dst.Subtitle = src.Subtitle
	}
	if dst.Logo == "" {
		dst.Logo = src.Logo
	}
	if dst.Icon == "" {
		dst.Icon = src.Icon
	}
	if dst.Footer == "" {
		dst.Footer = src.Footer
	}
	if dst.Columns == "" {
		dst.Columns = src.Columns
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *HomerPageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homerv1alpha1.HomerPage{}).
		Complete(r)
}
