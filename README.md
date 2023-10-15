# 整合Lib

简介
===
目前本包支持了nacos,redis,mysql,elastic,pulsar,mqtt 这些服务中间件，另外集成了常用的yaml,json 等工具

pulsar 比kafka更为强大，所以只集成了pulsar,不再集成kafka. 本包的pulsar 采用异步发送,发送耗时可以忽略不不计。

pulsar 第一次创建producer会慢，可以把常用的topic写到pulsar的配置里，提前创建，消息每隔1秒后台发送一次,满100条发送1次,耗时无感知。

app.shutdown() 需要程序推出时调用，会做收尾工作。

elastic 简单易用，不需要很懂es,免去了拼一堆term的烦恼，已经能支持大部分查询场景

案例地址 [案例 洁骏汽车服务有限公司](http://www.ch123.com.cn/ "洁骏汽车服务有限公司")

使用方法
===
* 依赖
  
  nacos 配置中心，配置中心的数据，会在redis生成缓存，更新后，需要清理缓存，参考示例 nacos [监听事件订阅](https://github.com/nacos-group/nacos-sdk-go/blob/master/README_CN.md),订阅到事件后，删除缓存
  
  Lib 对外提供删除key的方法
 
  ```Go
  package main
  /*
  ctx := logger.GetContext(context.Background(), "test")
  client, e := nacos.GetEngine(ctx,"nacosConf")
				if e != nil {
					logger.AddError(zap.Error(e))
				}
				key := client.DeleteCache(context.Background(), dataId, group, syncConf.Ns)
  */
  ```

  redis 缓存
  

* 环境变量

  GO_ENV 作为读取配置文件的目录，例如 值为test ,则读取配置文件 /conf/test/app.yml

  HOSTS:  127.0.0.1 nacosconf nacosconfredis pubmysql pubpulsar pubredis

* 配置文件
    * 默认的配置放在

      config/api.go 里，找不到配置文件，则使用这个配置，便于测试
    * 配置文件
        * [app.yml](https://github.com/flyerxp/lib/blob/main/config/test/conf/test/app.yml)

          综合的app配置
        * pulsar 的topic配置

          pulsar.yml 用来指定哪个topic，放到哪个集群 未指定clusert的，会按照 topic / 1000000 的整数，获取集群代号，按照topic_distribution 指定的配置获取集群

          topicinit.yml 是为了加速第一次producer消息，可以不配置，pulsar 的客户端建立producter 第一次会比较慢，这个配置是为了解决第一加载的问题，没有必要，不用配置

          参考 [middleware\pulsarL\test\conf\test\pulsar.yml](https://github.com/flyerxp/lib/blob/main/middleware/pulsarL/test/conf/test/pulsar.yml)
          
          参考 [topicinit](https://github.com/flyerxp/lib/blob/main/middleware/pulsarL/test/conf/test/topicinit.yml)

工具包
===
* Json工具包使用
  ```Go
  package main
  /*
  import (
        "strings"	     
        "fmt"
        myjson "github.com/flyerxp/lib/utils/json"
     )
     func main(){
       tmp := map[string]string{
          "a":"b",
       }
       r,e:=myjson.Encode(tmp)
       fmt.Println(string(r),e)
  } */
  ```
* Yaml工具包使用

  ```Go
  package main
  /*
  import (
    "fmt"
    myyml "github.com/flyerxp/lib/utils/yaml"
  )  

  func main() {
      var defaultConfig = []byte(`
      a: b
      `)
  tmp := map[string]string{}
  //myyml.DecodeByFile("app.yml", &tmp)
  myyml.DecodeByBytes(defaultConfig, tmp)
  fmt.Println(string(defaultConfig))
  }*/
  ```
* Logger使用

  go.uber.org/zap 使用此lib记录日志

  ```Go
  package main
  /*
  	logger.AddError(zap.Error(errors.New("aaaaaaaa")))
	logger.AddWarn(zap.Error(errors.New("bbbbb")))
	logger.AddNotice(zap.String("a", "bbbbbbbbbbbb"))
	logger.WriteLine()
	//logger.WriteErr()  //立即写入错误
  */
  ```

中间件使用
===
    ```Go
    package main
    /*

    import (
        "context"
        "fmt"
        "github.com/flyerxp/lib/app"
        "github.com/flyerxp/lib/middleware/elastic"
        "github.com/flyerxp/lib/middleware/mysqlL"
        "github.com/flyerxp/lib/middleware/pulsarL"
        "github.com/flyerxp/lib/middleware/redisL"
        "time"
    )
    
    type TestStt struct {
        Id int `json:"id"`
    }
    
    func main() {
        ctx := logger.GetContext(context.Background(), "test")
        //time.Sleep(time.Second * 1)
        defer app.Shutdown(context.Background())
        start := time.Now()
        count := 10000        
        objRedis, _ := redisL.GetEngine(ctx,"pubRedis")
        tmp2 := new(TestStt)
        fmt.Println("github.com/flyerxp")
        fmt.Println("win11 环境，开始了 ")
        for i := 0; i <= count; i++ {
            objRedis.C.Get(ctx, "a")
        }
        fmt.Printf("redis 读取 10000次耗时 %d 毫秒\n", time.Since(start).Milliseconds())
        start = time.Now()
        mysql, _ := mysqlL.GetEngine(ctx,"pubMysql")
        for i := 0; i <= count; i++ {
            err := mysql.GetDb().GetContext(ctx,tmp2, `select id from news_info limit 1`)
            if err != nil {
                fmt.Println(tmp2, err)
            }
        }
        fmt.Printf("mysql 数据库读取 10000次耗时 %d 毫秒\n", time.Since(start).Milliseconds())
        start = time.Now()
        for i := 0; i <= count; i++ {
            pulsarL.Producer(ctx,&pulsarL.OutMessage{
            Topic:      0,
            TopicStr:   "test",
            Content:    "太牛了",
            Properties: map[string]string{"a": "b"},
            Delay:      0,
        })
        }
        fmt.Printf("pulsar 发消息 10000次耗时 %d 毫秒\n", time.Since(start).Milliseconds())
        // es 相关，更多的用法见测试用例
        ec, _ := elastic.GetEngine(ctx,"pubEs")
        e := ec.GetElastic()
        e.SetTable("admin")
        ss := e.GetSearchService(ctx) //调用这个方法的时候,参数要从外面传过来ctx
        ss.Cols([]string{"id", "name", "parent_id", "root_id"})
        ss.WhereIn("id", ss.FieldIntArray([]int{5, 6}))
        p := make([]TestStt, 0)
        ss.Rows(&p)
        fmt.Println(p)
    }
     */

    ```
    ![测试结果](https://github.com/flyerxp/lib/blob/main/doc/image/test.png?raw=true)

  






