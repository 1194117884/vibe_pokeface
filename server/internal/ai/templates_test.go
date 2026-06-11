package ai

import "testing"

func TestRenderTemplateReplacesKnownVariables(t *testing.T) {
	got := RenderTemplate("你好${name}，阶段${phase}", map[string]string{
		"name":  "小李",
		"phase": "playing",
	})
	if got != "你好小李，阶段playing" {
		t.Fatalf("rendered = %q", got)
	}
}

func TestRenderTemplateKeepsUnknownVariables(t *testing.T) {
	got := RenderTemplate("${known}-${missing}", map[string]string{"known": "ok"})
	if got != "ok-${missing}" {
		t.Fatalf("rendered = %q", got)
	}
}

func TestTemplateVariablesDeduplicatesVariables(t *testing.T) {
	got := TemplateVariables("${a} ${b} ${a}")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("variables = %v", got)
	}
}
