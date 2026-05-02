package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMariadbGrammar_CompileCreate(t *testing.T) {
	g := newMariadbGrammar()

	tests := []struct {
		name      string
		table     string
		blueprint func(table *Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic table creation",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.ID()
				table.String("name", 255)
			},
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `name` VARCHAR(255) NOT NULL, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`))",
			wantErr: false,
		},
		{
			name:  "mariadb specific uuid type",
			table: "users",
			blueprint: func(table *Blueprint) {
				table.UUID("id")
			},
			want:    "CREATE TABLE `users` (`id` UUID NOT NULL)",
			wantErr: false,
		},
		{
			name:  "mariadb specific geometry types",
			table: "locations",
			blueprint: func(table *Blueprint) {
				table.Geometry("geom", "POINT")
				table.Point("pt")
				table.Geography("geog", "LINESTRING", 4326)
			},
			want:    "CREATE TABLE `locations` (`geom` POINT NOT NULL, `pt` POINT REF_SYSTEM_ID=4326 NOT NULL, `geog` LINESTRING REF_SYSTEM_ID=4326 NOT NULL)",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &Blueprint{name: tt.table, grammar: g}
			tt.blueprint(bp)
			got, err := g.CompileCreate(bp)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

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
