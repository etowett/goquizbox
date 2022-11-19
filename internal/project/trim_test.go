package project

import (
	"testing"
)

func TestTrimSpaceAndNonPrintable_unicode(t *testing.T) {
	t.Parallel()

	extraChars := "state\uFEFF"
	want := "state"
	got := TrimSpaceAndNonPrintable(extraChars)

	if want != got {
		t.Fatalf("wrong trim, want: %q got: %q", want, got)
	}
}

func TestTrimSpaceAndNonPrintable_space(t *testing.T) {
	t.Parallel()

	extraChars := " state  \r\t"
	want := "state"
	got := TrimSpaceAndNonPrintable(extraChars)

	if want != got {
		t.Fatalf("wrong trim, want: %q got: %q", want, got)
	}
}

func TestTrimSpace_unicode(t *testing.T) {
	t.Parallel()

	extraChars := "state\uFEFF"
	want := "state"
	got := TrimSpace(extraChars)

	if want != got {
		t.Fatalf("wrong trim, want: %q got: %q", want, got)
	}
}

func TestTrimSpace_space(t *testing.T) {
	t.Parallel()

	extraChars := " state  \r\t"
	want := "state"
	got := TrimSpace(extraChars)

	if want != got {
		t.Fatalf("wrong trim, want: %q got: %q", want, got)
	}
}
