package pulsarL

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/flyerxp/lib/logger"
	json2 "github.com/flyerxp/lib/utils/json"
	cmap "github.com/orcaman/concurrent-map/v2"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type PulsarMessage struct {
	ProducerTime  int64             `json:"producer_time"`
	ProducerTimeS string            `json:"producer_time_s"`
	From          string            `json:"from"`
	Topic         string            `json:"topic"`
	RequestId     string            `json:"request_id"`
	Content       json.RawMessage   `json:"content"`
	DelayTime     time.Duration     `json:"delay_time"`
	Properties    map[string]string `json:"properties"`
}

func (p PulsarMessage) String() string {
	tmp, _ := json2.Encode(p)
	return string(tmp)
}

type OutMessage struct {
	Topic      int               `json:"topic"`
	TopicStr   string            `json:"topic_str"`
	Content    any               `json:"content"`
	Properties map[string]string `json:"properties"`
	Delay      int               `json:"delay"`
	Key        string            `json:"key"`
	Track      string            `json:"track"`
}

var producerQue *pulsarProducer

type pulsarProducer struct {
	//Pool     *ants.Pool
	Que      cmap.ConcurrentMap[string, pulsar.Producer]
	isInitEd bool
	Sending  int32
	Wg       sync.WaitGroup
}

func init() {
	initProducer()
}
func initProducer() {
	producerQue = new(pulsarProducer)
	producerQue.isInitEd = true
	producerQue.Que = cmap.New[pulsar.Producer]()
}

func Producer(ctx context.Context, o *OutMessage) error {
	start := time.Now()
	codeStr := o.TopicStr
	if o.Topic > 0 {
		codeStr = strconv.Itoa(o.Topic)
	}
	objTopic, ok := getTopic(ctx, codeStr)
	if !ok {
		panic(errors.New(fmt.Sprintf("%s no find message %s", codeStr, o.Content)))
	}
	pMessage := getPulsarMessage(o, objTopic, ctx)
	pClient, e := GetEngine(ctx, objTopic.Cluster)
	if e != nil {
		logger.AddError(ctx, zap.Error(e), zap.Any("message", o))
	}
	var p pulsar.Producer
	p, ok = producerQue.Que.Get(codeStr)
	if !ok {
		//官方此处存在性能问题,协程下直接卡死
		//，NewRequestID
		p, e = pClient.CurrPulsar.CreateProducer(pulsar.ProducerOptions{
			Topic:              codeStr,
			ProducerAccessMode: pulsar.ProducerAccessModeShared,
			//DisableBlockIfQueueFull: true,
			BatchingMaxSize:                 1048576, //1M
			SendTimeout:                     time.Second * 5,
			BatchingMaxPublishDelay:         1000 * time.Millisecond,
			BatchingMaxMessages:             100,
			PartitionsAutoDiscoveryInterval: time.Second * 86400 * 5,
		})

		if p != nil {
			producerQue.Que.Set(codeStr, p)
		} else {
			logger.AddError(ctx, zap.Error(errors.New("producer 创建失败"+codeStr+":"+objTopic.Cluster)))
			logger.AddError(ctx, zap.Error(e))
			logger.WriteErr(ctx)
			return errors.New("topic 信息获取失败")
		}
	}
	if e != nil {
		logger.AddError(ctx, zap.String("pulsar", "pulsar producer error "+e.Error()), zap.Error(e))
		logger.WriteErr(ctx)
		logger.AddPulsarTime(ctx, int(time.Since(start).Microseconds()))
		panic(errors.New("pulsar producer create timeout," + e.Error()))
	}
	atomic.AddInt32(&producerQue.Sending, 1)
	producerQue.Wg.Add(1)
	p.SendAsync(ctx, pMessage, func(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
		if err != nil {
			logger.AddError(ctx, zap.Error(err), zap.Any(codeStr, pMessage.Payload))
		}
		atomic.AddInt32(&producerQue.Sending, -1)
		producerQue.Wg.Done()
	})
	logger.AddPulsarTime(ctx, int(time.Since(start).Microseconds()))
	return nil
}
func getPulsarMessage(o *OutMessage, objTopic *TopicS, ctx context.Context) *pulsar.ProducerMessage {
	if o.Properties == nil {
		o.Properties = map[string]string{}
	}
	//全局唯一id
	if rId, ok := ctx.Value("GlobalRequestId").(string); ok {
		o.Properties["GlobalRequestId"] = rId
	} else if rId, ok = ctx.Value(logger.GetLogIdKey()).(string); ok {
		o.Properties["GlobalRequestId"] = rId
	} else {
		o.Properties["GlobalRequestId"] = ""
	}
	BContent, err := json2.Encode(o.Content)
	if err != nil {
		logger.AddError(ctx, zap.Error(err))
		panic(err)
	}

	codeStr := o.TopicStr
	if o.Topic > 0 {
		codeStr = strconv.Itoa(objTopic.Code)
	}
	payload := PulsarMessage{
		ProducerTime:  time.Now().UnixMilli(),
		ProducerTimeS: time.Now().Format("2006-01-02 15:04:05"),
		From:          o.Track,
		Topic:         codeStr,
		RequestId:     o.Properties["GlobalRequestId"],
		Content:       BContent,
		DelayTime:     time.Duration(objTopic.Delay),
		Properties:    o.Properties,
	}
	if o.Delay > 0 {
		payload.DelayTime = time.Duration(o.Delay)
	}
	payloadB, errJ := json2.Encode(payload)
	if errJ != nil {
		logger.AddError(ctx, zap.Error(errJ), zap.Any("fail payload", payload))
		panic(errors.New("json Fail"))
	}
	return &pulsar.ProducerMessage{
		Payload:      payloadB,
		Properties:   payload.Properties,
		DeliverAfter: time.Second * payload.DelayTime,
		Key:          o.Key,
	}
}
func producerPreInit(ctx context.Context, t []string) {
	var p pulsar.Producer
	var err error
	for _, codeStr := range t {
		objTopic, ok := getTopic(ctx, codeStr)
		if !ok {
			logger.AddError(ctx, zap.Error(errors.New("topic no find "+codeStr)))
			return
		}
		pClient, _ := GetEngine(ctx, objTopic.Cluster)
		p, ok = producerQue.Que.Get(codeStr)
		if !ok {
			//官方此处存在性能问题
			p, err = pClient.CurrPulsar.CreateProducer(pulsar.ProducerOptions{
				Topic:              codeStr,
				ProducerAccessMode: pulsar.ProducerAccessModeShared,
				//DisableBlockIfQueueFull: true,
				BatchingMaxSize:                 1048576, //1M
				SendTimeout:                     time.Second * 5,
				BatchingMaxPublishDelay:         2000 * time.Millisecond,
				BatchingMaxMessages:             100,
				PartitionsAutoDiscoveryInterval: time.Second * 86400 * 5,
			})
			if err == nil && p != nil {
				producerQue.Que.Set(codeStr, p)
			}
		}
	}

}
func Flush(ctx context.Context) {
	if producerQue != nil {
		producerQue.Wg.Wait()
		for _, v := range producerQue.Que.Items() {
			if v.LastSequenceID() > 0 {
				e := v.Flush()
				if e != nil {
					logger.AddError(ctx, zap.Error(e))
				}
			}
		}
	}
}
