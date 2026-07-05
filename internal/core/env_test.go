package core

import "testing"

func TestResolveActiveEnv(t *testing.T) {
	e := NewEnvironmentService()
	e.SetEnvironment("dev", map[string]string{"HOST": "dev.example.com", "PORT": "8080"})
	e.SetActive("dev")

	got := e.Resolve("https://${HOST}:${PORT}/api")
	want := "https://dev.example.com:8080/api"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveDefaultAndMissing(t *testing.T) {
	e := NewEnvironmentService()
	if got := e.Resolve("${MISSING:-fallback}"); got != "fallback" {
		t.Fatalf("default: got %q", got)
	}
	if got := e.Resolve("${ALSO_MISSING}"); got != "" {
		t.Fatalf("missing no-default: got %q, want empty", got)
	}
}

func TestResolveEscape(t *testing.T) {
	e := NewEnvironmentService()
	e.SetEnvironment("p", map[string]string{"X": "1"})
	e.SetActive("p")
	if got := e.Resolve(`\${X}`); got != "${X}" {
		t.Fatalf("escape: got %q, want ${X}", got)
	}
}

func TestResolveNested(t *testing.T) {
	e := NewEnvironmentService()
	e.SetEnvironment("p", map[string]string{
		"BASE": "https://${HOST}",
		"HOST": "api.example.com",
	})
	e.SetActive("p")
	if got := e.Resolve("${BASE}/v1"); got != "https://api.example.com/v1" {
		t.Fatalf("nested: got %q", got)
	}
}

func TestMaskSecret(t *testing.T) {
	e := NewEnvironmentService()
	e.SetEnvironment("p", map[string]string{"TOKEN": "supersecret"})
	e.SetActive("p")
	e.MarkSecret("TOKEN")
	resolved := e.Resolve("Authorization: Bearer ${TOKEN}")
	if got := e.Mask(resolved); got != "Authorization: Bearer ••••" {
		t.Fatalf("mask: got %q", got)
	}
}
