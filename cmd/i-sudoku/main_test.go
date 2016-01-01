package main

import (
	"testing"
)

func TestEnsureSelected(t *testing.T) {
	model := newModel()

	if model.selected == nil {
		t.Error("New model had no selected cell")
	}

	model.selected = nil
	model.EnsureSelected()

	if model.selected == nil {
		t.Error("Model after EnsureSelected still had no selected cell")
	}
}
