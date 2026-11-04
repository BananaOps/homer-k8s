package v1alpha1

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddToSchemeRegistersAllKinds(t *testing.T) {
	s := runtime.NewScheme()
	if err := AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"HomerServices", "HomerServicesList", "HomerPage", "HomerPageList"} {
		if !s.Recognizes(GroupVersion.WithKind(kind)) {
			t.Errorf("kind %s not registered in scheme", kind)
		}
	}
}

// Homer reads camelCase keys (keywords, apiVersion, checkInterval...). The
// controllers marshal Group/Item with yaml.v3, so the yaml tags must match the
// json ones, and empty fields must be omitted.
func TestItemYAMLKeysMatchHomer(t *testing.T) {
	out, err := yaml.Marshal(Group{Name: "g", Items: []Item{{
		Name:          "svc",
		Keyword:       "k",
		ApiVersion:    "v2",
		CheckInterval: 5,
		LibraryType:   "movies",
		LocationId:    "42",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, key := range []string{"keywords: k", "apiVersion: v2", "checkInterval: 5", "libraryType: movies", "locationId: \"42\""} {
		if !strings.Contains(got, key) {
			t.Errorf("expected %q in:\n%s", key, got)
		}
	}
	for _, key := range []string{"keyword:", "apiversion:", "icon:", "target:", "limit:"} {
		if strings.Contains(got, key) {
			t.Errorf("unexpected %q in:\n%s", key, got)
		}
	}

	var back Group
	if err := yaml.Unmarshal(out, &back); err != nil {
		t.Fatal(err)
	}
	if back.Items[0].ApiVersion != "v2" || back.Items[0].Keyword != "k" {
		t.Errorf("round-trip lost fields: %+v", back.Items[0])
	}
}

func TestItemYAMLTagsMirrorJSONTags(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeFor[Group](), reflect.TypeFor[Item]()} {
		for i := range typ.NumField() {
			f := typ.Field(i)
			if j, y := f.Tag.Get("json"), f.Tag.Get("yaml"); j != y {
				t.Errorf("%s.%s: json tag %q != yaml tag %q", typ.Name(), f.Name, j, y)
			}
		}
	}
}

func TestDeepCopyIsIndependent(t *testing.T) {
	orig := &HomerPage{Spec: HomerPageSpec{
		PageName: "p",
		Groups:   []Group{{Name: "g", Items: []Item{{Name: "a", Environments: []string{"prod"}}}}},
	}}
	cp := orig.DeepCopyObject().(*HomerPage)
	cp.Spec.Groups[0].Items[0].Environments[0] = "dev"
	cp.Spec.Groups[0].Name = "changed"
	if orig.Spec.Groups[0].Name != "g" || orig.Spec.Groups[0].Items[0].Environments[0] != "prod" {
		t.Fatal("DeepCopy shares memory with the original")
	}

	svc := &HomerServicesList{Items: []HomerServices{{Spec: HomerServicesSpec{Groups: []Group{{Name: "g"}}}}}}
	svcCp := svc.DeepCopyObject().(*HomerServicesList)
	svcCp.Items[0].Spec.Groups[0].Name = "x"
	if svc.Items[0].Spec.Groups[0].Name != "g" {
		t.Fatal("HomerServicesList DeepCopy shares memory")
	}
}
