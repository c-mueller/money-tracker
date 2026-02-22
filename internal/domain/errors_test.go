package domain

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError("name", "must not be empty")

	if err.Field != "name" {
		t.Errorf("Field = %q, want %q", err.Field, "name")
	}
	if err.Message != "must not be empty" {
		t.Errorf("Message = %q, want %q", err.Message, "must not be empty")
	}
	if err.Error() != "name: must not be empty" {
		t.Errorf("Error() = %q, want %q", err.Error(), "name: must not be empty")
	}
}

func TestValidationErrorUnwrap(t *testing.T) {
	err := NewValidationError("field", "msg")

	if !errors.Is(err, ErrValidation) {
		t.Error("ValidationError should unwrap to ErrValidation")
	}
	if errors.Is(err, ErrNotFound) {
		t.Error("ValidationError should not match ErrNotFound")
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"not found", ErrNotFound, "not found"},
		{"forbidden", ErrForbidden, "forbidden"},
		{"conflict", ErrConflict, "conflict"},
		{"validation", ErrValidation, "validation error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestSentinelErrorsDistinct(t *testing.T) {
	errs := []error{ErrNotFound, ErrForbidden, ErrConflict, ErrValidation}
	for i, a := range errs {
		for j, b := range errs {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinel errors %q and %q should not match", a, b)
			}
		}
	}
}
