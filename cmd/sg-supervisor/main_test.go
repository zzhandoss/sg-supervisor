package main

import (
	"reflect"
	"testing"
)

func TestNormalizeArgsDefaultsToServe(t *testing.T) {
	got := normalizeArgs(nil)
	want := []string{"serve"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs(nil) = %v, want %v", got, want)
	}
}

func TestNormalizeArgsPrependsServeForFlags(t *testing.T) {
	got := normalizeArgs([]string{"--root", "."})
	want := []string{"serve", "--root", "."}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs(flags) = %v, want %v", got, want)
	}
}

func TestNormalizeArgsKeepsExplicitCommand(t *testing.T) {
	got := normalizeArgs([]string{"status", "--root", "."})
	want := []string{"status", "--root", "."}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs(command) = %v, want %v", got, want)
	}
}
