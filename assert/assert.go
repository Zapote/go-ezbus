package assert

import (
	"reflect"
	"testing"
)

// IsEqual checks if values are equal actual value (a) and expected value(b)
func IsEqual(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		return
	}
	t.Errorf("Received '%v' (type %v), expected '%v' (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}

// IsTrue checks if value is true
func IsTrue(t *testing.T, a bool, message string) {
	if a {
		return
	}

	t.Errorf(message)
}

// IsFalse checks if value is false
func IsFalse(t *testing.T, a bool, message string) {
	if a {
		t.Errorf(message)
	}
}

// IsNil checks if value is nil
func IsNil(t *testing.T, v interface{}) {
	if v != nil {
		t.Errorf("%s should be nil", v)
	}
}

// IsNil checks if value not is nil
func IsNotNil(t *testing.T, v interface{}) {
	if v == nil {
		t.Errorf("Value should not be nil")
	}
}
