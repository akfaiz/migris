package grammars

import (
	"fmt"
	"slices"
	"strings"

	"github.com/akfaiz/migris/internal/util"
	"github.com/akfaiz/migris/schema/blueprint"
)

type mariadbGrammar struct {
	mysqlGrammar
}

func newMariadbGrammar() *mariadbGrammar {
	g := &mariadbGrammar{}
	g.mysqlGrammar.serials = []string{
		blueprint.ColumnTypeBigInteger,
		blueprint.ColumnTypeInteger,
		blueprint.ColumnTypeMediumInteger,
		blueprint.ColumnTypeSmallInteger,
		blueprint.ColumnTypeTinyInteger,
	}
	g.mysqlGrammar.self = g
	return g
}

func (g *mariadbGrammar) GetType(col *blueprint.Column) string {
	typeFuncMap := g.getTypeFuncMap()
	if fn, ok := typeFuncMap[col.ColumnType]; ok {
		return fn(col)
	}
	return g.mysqlGrammar.GetType(col)
}

func (g *mariadbGrammar) getTypeFuncMap() map[string]func(*blueprint.Column) string {
	m := g.mysqlGrammar.getTypeFuncMap()
	m[blueprint.ColumnTypeUUID] = g.typeUUID
	m[blueprint.ColumnTypeGeometry] = g.typeGeometry
	m[blueprint.ColumnTypeGeography] = g.typeGeography
	m[blueprint.ColumnTypePoint] = g.typePoint
	return m
}

func (g *mariadbGrammar) typeUUID(_ *blueprint.Column) string {
	return "UUID"
}

func (g *mariadbGrammar) typeGeography(col *blueprint.Column) string {
	return g.typeGeometry(col)
}

func (g *mariadbGrammar) typeGeometry(col *blueprint.Column) string {
	subtype := util.Ternary(col.Subtype != nil, util.PtrOf(strings.ToUpper(*col.Subtype)), nil)
	if subtype != nil {
		if !slices.Contains(
			[]string{
				"POINT",
				"LINESTRING",
				"POLYGON",
				"GEOMETRYCOLLECTION",
				"MULTIPOINT",
				"MULTILINESTRING",
				"MULTIPOLYGON",
			},
			*subtype,
		) {
			subtype = nil
		}
	}

	if subtype == nil {
		subtype = util.PtrOf("GEOMETRY")
	}

	if col.Srid != nil && *col.Srid > 0 {
		return fmt.Sprintf("%s REF_SYSTEM_ID=%d", *subtype, *col.Srid)
	}

	return *subtype
}

func (g *mariadbGrammar) typePoint(col *blueprint.Column) string {
	if col.Srid != nil && *col.Srid > 0 {
		return fmt.Sprintf("POINT REF_SYSTEM_ID=%d", *col.Srid)
	}
	return "POINT"
}
