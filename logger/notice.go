package logger

import (
	"context"
	json2 "encoding/json"
	"fmt"
	config2 "github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/utils/json"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// var noticeLog zap.Logger
type AppLog struct {
	ZapLog   *zap.Logger
	once     sync.Once
	isInitEd bool
}
type noticeData struct {
	NoticeMetrics *noticeMetrics
	ExecMetrics   *eTimeMetrics
	MysqlMetrics  *CounterContainer
	Url           string
	Args          string
	Refer         string
	RemoteAddr    string
	LocalAddr     string
	Agent         string
	Expire        time.Time
}

// 中间件耗时
type MiddleExec struct {
	Name          string
	TotalExecTime int
	Count         int
	Max           int
	Avg           int
	ConnectTime   int
	ConnectCount  int
}

type MiddleExecTime struct {
	Redis    MiddleExec
	Mysql    MiddleExec
	Pulsar   MiddleExec
	Kafka    MiddleExec
	MemCache MiddleExec
	Rpc      MiddleExec
	RocketMq MiddleExec
	Elastic  MiddleExec
	Mongo    MiddleExec
	Nacos    MiddleExec
	Mqtt     MiddleExec
}

type ETimeStt struct {
	Start time.Time `json:"start"`
	Exec  int       `json:"exec"`
	Name  string    `json:"name"`
	Step  int       `json:"step"`
}
type eTimeMetrics struct {
	ETime []ETimeStt `json:"eTime"`
}

func (e ETimeStt) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt(e.Name, e.Exec)
	return nil
}
func (e *ETimeStt) Stop(ctx context.Context) {
	e.Exec = int(time.Since(e.Start).Milliseconds())
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.ExecMetrics.ETime = append(n.ExecMetrics.ETime, *e)
	}
	e.Step++
	e.Start = time.Now()
}
func (e *ETimeStt) StopName(ctx context.Context, Name string) {
	e.Exec = int(time.Since(e.Start).Milliseconds())
	oriName := e.Name
	e.Name = fmt.Sprintf("%s_%s", e.Name, Name)
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.ExecMetrics.ETime = append(n.ExecMetrics.ETime, *e)
	}
	e.Name = oriName
	e.Step++
	e.Start = time.Now()
}
func (e ETimeStt) GetExec() int {
	if e.Exec < 0 {
		e.Exec = int(time.Since(e.Start).Milliseconds())
	}
	return e.Exec
}
func (e eTimeMetrics) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for i := range e.ETime {
		//if e.ETime[i].Exec >= 0 {
		if e.ETime[i].Step > 0 {
			enc.AddInt(e.ETime[i].Name+"_"+strconv.Itoa(e.ETime[i].Step), e.ETime[i].Exec)
		} else {
			enc.AddInt(e.ETime[i].Name, e.ETime[i].Exec)
		}
		//}
	}
	return nil
}

// Log数据聚合
type noticeMetrics struct {
	Notice        []zap.Field
	Middle        MiddleExecTime
	TotalExecTime int
}

func (a MiddleExec) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("total", a.TotalExecTime)
	enc.AddInt("count", a.Count)
	enc.AddInt("avg", a.Avg)
	enc.AddInt("max", a.Max)
	if a.Name != "nacos" {
		enc.AddInt("ConnTime", a.ConnectTime)
		enc.AddInt("ConnCount", a.ConnectCount)
	}
	return nil
}

func (r MiddleExecTime) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	//enc.AddString("total", r.)
	if r.Redis.Count > 0 {
		r.Redis.Name = "redis"
		_ = enc.AddObject("redis", r.Redis)
	}
	if r.MemCache.Count > 0 {
		r.MemCache.Name = "memcache"
		_ = enc.AddObject("memCache", r.MemCache)
	}
	if r.Mongo.Count > 0 {
		r.Mongo.Name = "mongo"
		_ = enc.AddObject("mongo", r.Mongo)
	}
	if r.Elastic.Count > 0 {
		r.Elastic.Name = "elastic"
		_ = enc.AddObject("elastic", r.Elastic)
	}
	if r.Kafka.Count > 0 {
		r.Kafka.Name = "kafka"
		_ = enc.AddObject("kafka", r.Kafka)
	}
	if r.Pulsar.Count > 0 {
		r.Pulsar.Name = "pulsar"
		_ = enc.AddObject("pulsar", r.Pulsar)
	}
	if r.Rpc.Count > 0 {
		r.Rpc.Name = "rpc"
		_ = enc.AddObject("rpc", r.Rpc)
	}
	if r.Mysql.Count > 0 || r.Mysql.ConnectCount > 0 {
		r.Mysql.Name = "mysql"
		_ = enc.AddObject("mysql", r.Mysql)
	}
	if r.RocketMq.Count > 0 {
		r.RocketMq.Name = "rocket"
		_ = enc.AddObject("rocket", r.RocketMq)
	}
	if r.Mqtt.Count > 0 {
		r.Mqtt.Name = "mqtt"
		_ = enc.AddObject("mqtt", r.Mqtt)
	}
	if r.Nacos.Count > 0 {
		r.Nacos.Name = "nacos"
		_ = enc.AddObject("nacos", r.Nacos)
	}
	return nil
}

type CounterContainer struct {
	Mysql cmap.ConcurrentMap[string, *Counter]
}

type Counter struct {
	nums int32
	Name string
	Type string
}

func (s *Counter) String() string {
	return s.Name + "_%s_" + strconv.Itoa(int(s.nums))
}
func (s *Counter) GetString(name string) string {
	return fmt.Sprintf("%s_%s_%d", s.Name, name, s.nums)
}
func (s *Counter) Add() {
	atomic.AddInt32(&s.nums, 1)
}
func (s *Counter) Reset() {
	atomic.StoreInt32(&s.nums, 0)
}

var noticeLog = new(AppLog)

func getNoticeLog() {
	noticeLog.once.Do(func() {
		rawJSON, _ := json.Encode(config2.GetConf().App.Logger)
		var cfg zap.Config
		if err := json2.Unmarshal(rawJSON, &cfg); err != nil {
			log.Print(err)
		}
		cfg.OutputPaths = GetPath(config2.GetConf().App.Logger.OutputPaths, "notice")
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		cfg.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		noticeLog.ZapLog = zap.Must(cfg.Build())
		noticeLog.isInitEd = true
		RegistermakeFileEvent(Event{"notice", func() {
			noticeLog = new(AppLog)
			getNoticeLog()
		}})
	})
}

func SetUrl(ctx context.Context, url string) {
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.Url = url
	}
}
func SetArgs(ctx context.Context, args string) {
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.Args = args
	}

}
func SetRefer(ctx context.Context, refer string) {
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.Refer = refer
	}
}
func SetAddr(ctx context.Context, remote string, local string) {
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.RemoteAddr = remote
		n.LocalAddr = local
	}
}
func SetUserAgent(ctx context.Context, agent string) {
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if n, ok := dataContainer.NoticeData.Get(logId); ok {
		n.Agent = agent
	}
}
func addMiddleExecTime(m *MiddleExec, t int) {
	m.Count += 1
	m.TotalExecTime += t
	m.Avg = (int)(m.TotalExecTime / m.Count)
	if t > m.Max {
		m.Max = t
	}
}

func addMiddleConnTime(m *MiddleExec, t int) {
	m.ConnectCount += 1
	m.ConnectTime += t
}
func AddMongoConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Mongo, t)
	}
}
func AddRedisConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Redis, t)
	}
}
func AddPulsarConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Pulsar, t)
	}
}
func AddKafkaConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Kafka, t)
	}
}
func AddEsConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Elastic, t)
	}
}
func AddRpcConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Rpc, t)
	}
}
func AddRocketConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.RocketMq, t)
	}
}
func AddMqttConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Mqtt, t)
	}
}
func AddMysqlConnTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleConnTime(&n.NoticeMetrics.Middle.Mysql, t)
	}
}
func AddMongoTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Mongo, t)
	}
}
func AddRedisTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Redis, t)
	}
}
func AddPulsarTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Pulsar, t)
	}
}
func AddKafkaTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Kafka, t)
	}
}
func AddEsTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Elastic, t)
	}
}
func AddRpcTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Rpc, t)
	}
}
func AddMqttTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Mqtt, t)
	}
}
func AddRocketTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.RocketMq, t)
	}
}
func AddMysqlTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Mysql, t)
	}
}
func AddNacosTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		addMiddleExecTime(&n.NoticeMetrics.Middle.Nacos, t)
	}
}
func StartTime(name string) ETimeStt {
	return ETimeStt{
		Start: time.Now(),
		Exec:  -1,
		Name:  name,
		Step:  0,
	}
}
func SetExecTime(ctx context.Context, t int) {
	if n, ok := dataContainer.NoticeData.Get(GetLogId(ctx)); ok {
		n.NoticeMetrics.TotalExecTime = t
	}
}
