package util_test

import (
	"testing"

	"github.com/akfaiz/migris/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestOptional(t *testing.T) {
	// Test with default value when no values provided
	result := util.Optional("default")
	assert.Equal(t, "default", result)

	// Test with provided value
	result = util.Optional("default", "provided")
	assert.Equal(t, "provided", result)

	// Test with multiple values (should use first)
	result = util.Optional("default", "first", "second")
	assert.Equal(t, "first", result)
}

func TestOptionalPtr(t *testing.T) {
	// Test with default value
	result := util.OptionalPtr("default")
	assert.NotNil(t, result)
	assert.Equal(t, "default", *result)

	// Test with provided value
	result = util.OptionalPtr("default", "provided")
	assert.NotNil(t, result)
	assert.Equal(t, "provided", *result)
}

func TestOptionalNil(t *testing.T) {
	// Test with no values
	result := util.OptionalNil[string]()
	assert.Nil(t, result)

	// Test with provided value
	result = util.OptionalNil("value")
	assert.NotNil(t, result)
	assert.Equal(t, "value", *result)
}

func TestPtrOf(t *testing.T) {
	value := "test"
	result := util.PtrOf(value)
	assert.NotNil(t, result)
	assert.Equal(t, value, *result)
}

func TestTernary(t *testing.T) {
	// Test true condition
	result := util.Ternary(true, "true_val", "false_val")
	assert.Equal(t, "true_val", result)

	// Test false condition
	result = util.Ternary(false, "true_val", "false_val")
	assert.Equal(t, "false_val", result)
}
