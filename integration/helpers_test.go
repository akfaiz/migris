package integration_test

import "github.com/akfaiz/migris/schema"

func colNames(cols []*schema.Column) []string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name
	}
	return out
}

func indexNames(idxs []*schema.Index) []string {
	out := make([]string, len(idxs))
	for i, idx := range idxs {
		out[i] = idx.Name
	}
	return out
}
