package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMariadbGrammar_GetType(t *testing.T) {
	g := newMariadbGrammar()

	tests := []struct {
		name       string
		columnType string
		length     *int
		subtype    *string
		srid       *int
		want       string
	}{
		{
			name:       "uuid type",
			columnType: columnTypeUUID,
			want:       "UUID",
		},
		{
			name:       "geometry point",
			columnType: columnTypeGeometry,
			subtype:    ptr("POINT"),
			want:       "POINT",
		},
		{
			name:       "geometry collection with srid",
			columnType: columnTypeGeometry,
			subtype:    ptr("GEOMETRYCOLLECTION"),
			srid:       ptr(4326),
			want:       "GEOMETRYCOLLECTION REF_SYSTEM_ID=4326",
		},
		{
			name:       "point with srid",
			columnType: columnTypePoint,
			srid:       ptr(4326),
			want:       "POINT REF_SYSTEM_ID=4326",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &columnDefinition{
				columnType: tt.columnType,
				length:     tt.length,
				subtype:    tt.subtype,
				srid:       tt.srid,
			}
			assert.Equal(t, tt.want, g.getType(col))
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
