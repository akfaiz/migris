package schema

import (
	"fmt"
	"slices"
	"strings"

	"github.com/akfaiz/migris/internal/util"
)

type mariadbGrammar struct {
	mysqlGrammar
}

func newMariadbGrammar() *mariadbGrammar {
	g := &mariadbGrammar{
		mysqlGrammar: *newMysqlGrammar(),
	}
	return g
}

func (g *mariadbGrammar) getTypeFuncMap() map[string]func(*columnDefinition) string {
	m := g.mysqlGrammar.getTypeFuncMap()
	m[columnTypeUUID] = g.typeUUID
	m[columnTypeGeography] = g.typeGeography
	m[columnTypeGeometry] = g.typeGeometry
	m[columnTypePoint] = g.typePoint
	return m
}

func (g *mariadbGrammar) getType(col *columnDefinition) string {
	typeFuncMap := g.getTypeFuncMap()
	if fn, ok := typeFuncMap[col.columnType]; ok {
		return fn(col)
	}
	return col.columnType
}

func (g *mariadbGrammar) typeUUID(_ *columnDefinition) string {
	return "UUID"
}

func (g *mariadbGrammar) typeGeometry(col *columnDefinition) string {
	subtype := util.Ternary(col.subtype != nil, util.PtrOf(strings.ToUpper(*col.subtype)), nil)
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

	if col.srid != nil && *col.srid > 0 {
		return fmt.Sprintf("%s REF_SYSTEM_ID=%d", *subtype, *col.srid)
	}

	return *subtype
}

func (g *mariadbGrammar) typeGeography(col *columnDefinition) string {
	return g.typeGeometry(col)
}

func (g *mariadbGrammar) typePoint(col *columnDefinition) string {
	if col.srid != nil && *col.srid > 0 {
		return fmt.Sprintf("POINT REF_SYSTEM_ID=%d", *col.srid)
	}
	return "POINT"
}
