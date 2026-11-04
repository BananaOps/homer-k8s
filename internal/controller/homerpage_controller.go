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
	"io"
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

	// Resolve the config directory.
	configDir := os.Getenv("HOMER_CONFIG_DIR")
	if configDir == "" {
		configDir = "/assets"
	}
	configDir = filepath.Clean(configDir)

	// Open a scoped root for the config directory (Go >=1.24, CWE-22).
	// All subsequent file operations go through this root handle, so the kernel
	// rejects any path that escapes the directory — even via symlinks or ".." components.
	root, err := os.OpenRoot(configDir)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("cannot open config root %q: %w", configDir, err)
	}
	defer func() {
		if cerr := root.Close(); cerr != nil {
			logger.Error(cerr, "failed to close config root", pageLogCtx...)
		}
	}()

	// --- Fetch the HomerPage resource ---
	var homerPage homerv1alpha1.HomerPage
	if err := r.Get(ctx, req.NamespacedName, &homerPage); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	pageName := homerPage.Spec.PageName

	// Defence-in-depth: the CRD regex is the primary guard, but we also reject
	// names containing path-sensitive characters before they reach the filesystem.
	if strings.ContainsAny(pageName, "/\\..") || pageName == "" {
		return ctrl.Result{}, fmt.Errorf("invalid pageName %q: must not contain path separators or dots", pageName)
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

	// Read the existing page file through the scoped root to detect changes.
	pageFileName := pageName + ".yml"
	existingPageBytes, _ := rootReadFile(root, pageFileName)
	if !reflect.DeepEqual(existingPageBytes, pageBytes) {
		if err := rootWriteFile(root, pageFileName, pageBytes, 0600); err != nil {
			logger.Error(err, "failed to write page file", pageLogCtx...)
			return ctrl.Result{}, err
		}
		logger.Info("Homer Page file updated", append(pageLogCtx, "file", filepath.Join(configDir, pageFileName))...)
	}

	// --- autoLink: keep config.yml links in sync ---
	if err := r.reconcileAutoLink(ctx, root, configDir, pageLogCtx); err != nil {
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
// All file access goes through the scoped root handle (CWE-22).
func (r *HomerPageReconciler) reconcileAutoLink(ctx context.Context, root *os.Root, configDir string, logCtx []interface{}) error {
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

	// global_config.yml lives outside the scoped root (/config/), read it directly.
	// Its path is hard-coded and not user-controlled, so os.ReadFile is safe here.
	var globalConfig homerconfig.HomerConfig
	rawConfig, _ := os.ReadFile("/config/global_config.yml") // #nosec G304 — fixed path
	if err := yaml.Unmarshal(rawConfig, &globalConfig); err != nil {
		logger.Error(err, "failed to unmarshal global_config.yml", logCtx...)
	}

	// Read config.yml through the scoped root.
	var currentConfig homerconfig.HomerConfig
	rawCurrent, _ := rootReadFile(root, "config.yml")
	if err := yaml.Unmarshal(rawCurrent, &currentConfig); err != nil {
		logger.Error(err, "failed to unmarshal config.yml", logCtx...)
	}

	// Build the new links list: keep non-auto links, then add/update auto links.
	autoURLs := map[string]bool{}
	for url := range desiredLinks {
		autoURLs[url] = true
	}
	// Also track urls of pages that no longer have autoLink (so we can remove them).
	for _, l := range currentConfig.Links {
		if strings.HasPrefix(l.URL, "#") {
			autoURLs[l.URL] = true
		}
	}

	newLinks := make([]homerconfig.Link, 0, len(globalConfig.Links)+len(desiredLinks))
	newLinks = append(newLinks, globalConfig.Links...)
	for _, l := range currentConfig.Links {
		if !autoURLs[l.URL] {
			newLinks = append(newLinks, l)
		}
	}
	for _, l := range desiredLinks {
		newLinks = append(newLinks, l)
	}

	currentConfig.Links = newLinks
	mergeGlobalIntoConfig(&currentConfig, &globalConfig)

	newBytes, err := yaml.Marshal(currentConfig)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(rawCurrent, newBytes) {
		// Write config.yml through the scoped root.
		if err := rootWriteFile(root, "config.yml", newBytes, 0600); err != nil {
			return err
		}
		logger.Info("Homer config.yml links updated", logCtx...)
	}

	return nil
}

// rootReadFile reads a file by name through an os.Root handle.
// The name must be a plain filename (no path separators); the root handle
// ensures the access cannot escape the configured directory.
func rootReadFile(root *os.Root, name string) ([]byte, error) {
	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	// Always close; if both ReadAll and Close fail, return the ReadAll error.
	if cerr := f.Close(); cerr != nil && err == nil {
		err = cerr
	}
	return data, err
}

// rootWriteFile writes data to a file by name through an os.Root handle,
// creating or truncating it. Permissions are set to perm.
func rootWriteFile(root *os.Root, name string, data []byte, perm os.FileMode) error {
	f, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	return cerr
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
