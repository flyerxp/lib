package logger

import (
	"context"
	"fmt"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	dataContainer = new(DataContainer)

	go func() {
		getNoticeLog()
		getErrorLog()
		getWarnLog()
	}()
}

// 快速获取一个带logid上下文的context
func GetContext(ctx context.Context, logId string) context.Context {
	//第二天重置
	if CurrDay != time.Now().Day() {
		makeFileEventNew := make([]Event, 0, 15)
		for _, v := range makeFileEvent {
			makeFileEventNew = append(makeFileEventNew, v)
		}
		makeFileEvent = make([]Event, 0, 15)
		for _, f := range makeFileEventNew {
			f.F()
		}
		CurrDay = time.Now().Day()
	}
	ctx = context.WithValue(ctx, GetLogIdKey(), logId)
	RegisterNewLog(ctx)
	return ctx
}
func Sync() {
	errLogV.ZapLog.Sync()
	warnLogV.ZapLog.Sync()
	noticeLog.ZapLog.Sync()
	if AccessLogV != nil {
		AccessLogV.Sync()
	}
}

// 初始化的时候使用
func RegisterNewLog(ctx context.Context) {
	id := GetLogId(ctx)
	if id == "" {
		panic("ctx no find " + GetLogIdKey() + " key")
	}
	if !dataContainer.isInitEnd {
		dataContainer.NoticeData = cmap.New[*noticeData]()
		dataContainer.ErrData = cmap.New[*errMetrics]()
		dataContainer.WarnData = cmap.New[*warMetrics]()
		dataContainer.Clear = cmap.New[*time.Timer]()
		dataContainer.isInitEnd = true
	}
	ndata := &noticeMetrics{}
	edata := &errMetrics{}
	wdata := &warMetrics{}
	ndata.Middle = MiddleExecTime{}
	ndata.Notice = make([]zap.Field, 1, 10)
	ndata.Notice[0] = zap.Namespace("notice")
	ndata.TotalExecTime = 0
	counter := new(CounterContainer)
	counter.Mysql = cmap.New[*Counter]()
	dataContainer.NoticeData.Set(id, &noticeData{
		NoticeMetrics: ndata,
		MysqlMetrics:  counter,
		ExecMetrics:   new(eTimeMetrics),
		Expire:        time.Now()})
	edata.Error = make([]zap.Field, 1, 10)
	edata.Expire = time.Now()
	edata.Error[0] = zap.Namespace("error")
	wdata.Warn = make([]zap.Field, 1, 10)
	wdata.Warn[0] = zap.Namespace("warn")
	wdata.Expire = time.Now()
	dataContainer.ErrData.Set(id, edata)
	dataContainer.WarnData.Set(id, wdata)
	dataContainer.Clear.Set(id, time.AfterFunc(time.Second*30, func() {
		dataContainer.NoticeData.Remove(id)
		dataContainer.WarnData.Remove(id)
		dataContainer.ErrData.Remove(id)
	}))
}
func GetLogIdKey() string {
	return "loggerId"
}
func AddNotice(ctx context.Context, field ...zap.Field) {
	if !noticeLog.isInitEd {
		getNoticeLog()
	}
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if nData, ok := dataContainer.NoticeData.Get(logId); ok {
		nData.NoticeMetrics.Notice = append(nData.NoticeMetrics.Notice, field...)
	}
}
func AddError(ctx context.Context, field ...zap.Field) {
	if !errLogV.isInitEd {
		getErrorLog()
	}
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if eData, ok := dataContainer.ErrData.Get(logId); ok {
		eData.Error = append(eData.Error, field...)
	}
}
func AddWarn(ctx context.Context, field ...zap.Field) {
	if !warnLogV.isInitEd {
		getWarnLog()
	}
	logId := GetLogId(ctx)
	if logId == "" {
		return
	}
	if wData, ok := dataContainer.WarnData.Get(logId); ok {
		wData.Warn = append(wData.Warn, field...)
	}
}

var CurrDay = time.Now().Day()

type Event struct {
	Name string
	F    func()
}

var makeFileEvent = make([]Event, 0, 15)

// 生成文件的时候执行
func RegistermakeFileEvent(F Event) {
	makeFileEvent = append(makeFileEvent, F)
}

func GetPath(paths []string, vtype string) []string {
	pathNew := make([]string, 0, len(paths))
	for i := range paths {
		if strings.Contains(paths[i], "/") {
			pathTmp := filepath.Dir(paths[i])
			_, err := os.Stat(pathTmp)
			if os.IsNotExist(err) {
				_ = os.MkdirAll(pathTmp, os.ModePerm)
			}
			if prefix == "" {
				pathNew = append(pathNew, paths[i]+fmt.Sprintf("_%s_%02d_%02d.log", vtype, time.Now().Month(), time.Now().Day()))
			} else {
				pathNew = append(pathNew, paths[i]+fmt.Sprintf("_%s_%s_%02d_%02d.log", prefix, vtype, time.Now().Month(), time.Now().Day()))
			}
		} else {
			pathNew = append(pathNew, paths[i])
		}
	}
	return pathNew
}

func WriteLine(ctx context.Context) {
	if !noticeLog.isInitEd {
		getNoticeLog()
	}
	logId := ctx.Value(GetLogIdKey()).(string)
	if logId == "" {
		return
	}
	nData, ok := dataContainer.NoticeData.Get(logId)
	if !ok {
		return
	}
	wData, _ := dataContainer.WarnData.Get(logId)
	eData, _ := dataContainer.ErrData.Get(logId)
	tData, _ := dataContainer.Clear.Get(logId)
	noticeLog.ZapLog.With(
		zap.String("url", nData.Url),
		zap.String("args", nData.Args),
		zap.String("logid", logId),
		zap.String("referer", nData.Refer),
		zap.String("remoteIp", nData.RemoteAddr),
		zap.String("serverIp", nData.LocalAddr),
		zap.String("user-agent", nData.Agent),
		zap.Object("middle", nData.NoticeMetrics.Middle),
		zap.Object("execTime", nData.ExecMetrics),
		zap.Float32("execTotalTime", nData.NoticeMetrics.TotalExecTime),
	).With(nData.NoticeMetrics.Notice...).Info("info")
	if len(eData.Error) > 1 {
		if !errLogV.isInitEd {
			getErrorLog()
		}
		errLogV.ZapLog.With(zap.String("url", nData.Url),
			zap.String("args", nData.Args),
			zap.String("logid", logId),
			zap.Object("middle", nData.NoticeMetrics.Middle),
			zap.Object("execTime", nData.ExecMetrics),
			zap.Float32("execTotalTime", nData.NoticeMetrics.TotalExecTime),
		).With(eData.Error...).WithOptions(zap.AddCallerSkip(2)).Error("error")
	}
	if len(wData.Warn) > 1 {
		if !warnLogV.isInitEd {
			getWarnLog()
		}
		warnLogV.ZapLog.With(
			zap.String("args", nData.Args),
			zap.String("logid", logId),
			zap.Object("middle", nData.NoticeMetrics.Middle),
			zap.Object("execTime", nData.ExecMetrics),
			zap.Float32("execTotalTime", nData.NoticeMetrics.TotalExecTime)).
			With(wData.Warn...).WithOptions(zap.AddCallerSkip(2)).Warn("warn")
	}
	tData.Stop()
	dataContainer.NoticeData.Remove(logId)
	dataContainer.WarnData.Remove(logId)
	dataContainer.ErrData.Remove(logId)
	dataContainer.Clear.Remove(logId)
	//第二天重置
	if CurrDay != time.Now().Day() {
		makeFileEventNew := make([]Event, 0, 15)
		for _, v := range makeFileEvent {
			makeFileEventNew = append(makeFileEventNew, v)
		}
		makeFileEvent = make([]Event, 0, 15)
		for _, f := range makeFileEventNew {
			f.F()
		}
		CurrDay = time.Now().Day()
	}
}
func WriteErr(ctx context.Context) {
	logId := GetLogId(ctx)
	eData, _ := dataContainer.ErrData.Get(logId)
	nData, ok := dataContainer.NoticeData.Get(logId)
	wData, _ := dataContainer.WarnData.Get(logId)
	if !ok {
		return
	}
	if len(eData.Error) > 1 {
		if !errLogV.isInitEd {
			getErrorLog()
		}
		errLogV.ZapLog.With(zap.String("url", nData.Url),
			zap.String("args", nData.Args),
			zap.String("logid", logId),
			zap.Object("middle", nData.NoticeMetrics.Middle),
			zap.Object("execTime", nData.ExecMetrics),
			zap.Float32("execTotalTime", nData.NoticeMetrics.TotalExecTime),
		).With(eData.Error...).WithOptions(zap.AddCallerSkip(2)).Error("error")
		eData.Error = make([]zap.Field, 1, 10)
		eData.Error[0] = zap.Namespace("error")
	}
	if len(wData.Warn) > 1 {
		if !warnLogV.isInitEd {
			getWarnLog()
		}
		warnLogV.ZapLog.With(
			zap.String("args", nData.Args),
			zap.String("logid", logId),
			zap.Object("middle", nData.NoticeMetrics.Middle),
			zap.Object("execTime", nData.ExecMetrics),
			zap.Float32("execTotalTime", nData.NoticeMetrics.TotalExecTime)).
			With(wData.Warn...).WithOptions(zap.AddCallerSkip(1)).Warn("warn")
		wData.Warn = make([]zap.Field, 1, 10)
		wData.Warn[0] = zap.Namespace("warn")
	}
}
func ErrWithoutCtx(field ...zap.Field) {
	errLogV.ZapLog.With(field...).WithOptions(zap.AddCallerSkip(2)).Error("error")
}
func WarnWithoutCtx(field ...zap.Field) {
	errLogV.ZapLog.With(field...).WithOptions(zap.AddCallerSkip(2)).Warn("warn")
}
func RegisterMysqlCounter(ctx context.Context, name string) *Counter {
	logId := GetLogId(ctx)
	if logId == "" {
		return new(Counter)
	}
	nData, ok := dataContainer.NoticeData.Get(logId)
	if !ok {
		return new(Counter)
	}
	C, ok := nData.MysqlMetrics.Mysql.Get(name)
	if ok {
		return C
	} else {
		co := new(Counter)
		co.Name = name
		co.Type = "mysql"
		nData.MysqlMetrics.Mysql.Set(name, co)
		return co
	}
}
func GetLogId(ctx context.Context) string {
	if ctx.Value(GetLogIdKey()) == nil {
		return ""
	}
	return ctx.Value(GetLogIdKey()).(string)
}

// 一般由consumer使用
var prefix = ""

func SetLogPathPrefix(p string) {
	prefix = p
	go func() {
		noticeLog = new(AppLog)
		getNoticeLog()
		errLogV = new(errLog)
		getErrorLog()
		warnLogV = new(warnLog)
		getWarnLog()
	}()
}
