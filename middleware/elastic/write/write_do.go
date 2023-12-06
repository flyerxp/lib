package write

import (
	"context"
	"errors"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/elastic/tool"
	fasthttp "github.com/valyala/fasthttp"
	"time"
)

type WriteResult struct {
	Took   int  `json:"took"`
	Errors bool `json:"errors"`
	Items  []struct {
		Update struct {
			Index   string `json:"_index"`
			Type    string `json:"_type"`
			Id      string `json:"_id"`
			Version int    `json:"_version"`
			Result  string `json:"result"`
			Shards  struct {
				Total      int `json:"total"`
				Successful int `json:"successful"`
				Failed     int `json:"failed"`
			} `json:"_shards"`
			SeqNo       int `json:"_seq_no"`
			PrimaryTerm int `json:"_primary_term"`
			Status      int `json:"status"`
		} `json:"update"`
	} `json:"items"`
}
type writeDo struct {
	context      context.Context
	isError      bool
	ErrorDetails error
}

func (sd *writeDo) insert(s *WriteService, write []*Write) (*WriteResult, error) {
	return sd.WriteData(s, "index", write)
}
func (sd *writeDo) update(s *WriteService, write []*Write) (*WriteResult, error) {
	return sd.WriteData(s, "update", write)
}
func (sd *writeDo) WriteData(s *WriteService, active string, write []*Write) (*WriteResult, error) {
	decode := tool.DefaultDecoder{}
	var head map[string]map[string]interface{}
	var e error
	var tmp []byte
	body := ""
	for _, v := range write {
		if v.Index == "" {
			s.setError(errors.New("index not find"))
			return new(WriteResult), e
		}
		if v.PrimaryInt == 0 && v.PrimaryString == "" {
			s.setError(errors.New("primaryInt not find"))
			return new(WriteResult), e
		}
		head = map[string]map[string]interface{}{
			active: {},
		}
		if v.PrimaryInt >= 0 {
			head[active]["_id"] = v.PrimaryInt
		} else {
			head[active]["_id"] = v.PrimaryString
		}
		if v.Routing != "" {
			head[active]["routing"] = v.Routing
		}
		head[active]["_index"] = v.Index
		tmp, e = decode.EnCode(head)

		if e != nil {
			s.setError(e)
			return new(WriteResult), e
		}
		body += string(tmp) + "\n"
		if active == "index" {
			tmp, e = decode.EnCode(v.Data)
		} else {
			tmp, e = decode.EnCode(map[string]interface{}{"doc": v.Data})
		}
		if e != nil {
			s.setError(e)
			return new(WriteResult), e
		}
		body += string(tmp) + "\n"
	}
	b, e := sd.Do(s, body)
	r := new(WriteResult)
	e = decode.Decode(b, r)
	return r, e
}
func (sd *writeDo) Do(s *WriteService, body string) ([]byte, error) {
	s.Dsl = body
	start := time.Now()
	r, e := s.httpClient.SendRequest(s.context, fasthttp.MethodPost, sd.getUrl(s), body, s.timeOut, 0, true)
	tTime := time.Since(start).Microseconds()
	logger.AddEsTime(s.context, int(tTime))
	return r, e
}

func (sd *writeDo) ContextF(ctx context.Context) {
	sd.context = ctx
}
func (sd *writeDo) getUrl(s *WriteService) string {
	if s.table.Name == "" {
		sd.setError(s, errors.New("unkow table name"))
		return "/"
	}
	url := "/" + s.table.Name + "/_bulk"
	/*qs := make([]string, 0)
	if len(qs) > 0 {
		url += "?" + strings.Join(qs, "&")
	}*/
	return url
}
func (sd *writeDo) setError(s *WriteService, e error) {
	sd.isError = true
	sd.ErrorDetails = e
	s.setError(e)
}
