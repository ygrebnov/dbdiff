package table

import (
	"strings"
	"unicode"

	"golang.org/x/sync/errgroup"

	"github.com/ygrebnov/dbdiff/entity/field"
)

type postgresqlParser struct{}

func NewPostgresqlParser() Parser {
	return &postgresqlParser{}
}

func (p *postgresqlParser) Parse(table Table) error {
	g := new(errgroup.Group)

	for _, rawField := range strings.Split(table.GetSchema(), ",") {
		g.Go(func() error {
			rawField = strings.TrimFunc(
				strings.ToLower(rawField),
				func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) },
			)
			if len(rawField) == 0 {
				return nil
			}

			f, err := field.New(rawField)
			if err == nil {
				table.AddField(f)
			}

			return err
		})
	}

	return g.Wait()
}
