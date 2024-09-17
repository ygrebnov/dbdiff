package table

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/sync/errgroup"

	"github.com/ygrebnov/dbdiff/entity/field"
)

type sqliteParser struct{}

func NewSqliteParser() Parser {
	return &sqliteParser{}
}

func (p *sqliteParser) Parse(table Table) error {
	sql := regexp.MustCompile(`(?s)\(.+\)`).FindString(table.GetSchema())
	rawFields := regexp.MustCompile(`(?s)(\S+[^,]+?(,|$))`).FindAllString(sql[1:len(sql)-1], -1)
	if len(rawFields) == 0 {
		return errors.New("cannot parse table schema")
	}

	g := new(errgroup.Group)

	for _, rawField := range rawFields {
		g.Go(func() error {
			rawField = strings.TrimFunc(
				strings.ToLower(rawField),
				func(r rune) bool { return unicode.IsSpace(r) || unicode.IsPunct(r) },
			)
			if strings.HasPrefix(rawField, "create table") ||
				strings.HasPrefix(rawField, "foreign key") ||
				strings.HasPrefix(rawField, ")") ||
				len(rawField) == 0 {
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
