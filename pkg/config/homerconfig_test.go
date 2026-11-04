package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestHomerConfigKeepsUnknownKeys(t *testing.T) {
	in := []byte(`title: Dashboard
stylesheet:
  - assets/custom.css
theme: default
defaults:
  colorTheme: dark
`)
	var c HomerConfig
	if err := yaml.Unmarshal(in, &c); err != nil {
		t.Fatal(err)
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := yaml.Unmarshal(out, &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"stylesheet", "theme", "defaults"} {
		if _, ok := got[k]; !ok {
			t.Errorf("key %q dropped after round trip:\n%s", k, out)
		}
	}
	if got["title"] != "Dashboard" {
		t.Errorf("title = %v", got["title"])
	}
}
