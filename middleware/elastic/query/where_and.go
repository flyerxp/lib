package query

import (
	"context"
	"errors"
	"github.com/flyerxp/lib/v2/logger"
	"go.uber.org/zap"
)

type ConnAnd struct {
	And          []map[string]map[string][]interface{}
	isError      bool
	ErrorDetails error
}

func (c *ConnAnd) Where(ctx context.Context, f string, v *EsFields) error {
	t := Term{f, v.getData()}
	q, e := t.GetSourceMap()
	if e == nil {
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)
		return nil
	} else {
		logger.AddWarn(ctx, zap.Error(e))
		c.isError = true
		c.ErrorDetails = e
		return errors.New("filter where error" + e.Error())
	}
}

func (c *ConnAnd) WhereGeo(v *EsFields) error {
	q := v.getGeoData()
	qNew := map[string]map[string][]interface{}{
		"bool": {
			"must": []interface{}{
				q,
			},
		},
	}
	c.And = append(c.And, qNew)
	return nil

}
func (c *ConnAnd) NotWhere(ctx context.Context, f string, v *EsFields) error {
	t := Term{f, v}
	q, e := t.GetSourceMap()
	if e == nil {
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must_not": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)

		return nil
	} else {
		logger.AddWarn(ctx, zap.Error(e))
		c.isError = true
		c.ErrorDetails = e
		return errors.New("filter where error" + e.Error())
	}
}
func (c *ConnAnd) WhereIn(ctx context.Context, f string, v *EsFields) error {
	t := Terms{f, v.getArrayData()}
	q, e := t.GetSourceMap()
	if e == nil {
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)

		return nil
	} else {
		logger.AddWarn(ctx, zap.Error(e))
		c.isError = true
		c.ErrorDetails = e
		return errors.New("filter where error" + e.Error())
	}
}
func (c *ConnAnd) NotWhereIn(ctx context.Context, f string, v *EsFields) error {
	t := Terms{f, v.getArrayData()}
	q, e := t.GetSourceMap()
	if e == nil {
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must_not": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)
		return nil
	} else {
		logger.AddWarn(ctx, zap.Error(e))
		c.isError = true
		c.ErrorDetails = e
		return errors.New("filter where error" + e.Error())
	}
}
func (c *ConnAnd) OrWhereM(and []*ConnAnd) error {
	if and == nil {
		return errors.New("and is empty")
	}
	m := map[string]map[string][]interface{}{
		"bool": {},
	}
	for _, w := range and {
		t := w.GetSourceMap()
		if w.isError {
			c.isError = true
			c.ErrorDetails = w.ErrorDetails
			return c.ErrorDetails
		} else {
			for _, tt := range t {
				m["bool"]["should"] = append(m["bool"]["should"], tt)
			}
		}
	}
	c.And = append(c.And, m)
	return nil
}
func (c *ConnAnd) WhereM(and []*ConnAnd) error {
	if and == nil {
		return errors.New("and is empty")
	}
	m := map[string]map[string][]interface{}{
		"bool": {},
	}

	for _, w := range and {
		t := w.GetSourceMap()

		if w.isError {
			c.isError = true
			c.ErrorDetails = w.ErrorDetails
			return w.ErrorDetails
		} else {
			for _, tt := range t {
				m["bool"]["must"] = append(m["bool"]["must"], tt)
			}
		}
	}

	c.And = append(c.And, m)
	return nil
}

func (c *ConnAnd) NotWhereM(and []*ConnAnd) error {
	if and == nil {
		return errors.New("and is empty")
	}
	m := map[string]map[string][]interface{}{
		"bool": {
			"must_not": []interface{}{},
		},
	}
	for _, w := range and {
		t := w.GetSourceMap()
		if w.isError {
			c.isError = true
			c.ErrorDetails = w.ErrorDetails
			return c.ErrorDetails
		} else {
			for _, tt := range t {
				m["bool"]["must_not"] = append(m["bool"]["must_not"], tt)
			}
		}
	}
	c.And = append(c.And, m)
	return nil
}
func (c *ConnAnd) Between(f string, min int, max int) {
	t := Between{f, min, max}
	q, _ := t.GetSourceMap()
	qNew := map[string]map[string][]interface{}{
		"bool": {
			"must": []interface{}{
				q,
			},
		},
	}
	c.And = append(c.And, qNew)
}
func (c *ConnAnd) Compare(f string, compare string, v int) error {
	switch compare {
	case ">":
		o := Gt{f, v}
		q, _ := o.GetSourceMap()
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)

	case ">=":
		o := Gte{f, v}
		q, _ := o.GetSourceMap()
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)
	case "<=":
		o := Lte{f, v}
		q, _ := o.GetSourceMap()
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)
	case "<":
		o := Lt{f, v}
		q, _ := o.GetSourceMap()
		qNew := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{
					q,
				},
			},
		}
		c.And = append(c.And, qNew)
	default:
		c.isError = true
		c.ErrorDetails = errors.New("compare error > < !=  >= <=" + f)
		return errors.New("compare error > < !=  >= <=" + f)
	}
	return nil
}
func (c *ConnAnd) GetSourceMap() []map[string]map[string][]interface{} {

	if c.And == nil || len(c.And) == 0 {
		t := map[string]map[string][]interface{}{
			"bool": {
				"must": []interface{}{},
			},
		}
		return []map[string]map[string][]interface{}{t}
	} else {
		resp := map[string]map[string][]interface{}{
			"bool": {},
		}
		tmpAnd := []interface{}{}
		tmpShould := []interface{}{}
		tmpNot := []interface{}{}
		//语句合并
		for _, oneAnd := range c.And {
			if andW, existsW := oneAnd["bool"]["must"]; existsW {
				for _, t := range andW {
					tmpAnd = append(tmpAnd, t)
				}

				resp["bool"]["must"] = tmpAnd
			}
			if andN, existsN := oneAnd["bool"]["must_not"]; existsN {
				for _, t := range andN {
					tmpNot = append(tmpNot, t)
				}
				resp["bool"]["must_not"] = tmpNot
			}
			if andO, existsO := oneAnd["bool"]["should"]; existsO {
				for _, t := range andO {
					tmpShould = append(tmpShould, t)
				}
				resp["bool"]["should"] = tmpShould
			}
		}

		return []map[string]map[string][]interface{}{resp}
	}
}
