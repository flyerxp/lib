package mysqlL

import "database/sql"

type Rows struct {
	ReturnError bool
}

func (r *Rows) GetRowMapKInt(rows *sql.Rows) (map[int]int, error) {
	var id, c int
	var e error
	m := map[int]int{}
	for rows.Next() {
		e = rows.Scan(&id, &c)
		if e != nil {
			if r.ReturnError {
				return m, e
			} else {
				panic(e)
			}
		}
		m[id] = c
	}
	return m, nil
}

func (r *Rows) GetRowMapKString(rows *sql.Rows) (map[string]int, error) {
	var id string
	var c int
	var e error
	m := map[string]int{}
	for rows.Next() {
		e = rows.Scan(&id, &c)
		if e != nil {
			if r.ReturnError {
				return m, e
			} else {
				panic(e)
			}
		}
	}
	return m, nil
}
