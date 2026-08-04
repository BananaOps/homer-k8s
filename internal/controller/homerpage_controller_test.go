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
	"os"
	"path/filepath"
	"testing"

	homerv1alpha1 "github.com/BananaOps/homer-k8s/api/v1alpha1"
	homerconfig "github.com/BananaOps/homer-k8s/pkg/config"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// ── mergeGlobalIntoConfig ─────────────────────────────────────────────────────

func TestMergeGlobalIntoConfig_FillsEmptyFields(t *testing.T) {
	dst := &homerconfig.HomerConfig{}
	src := &homerconfig.HomerConfig{
		Title:    "My Homer",
		Subtitle: "Dashboard",
		Logo:     "assets/logo.png",
		Icon:     "fas fa-home",
		Footer:   "<p>footer</p>",
		Columns:  "3",
	}
	mergeGlobalIntoConfig(dst, src)

	if dst.Title != "My Homer" {
		t.Errorf("Title: want %q, got %q", "My Homer", dst.Title)
	}
	if dst.Subtitle != "Dashboard" {
		t.Errorf("Subtitle: want %q, got %q", "Dashboard", dst.Subtitle)
	}
	if dst.Logo != "assets/logo.png" {
		t.Errorf("Logo: want %q, got %q", "assets/logo.png", dst.Logo)
	}
	if dst.Icon != "fas fa-home" {
		t.Errorf("Icon: want %q, got %q", "fas fa-home", dst.Icon)
	}
	if dst.Footer != "<p>footer</p>" {
		t.Errorf("Footer: want %q, got %q", "<p>footer</p>", dst.Footer)
	}
	if dst.Columns != "3" {
		t.Errorf("Columns: want %q, got %q", "3", dst.Columns)
	}
}

func TestMergeGlobalIntoConfig_DoesNotOverwriteExisting(t *testing.T) {
	dst := &homerconfig.HomerConfig{
		Title:   "Local Title",
		Columns: "4",
	}
	src := &homerconfig.HomerConfig{
		Title:   "Global Title",
		Columns: "3",
		Logo:    "assets/logo.png",
	}
	mergeGlobalIntoConfig(dst, src)

	// Pre-existing values must not be overwritten
	if dst.Title != "Local Title" {
		t.Errorf("Title must not be overwritten; got %q", dst.Title)
	}
	if dst.Columns != "4" {
		t.Errorf("Columns must not be overwritten; got %q", dst.Columns)
	}
	// Empty field should be filled from src
	if dst.Logo != "assets/logo.png" {
		t.Errorf("Logo should be filled from src; got %q", dst.Logo)
	}
}

func TestMergeGlobalIntoConfig_BothEmpty(t *testing.T) {
	dst := &homerconfig.HomerConfig{}
	src := &homerconfig.HomerConfig{}
	mergeGlobalIntoConfig(dst, src) // must not panic
	if dst.Title != "" {
		t.Errorf("expected empty title, got %q", dst.Title)
	}
}

// ── pageName path-traversal validation ───────────────────────────────────────

func TestPageNameValidation_RejectsPathTraversal(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	// Create a page with a malicious pageName. The CRD regex would normally
	// block this, but we verify the controller-level guard too.
	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "bad-page", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "../etc/passwd", // path traversal attempt
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(page).Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "bad-page", Namespace: "default"},
	})
	if err == nil {
		t.Fatal("expected error for path-traversal pageName, got nil")
	}
}

func TestPageNameValidation_RejectsDotInName(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "dotname", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "my.page",
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(page).Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "dotname", Namespace: "default"},
	})
	if err == nil {
		t.Fatal("expected error for pageName containing dot, got nil")
	}
}

// ── HomerPage reconciler: page file writing ───────────────────────────────────

func newHomerPageScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := homerv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}
	return s
}

func TestHomerPageReconcile_WritesPageFile(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	// Stub config.yml so reconcileAutoLink doesn't fail on missing file
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte(""), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "infra", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "infrastructure",
			Groups: []homerv1alpha1.Group{
				{Name: "monitoring", Items: []homerv1alpha1.Item{
					{Name: "Prometheus", Url: "https://prometheus.example.com"},
				}},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(page).
		WithStatusSubresource(page).
		Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "infra", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	// The page file infrastructure.yml must exist and contain the group
	pageFile := filepath.Join(tmpDir, "infrastructure.yml")
	data, err := os.ReadFile(pageFile)
	if err != nil {
		t.Fatalf("page file not created: %v", err)
	}

	var got homerconfig.HomerPageConfig
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to parse page file: %v", err)
	}
	if len(got.Services) == 0 {
		t.Fatal("expected at least one service group in page file")
	}
	if got.Services[0].Name != "monitoring" {
		t.Errorf("expected group 'monitoring', got %q", got.Services[0].Name)
	}
}

func TestHomerPageReconcile_NoRewriteWhenUnchanged(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte(""), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "infra", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "infra",
			Groups:   []homerv1alpha1.Group{{Name: "net"}},
		},
	}

	// Pre-write exactly what the reconciler would produce
	expectedConfig := homerconfig.HomerPageConfig{
		Services: page.Spec.Groups,
	}
	expectedBytes, _ := yaml.Marshal(expectedConfig)
	if err := os.WriteFile(filepath.Join(tmpDir, "infra.yml"), expectedBytes, 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}
	stat1, _ := os.Stat(filepath.Join(tmpDir, "infra.yml"))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(page).
		WithStatusSubresource(page).
		Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "infra", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	stat2, _ := os.Stat(filepath.Join(tmpDir, "infra.yml"))
	if !stat1.ModTime().Equal(stat2.ModTime()) {
		t.Error("page file was rewritten even though content was unchanged")
	}
}

// ── autoLink: adds link in config.yml ────────────────────────────────────────

func TestHomerPageReconcile_AutoLinkAddsLinkToConfig(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	// Start with an empty config.yml
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte("title: Homer\n"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "infra", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "infrastructure",
			AutoLink: true,
			LinkName: "Infrastructure",
			LinkIcon: "fas fa-server",
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(page).
		WithStatusSubresource(page).
		Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "infra", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("config.yml not readable: %v", err)
	}

	var cfg homerconfig.HomerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("failed to parse config.yml: %v", err)
	}

	found := false
	for _, l := range cfg.Links {
		if l.URL == "#infrastructure" {
			found = true
			if l.Name != "Infrastructure" {
				t.Errorf("link name: want 'Infrastructure', got %q", l.Name)
			}
			if l.Icon != "fas fa-server" {
				t.Errorf("link icon: want 'fas fa-server', got %q", l.Icon)
			}
		}
	}
	if !found {
		t.Errorf("expected link '#infrastructure' in config.yml links, got: %+v", cfg.Links)
	}
}

func TestHomerPageReconcile_AutoLinkDefaultsLinkNameToPageName(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte(""), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "p2", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "page2",
			AutoLink: true,
			// LinkName intentionally left empty → should default to "page2"
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(page).
		WithStatusSubresource(page).
		Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "p2", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.yml"))
	var cfg homerconfig.HomerConfig
	_ = yaml.Unmarshal(data, &cfg)

	for _, l := range cfg.Links {
		if l.URL == "#page2" && l.Name != "page2" {
			t.Errorf("expected default link name 'page2', got %q", l.Name)
		}
	}
}

func TestHomerPageReconcile_NoAutoLink_DoesNotModifyConfig(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	initial := "title: Homer\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte(initial), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	page := &homerv1alpha1.HomerPage{
		ObjectMeta: metav1.ObjectMeta{Name: "p3", Namespace: "default"},
		Spec: homerv1alpha1.HomerPageSpec{
			PageName: "page3",
			AutoLink: false,
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(page).
		WithStatusSubresource(page).
		Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, _ = r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "p3", Namespace: "default"},
	})

	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.yml"))
	var cfg homerconfig.HomerConfig
	_ = yaml.Unmarshal(data, &cfg)

	for _, l := range cfg.Links {
		if l.URL == "#page3" {
			t.Errorf("expected no link '#page3' in config when autoLink=false, but found one")
		}
	}
}

// ── HomerPage reconciler: not-found is not an error ───────────────────────────

func TestHomerPageReconcile_NotFoundIsIgnored(t *testing.T) {
	scheme := newHomerPageScheme(t)
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	// Empty fake client — the resource does not exist
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	r := &HomerPageReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "gone", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("expected no error for not-found resource, got: %v", err)
	}
}
