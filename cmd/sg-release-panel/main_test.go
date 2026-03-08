package main

import (
	"reflect"
	"testing"
)

func TestNormalizeArgsDefaultsToServe(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "empty", in: nil, want: []string{"serve"}},
		{name: "flags only", in: []string{"--root", ".release-panel"}, want: []string{"serve", "--root", ".release-panel"}},
		{name: "explicit command", in: []string{"status"}, want: []string{"status"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := normalizeArgs(test.in)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("normalizeArgs(%v) = %v, want %v", test.in, got, test.want)
			}
		})
	}
}
