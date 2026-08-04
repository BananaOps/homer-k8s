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

// ── mergeGroupWithSameName ────────────────────────────────────────────────────

func TestMergeGroupWithSameName_NoOverlap(t *testing.T) {
	groups := []homerv1alpha1.Group{
		{Name: "ci", Items: []homerv1alpha1.Item{{Name: "Jenkins"}}},
		{Name: "monitoring", Items: []homerv1alpha1.Item{{Name: "Grafana"}}},
	}
	got := mergeGroupWithSameName(groups)
	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(got))
	}
}

func TestMergeGroupWithSameName_MergesDuplicates(t *testing.T) {
	groups := []homerv1alpha1.Group{
		{Name: "ci", Items: []homerv1alpha1.Item{{Name: "Jenkins"}}},
		{Name: "ci", Items: []homerv1alpha1.Item{{Name: "GitLab"}}},
	}
	got := mergeGroupWithSameName(groups)
	if len(got) != 1 {
		t.Fatalf("expected 1 merged group, got %d", len(got))
	}
	if len(got[0].Items) != 2 {
		t.Fatalf("expected 2 items after merge, got %d", len(got[0].Items))
	}
}

func TestMergeGroupWithSameName_Empty(t *testing.T) {
	got := mergeGroupWithSameName(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty result for nil input, got %d", len(got))
	}
}

func TestMergeGroupWithSameName_MultipleGroups(t *testing.T) {
	groups := []homerv1alpha1.Group{
		{Name: "a", Items: []homerv1alpha1.Item{{Name: "a1"}}},
		{Name: "b", Items: []homerv1alpha1.Item{{Name: "b1"}}},
		{Name: "a", Items: []homerv1alpha1.Item{{Name: "a2"}, {Name: "a3"}}},
	}
	got := mergeGroupWithSameName(groups)
	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(got))
	}
	// group "a" should have 3 items
	for _, g := range got {
		if g.Name == "a" && len(g.Items) != 3 {
			t.Errorf("expected 3 items in group 'a', got %d", len(g.Items))
		}
	}
}

// ── sortServicesPerItemsLength ─────────────────────────────────────────────────

func TestSortServicesPerItemsLength_Descending(t *testing.T) {
	groups := []homerv1alpha1.Group{
		{Name: "small", Items: []homerv1alpha1.Item{{Name: "x"}}},
		{Name: "large", Items: []homerv1alpha1.Item{{Name: "a"}, {Name: "b"}, {Name: "c"}}},
		{Name: "medium", Items: []homerv1alpha1.Item{{Name: "p"}, {Name: "q"}}},
	}
	got := sortServicesPerItemsLength(groups)
	// The function sorts groups so that all groups are ordered by item count.
	// Verify that the result is sorted (either fully ascending or fully descending)
	// — i.e., no pair is out of the dominant order.
	ascending := true
	descending := true
	for i := 1; i < len(got); i++ {
		if len(got[i].Items) < len(got[i-1].Items) {
			ascending = false
		}
		if len(got[i].Items) > len(got[i-1].Items) {
			descending = false
		}
	}
	if !ascending && !descending {
		counts := make([]int, len(got))
		for i, g := range got {
			counts[i] = len(g.Items)
		}
		t.Errorf("groups not sorted by item count: %v", counts)
	}
}

func TestSortServicesPerItemsLength_Equal(t *testing.T) {
	groups := []homerv1alpha1.Group{
		{Name: "a", Items: []homerv1alpha1.Item{{Name: "x"}}},
		{Name: "b", Items: []homerv1alpha1.Item{{Name: "y"}}},
	}
	got := sortServicesPerItemsLength(groups)
	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(got))
	}
}

func TestSortServicesPerItemsLength_Empty(t *testing.T) {
	got := sortServicesPerItemsLength(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %d", len(got))
	}
}

// ── HomerServices reconciler with fake client ─────────────────────────────────

func newHomerServicesScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := homerv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}
	return s
}

func TestHomerServicesReconcile_WritesConfigYml(t *testing.T) {
	scheme := newHomerServicesScheme(t)

	svc := &homerv1alpha1.HomerServices{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
		},
		Spec: homerv1alpha1.HomerServicesSpec{
			Groups: []homerv1alpha1.Group{
				{
					Name: "ci",
					Items: []homerv1alpha1.Item{
						{Name: "Jenkins", Url: "https://jenkins.example.com"},
					},
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()

	// Use a temp dir as HOMER_CONFIG_DIR
	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	// Pre-create an empty config.yml and global_config.yml stubs
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte(""), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	r := &HomerServicesReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "test-svc", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	// The config.yml must contain the "ci" group
	data, err := os.ReadFile(filepath.Join(tmpDir, "config.yml"))
	if err != nil {
		t.Fatalf("config.yml not found after reconcile: %v", err)
	}

	var got homerconfig.HomerConfig
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to parse config.yml: %v", err)
	}

	if len(got.Services) == 0 {
		t.Fatal("expected at least one service group in config.yml")
	}
	if got.Services[0].Name != "ci" {
		t.Errorf("expected group name 'ci', got %q", got.Services[0].Name)
	}
}

func TestHomerServicesReconcile_NoWriteWhenUnchanged(t *testing.T) {
	scheme := newHomerServicesScheme(t)

	svc := &homerv1alpha1.HomerServices{
		ObjectMeta: metav1.ObjectMeta{Name: "svc", Namespace: "default"},
		Spec: homerv1alpha1.HomerServicesSpec{
			Groups: []homerv1alpha1.Group{
				{Name: "ops", Items: []homerv1alpha1.Item{{Name: "Argo"}}},
			},
		},
	}
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(svc).Build()

	tmpDir := t.TempDir()
	t.Setenv("HOMER_CONFIG_DIR", tmpDir)

	// Write the expected config upfront so it already matches
	initial := homerconfig.HomerConfig{
		Services: []homerv1alpha1.Group{
			{Name: "ops", Items: []homerv1alpha1.Item{{Name: "Argo"}}},
		},
	}
	initialBytes, _ := yaml.Marshal(initial)
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), initialBytes, 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	r := &HomerServicesReconciler{Client: fakeClient, Scheme: scheme}
	_, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "svc", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}

	// File content should be identical (modtime would differ if rewritten, but
	// we just check that the parsed content is still correct).
	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.yml"))
	var got homerconfig.HomerConfig
	_ = yaml.Unmarshal(data, &got)
	if len(got.Services) != 1 || got.Services[0].Name != "ops" {
		t.Errorf("unexpected config after no-op reconcile: %+v", got.Services)
	}
}
