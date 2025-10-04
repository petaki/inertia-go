package inertia

import (
	"html/template"
	"testing"
)

func TestMarshal(t *testing.T) {
	app := struct {
		Name   string `json:"name"`
		Locale string `json:"locale"`
	}{
		Name:   "test",
		Locale: "en",
	}

	expected := template.JS(`{"name":"test","locale":"en"}`)
	got, _ := marshal(app)

	if got != expected {
		t.Errorf("expected value: %s, got: %s", expected, got)
	}
}

func TestRaw(t *testing.T) {
	block := []string{
		"<h1>Hello</h1>",
		"<p>From Inertia-Go-Test</p>",
	}

	expected := template.HTML("<h1>Hello</h1>\n<p>From Inertia-Go-Test</p>")
	got, _ := raw(block)

	if got != expected {
		t.Errorf("expected value: %s, got: %s", expected, got)
	}

	line := "Inertia-Go-Test<br>"

	expected = "Inertia-Go-Test<br>"
	got, _ = raw(line)

	if got != expected {
		t.Errorf("expected value: %s, got: %s", expected, got)
	}
}
