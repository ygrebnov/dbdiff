package comparer

import (
	"fmt"

	"github.com/ygrebnov/dbdiff/entity/row"
	"github.com/ygrebnov/dbdiff/entity/table"
)

// compareData compares given table data in two databases.
func (c *databaseComparer) compareData(t table.Table) {
	out := c.results.addData(t.GetID())

	equal, comparedPks, err := c.compareDataForD1TablePks(t)
	if err != nil {
		out <- err.Error()
		c.results.output(t, false)
		return
	}

	if err = c.addDataForRemainingD2TablePks(t, comparedPks); err != nil {
		out <- err.Error()
		c.results.output(t, false)
		return
	}

	c.results.output(t, equal)
}

func (c *databaseComparer) compareDataForD1TablePks(t table.Table) (bool, []string, error) {
	rows, errs, err := c.d1.GetTableRowsAll(t)
	if err != nil {
		return false, nil, err
	}

	var comparedPks []string // keeps track of already compared primary key values
	equal := true
	out := c.results.getData(t.GetID())

	for r1 := range rows {
		select {
		case err = <-errs:
			out <- fmt.Sprintf("    %v", err)

		default:
		}

		comparedPks = append(comparedPks, fmt.Sprintf("'%s'", r1.GetPKValue()))

		// get table data row for a given primary key from the second database.
		r2, err := c.d2.GetTableRow(t, r1.GetPKValue())
		if err != nil {
			r2 = row.NewError(t.GetFields()) // fill fields with error message values.
		}

		d, equalLine, err := NewDataRowDifferences(r1, r2, t, c.verbosity > 2)

		switch {
		case err != nil:
			out <- fmt.Sprintf("    %v", err)

		case (!equalLine || c.verbosity > 2) && d != nil:
			out <- fmt.Sprintf("  '%s'='%s':", t.GetPrimaryKey().Name, r1.GetPKValue())
			out <- d.Format(c.verbosity > 1)
		}

		if !equalLine {
			equal = false
		}
	}

	return equal, comparedPks, nil
}

func (c *databaseComparer) addDataForRemainingD2TablePks(t table.Table, comparedPks []string) error {
	rows, errs, err := c.d2.GetTableRowsExcluding(t, comparedPks)
	if err != nil {
		return err
	}

	out := c.results.getData(t.GetID())

	for r2 := range rows {
		select {
		case err = <-errs:
			out <- fmt.Sprintf("    %v", err)
		default:
		}

		d, _, err := NewDataRowDifferences(nil, r2, t, c.verbosity > 2)
		switch {
		case err != nil:
			out <- fmt.Sprintf("    %v", err)

		default:
			out <- fmt.Sprintf("  '%s'='%s':", t.GetPrimaryKey().Name, r2.GetPKValue())
			out <- d.Format(c.verbosity > 1)
		}
	}

	return nil
}
