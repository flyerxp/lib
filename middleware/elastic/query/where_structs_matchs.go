package query

type Matchs struct {
	Field              []string `json:"field"`                //搜索的字段
	Values             string   `json:"value"`                //搜索的词
	Operator           string   `json:"operator"`             //分词之间的关系
	MinimumShouldMatch string   `json:"minimum_should_match"` //默认50%
	Type               string   `json:"type"`                 //类型
	Fuzziness          string   `json:"fuzziness"`            //模糊性
}

func (m *Matchs) AddField(f string) *Matchs {
	m.Field = append(m.Field, f)
	return m
}
func (m *Matchs) GetSourceMap() (interface{}, error) {
	if m.Operator != "or" {
		m.Operator = "and"
	}
	if m.MinimumShouldMatch == "" {
		m.MinimumShouldMatch = "50%"
	}
	if len(m.Field) == 1 {
		r := map[string]map[string]map[string]string{
			"match": {m.Field[0]: {
				"query":    m.Values,
				"operator": m.Operator,
			}},
		}
		if m.Fuzziness != "" {
			r["match"][m.Field[0]]["fuzziness"] = m.Fuzziness
		}
		if m.Type == "prefix" {
			r = map[string]map[string]map[string]string{
				"match_phrase_prefix": {m.Field[0]: {
					"query": m.Values,
				}},
			}
			if m.Fuzziness != "" {
				r["match_phrase_prefix"][m.Field[0]]["fuzziness"] = m.Fuzziness
			}
		}
		if m.Operator == "or" {
			r["match"][m.Field[0]]["minimum_should_match"] = m.MinimumShouldMatch
		}
		return r, nil
	} else {
		r := map[string]map[string]interface{}{
			"multi_match": {
				"query":    m.Values,
				"operator": m.Operator,
				"fields":   m.Field,
			},
		}
		if m.Type == "prefix" {
			r = map[string]map[string]interface{}{
				"multi_match": {
					"query":  m.Values,
					"fields": m.Field,
					"type":   "phrase_prefix",
				},
			}
		}
		if m.Fuzziness != "" {
			r["multi_match"]["fuzziness"] = m.Fuzziness
		}
		if m.Operator == "or" {
			r["multi_match"]["minimum_should_match"] = m.MinimumShouldMatch
		}
		return r, nil
	}
}
