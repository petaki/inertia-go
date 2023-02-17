package inertia

import "testing"

func TestEnableSsr(t *testing.T) {
	i := New("", "", "")
	i.EnableSsr("ssr.test")

	if i.ssrURL != "ssr.test" {
		t.Errorf("expected: ssr.test, got: %v", i.ssrURL)
	}
}

func TestEnableSsrWithDefault(t *testing.T) {
	i := New("", "", "")
	i.EnableSsrWithDefault()

	if i.ssrURL != "http://127.0.0.1:13714" {
		t.Errorf("expected: http://127.0.0.1:13714, got: %v", i.ssrURL)
	}
}

func TestIsSsrEnabled(t *testing.T) {
	i := New("", "", "")

	if i.IsSsrEnabled() {
		t.Errorf("expected: false, got: true")
	}

	i.EnableSsrWithDefault()

	if !i.IsSsrEnabled() {
		t.Errorf("expected: true, got: false")
	}
}
