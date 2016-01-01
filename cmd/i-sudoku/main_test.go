package main

import (
	"testing"
)

func TestEnsureSelected(t *testing.T) {
	model := newModel()

	if model.Selected == nil {
		t.Error("New model had no selected cell")
	}

	model.Selected = nil
	model.EnsureSelected()

	if model.Selected == nil {
		t.Error("Model after EnsureSelected still had no selected cell")
	}
}
