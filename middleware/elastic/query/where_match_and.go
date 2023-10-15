package query

import (
	"context"
	"errors"
	"github.com/flyerxp/lib/logger"
	"go.uber.org/zap"
)

type ConnMatchAnd struct {
	And          []map[string]map[string][]interface{}
	context      context.Context
	isError      bool
	ErrorDetails error
}

func (c *ConnMatchAnd) ContextF(ctx context.Context) {
	c.context = ctx
}
func (c *ConnMatchAnd) Where(ctx context.Context, v *Matchs) error {
	q, e := v.GetSourceMap()
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
		return errors.New("match where error" + e.Error())
	}
}

func (c *ConnMatchAnd) NotWhere(ctx context.Context, v *Matchs) error {
	q, e := v.GetSourceMap()
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
		return errors.New("match where error" + e.Error())
	}
}
func (c *ConnMatchAnd) OrWhereM(and []*ConnMatchAnd) error {
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

func (c *ConnMatchAnd) WhereM(and []*ConnMatchAnd) error {
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

func (c *ConnMatchAnd) NotWhereM(and []*ConnMatchAnd) error {
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

func (c *ConnMatchAnd) GetSourceMap() []map[string]map[string][]interface{} {
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
