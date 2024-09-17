package query

import (
	"fmt"
	"strings"
)

type postgresql struct{}

func NewPostgresqlQuery() Schema {
	return &postgresql{}
}

func (*postgresql) All() string {
	return `SELECT 
    table_name AS name, 
    string_agg(column_name || ' ' || data_type || ' ' || COALESCE(constraint_type, ''), ', ' order by column_name) AS sql
FROM (
	SELECT s.table_name, s.column_name, s.data_type, p.constraint_type
	FROM information_schema.columns s
	LEFT JOIN (
		SELECT c.table_name, c.column_name, tc.constraint_type
		FROM information_schema.table_constraints tc 
		JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name) 
		JOIN information_schema.columns AS c ON c.table_schema = tc.constraint_schema
		AND tc.table_name = c.table_name AND ccu.column_name = c.column_name
		WHERE tc.constraint_type = 'PRIMARY KEY' and c.table_schema='public'
	) p
	ON s.table_name = p.table_name AND s.column_name = p.column_name
	WHERE table_schema = 'public') comb
GROUP BY table_name;`
}

func (*postgresql) One(name string) string {
	return fmt.Sprintf(`SELECT 
	    STRING_AGG(
			column_name || ' ' || data_type || ' ' || COALESCE(constraint_type, ''), 
			', ' order by column_name
		) AS sql
	FROM (
		SELECT s.column_name, s.data_type, p.constraint_type
		FROM information_schema.columns s
		LEFT JOIN (
			SELECT c.column_name, tc.constraint_type
			FROM information_schema.table_constraints tc 
			JOIN information_schema.constraint_column_usage AS ccu USING (constraint_schema, constraint_name) 
			JOIN information_schema.columns AS c ON c.table_schema = tc.constraint_schema
			AND tc.table_name = c.table_name AND ccu.column_name = c.column_name
			WHERE tc.constraint_type = 'PRIMARY KEY' and c.table_schema='public' and c.table_name='%s'
		) p
		ON s.column_name = p.column_name
		WHERE table_schema = 'public' AND table_name = '%s') comb;`, name, name)
}

func (*postgresql) Excluded(names []string) string {
	var suffix string
	if len(names) > 0 {
		suffix = fmt.Sprintf(" AND table_name NOT IN (%s)", strings.Join(names, ","))
	}
	return fmt.Sprintf(`SELECT DISTINCT(table_name) AS name 
	FROM information_schema.columns 
	WHERE table_schema = 'public'%s;`, suffix)
}

var Postgresql = NewPostgresqlQuery()
