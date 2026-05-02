package grammars_test

import (
	"testing"

	"github.com/akfaiz/migris/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMariadbGrammar_CompileCreate(t *testing.T) {
	g, err := schema.NewGrammar("mariadb")
	require.NoError(t, err)

	tests := []struct {
		name      string
		table     string
		blueprint func(table *schema.Blueprint)
		want      string
		wantErr   bool
	}{
		{
			name:  "basic table creation",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.ID()
				table.String("name", 255)
			},
			want:    "CREATE TABLE `users` (`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `name` VARCHAR(255) NOT NULL, CONSTRAINT `users_id_primary` PRIMARY KEY (`id`))",
			wantErr: false,
		},
		{
			name:  "mariadb specific uuid type",
			table: "users",
			blueprint: func(table *schema.Blueprint) {
				table.UUID("id")
			},
			want:    "CREATE TABLE `users` (`id` UUID NOT NULL)",
			wantErr: false,
		},
		{
			name:  "mariadb specific geometry types",
			table: "locations",
			blueprint: func(table *schema.Blueprint) {
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
			bp := schema.NewBlueprintForTesting(tt.table, g)
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
