package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetDialect(t *testing.T) {
	testCases := []struct {
		dialect     string
		expectError bool
	}{
		{"postgres", false},
		{"pgx", false},
		{"mysql", false},
		{"mariadb", false},
		{"sqlite", true},
	}
	for _, tc := range testCases {
		t.Run(tc.dialect, func(t *testing.T) {
			err := SetDialect(tc.dialect)
			if tc.expectError {
				assert.Error(t, err, "Expected error for unsupported dialect %s", tc.dialect)
				assert.ErrorIs(t, err, ErrUnknownDialect, "Expected ErrUnknownDialect for unsupported dialect %s", tc.dialect)
			} else {
				assert.NoError(t, err, "Did not expect error for supported dialect %s", tc.dialect)
			}
		})
	}
}
