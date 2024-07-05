package gmterrors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := errors.New("test error")
	gcErr := New(err)
	if gcErr.Error() != "gorm-multitenancy: test error" {
		t.Errorf("Expected 'gorm-multitenancy: test error', got '%s'", gcErr.Error())
	}
}

func TestNewWithScheme(t *testing.T) {
	err := errors.New("test error")
	gcErr := NewWithScheme("myscheme", err)
	if gcErr.Error() != "gorm-multitenancy/myscheme: test error" {
		t.Errorf("Expected 'gorm-multitenancy/myscheme: test error', got '%s'", gcErr.Error())
	}
}

func TestUnwrap(t *testing.T) {
	err := errors.New("test error")
	gcErr := New(err)
	unwrappedErr := gcErr.Unwrap()
	if unwrappedErr != err {
		t.Errorf("Expected unwrapped error to be '%v', got '%v'", err, unwrappedErr)
	}
}
