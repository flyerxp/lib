package test

import (
	"context"
	"fmt"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/elastic"
	"github.com/flyerxp/lib/middleware/elastic/query"
	"github.com/flyerxp/lib/utils/json"
	"github.com/valyala/fasthttp"
	"io"
	"testing"
)

func TestInfo(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.Where("id", ss.FieldInt(5))
	p := new(Admin)
	err, ok := ss.Find(p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p, ok, ss.Dsl, ss.ErrorDetails)
}
func TestList(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.Cols([]string{"id", "name", "parent_id", "root_id"})
	ss.WhereIn("id", ss.FieldIntArray([]int{5, 6}))
	ss.Between("id", 1, 100)
	ss.Compare("id", "<", 1000)
	ss.AddFieldSort("id", "asc")
	ss.AddFieldSort("parent_id", "asc")
	ss.LimitF(0, 3)
	p := make([]Admin, 0, 3)
	err, ok := ss.Rows(&p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p, ok, ss.Dsl, ss.ErrorDetails)
}
func TestGeo(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.Cols([]string{"id", "name", "parent_id", "root_id"})
	ss.WhereGeo(ss.FieldGeo("location", 1, 0, 0))
	ss.AddGeoSort("location", 18.000, 90, "asc")
	p := make([]Admin, 0, 3)
	err, ok := ss.Rows(&p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p, ok, ss.Dsl, ss.ErrorDetails)
}
func TestBody(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.Cols([]string{"id", "name", "parent_id", "root_id"})
	body, _ := ss.GetQeuryBody()
	ss = e.GetSearchService(ctx)
	//body 可以写原生的body
	result, err := ss.RequestApi(fasthttp.MethodPost, "/admin/_search", body)
	fmt.Println(string(result), err)
}
func TestBatch(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.Cols([]string{"name", "id", "parent_id", "root_id"})
	ss.Compare("id", ">", 1)
	ss.SizeF(2)
	aa := ss.Batch("1m")
	var tmpArray []Admin

	for {
		tmpObj := Admin{}
		e, bb := aa.Next()
		//time.Sleep(time.Second * 3)
		if e == io.EOF {
			fmt.Println("结束")
			break
		} else if e != nil {
			fmt.Println("错误结束", e)
			break
		}
		for i := range bb.Hits.Hits {
			_ = json.Decode(bb.Hits.Hits[i].Source, &tmpObj)
			tmpArray = append(tmpArray, tmpObj)
		}
	}
	fmt.Println(tmpArray)
}

func TestTopHits(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.SizeF(10)
	sort := make([]map[string]map[string]string, 0)
	sort = append(sort, map[string]map[string]string{
		"id": {
			"order": "desc",
		},
	})
	o := &query.AggeSearchOrder{
		Field:      "",
		Alias:      "top_res",
		Cal:        "top_hits",
		SortMethod: "desc",
		Sort:       sort,
		Source:     []string{"id"},
	}
	tt := ss.FieldGroupBy("parent_id", "top_time").AddAggregations(o)
	//tt := ss.CEsGroupBy("uid", "top_time")
	ss.GroupBy(tt)
	ok, aa := ss.GroupRows()
	ttt := new(TopHits)
	err := json.Decode(aa.Aggregations, ttt)
	if ok != nil || err != nil {
		t.Fatalf("expected %v, nofind %v", ss.Dsl, ss.ErrorDetails)
	}
	fmt.Println(ss.ErrorDetails, ss.Dsl, "\n", ttt)
}
func TestCountAgge(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.SizeF(10)
	o := &query.AggeSearchOrder{
		Field:      "id",
		Alias:      "id",
		Cal:        "max",
		SortMethod: "desc",
	}
	tt := ss.FieldGroupBy("parent_id", "parent_id").AddOrder(map[string]string{
		"id": "desc",
	}).AddOrder(map[string]string{
		"_key": "asc",
	})
	tt.AddAggregations(o)
	ss.GroupBy(tt)
	ok, aa := ss.GroupRows()

	ttt := new(AggeOrder)
	println(string(aa.Aggregations))
	err := json.Decode(aa.Aggregations, ttt)
	if ok != nil || err != nil {
		t.Fatalf("expected %v, nofind %v", ss.Dsl, ss.ErrorDetails)
	}
	fmt.Println(ss.ErrorDetails, ss.Dsl, "\n", ttt)
}

func TestWhere(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.SizeF(10)
	// or条件的容器
	qand := make([]*query.ConnAnd, 0)
	w1 := ss.GetNewWhere()
	_ = w1.WhereIn(ctx, "id", ss.FieldIntArray([]int{5, 6, 7}))
	w2 := ss.GetNewWhere()
	_ = w2.WhereIn(ctx, "id", ss.FieldIntArray([]int{8, 9, 10}))
	qand = append(qand, w1, w2)

	//加入更多的的逻辑
	w3 := ss.GetNewWhere()
	_ = w3.Compare("id", ">", 1)
	qand2 := make([]*query.ConnAnd, 0)
	qand2 = append(qand2, w3)
	ss.WhereM(qand2)

	//加上前面的Or
	ss.OrWhereM(qand)
	p := make([]Admin, 0)
	err, st := ss.Rows(&p)
	if err != nil {
		t.Fatalf("expected %v, nofind %v", true, ss.ErrorDetails)
	}
	fmt.Println(ss.ErrorDetails, ss.Dsl, st, "\n", p)

}

// 类似Where
func TestMatch(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.SizeF(10)
	ss.Cols([]string{"id", "name", "url"})

	//复杂的
	m := make([]*query.ConnMatchAnd, 0, 2)
	tq1 := ss.GetNewWhereMatch()
	_ = tq1.Where(ctx, ss.FieldMatchPrefix("name", "管理"))
	tq2 := ss.GetNewWhereMatch()
	_ = tq1.Where(ctx, ss.FieldMatch("name", "管理"))
	m = append(m, tq1, tq2)
	ss.OrWhereMatchM(m)

	//简单的
	//ss.Compare("id", ">", 1)
	ss.WhereMatch(ss.FieldMatch("name", "管理"))
	p := make([]Admin, 0)
	ok, st := ss.Rows(&p)
	if ok != nil {
		t.Fatalf("expected %v, nofind %v", true, ss.ErrorDetails)
	}
	fmt.Println(ss.ErrorDetails, ss.Dsl, st, p)
}
func TestScript(t *testing.T) {
	ctx := logger.GetContext(context.Background(), "test")
	ec, _ := elastic.GetEngine(ctx, "pubEs")
	e := ec.GetElastic()
	e.SetTable("admin")
	ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
	ss.Between("pid", 1, 100)

	str := "int w = 0; " +
		"w=params.a;" +
		"return w;"
	ss.AddScriptSort(str, map[string]interface{}{
		"a": 1,
	}, "desc")
	p := make([]Admin, 0)
	_, _ = ss.Rows(&p)
	fmt.Println(ss.ErrorDetails, ss.Dsl, p)
	logger.WriteLine(ctx)
}
