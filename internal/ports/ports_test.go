package ports

import (
	"reflect"
	"testing"
)

func TestRangesSupport(t *testing.T) {
	got, err := parseList("22,80,8001,8002,8000-8003")
	if err != nil {
		t.Fatalf("returned error: %v", err)
	}

	want := []int{22, 80, 8000, 8001, 8002, 8003}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mismatch: got %v want %v", got, want)
	}
}

func TestRejectInvalidRanges(t *testing.T) {
	_, err := parseList("8001-8000,1-,a-b")
	if err == nil {
		t.Fatalf("expected parseList error")
	}

	want := "invalid ports: 8001-8000,1-,a-b"
	if err.Error() != want {
		t.Fatalf("unexpected error: got %q want %q", err.Error(), want)
	}
}

func TestCustomPreservesRanges(t *testing.T) {
	got, err := NormalizeCustom("82-84,80,83,80, 82 - 84 ,9000-9000")
	if err != nil {
		t.Fatalf("returned error: %v", err)
	}

	want := "80,82-84,9000"
	if got != want {
		t.Fatalf("mismatch: got %q want %q", got, want)
	}
}

func TestCustomOverlappingRanges(t *testing.T) {
	got, err := NormalizeCustom("1-100,50-70,150-200,180-220")
	if err != nil {
		t.Fatalf("returned error: %v", err)
	}

	want := "1-100,150-220"
	if got != want {
		t.Fatalf("mismatch: got %q want %q", got, want)
	}
}
