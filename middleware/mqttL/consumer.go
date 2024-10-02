package mqttL

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/flyerxp/lib/v2/app"
	"github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	json2 "github.com/flyerxp/lib/v2/utils/json"
	"github.com/flyerxp/lib/v2/utils/netL"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

type Consumer struct {
	Topics            map[string][]string
	Name              string
	CleanSession      bool //是否清理session,如不设置为false ，会消费到历史消息
	Options           *Options
	ConsumerContainer map[string]mqtt.Token
	ConnContainer     map[string]*MqttClient
	IsStop            bool
	Counter           int64
}
type Options struct {
	Name        int
	MaxConsumer int
	DecodeFun   func([]byte, *MqttMessage) error
}
type Option func(opts *Options)

func NewConsumer(ctx context.Context, s map[string][]string, subName string, f ...Option) *Consumer {
	c := new(Consumer)
	c.Topics = map[string][]string{}
	for cluster, topics := range s {
		c.Topics[cluster] = append(c.Topics[cluster], topics...)
	}
	c.Name = subName
	c.Counter = 0
	c.Options = loadOptions(f...)
	return c
}

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}
func WithMaxConsumerCnt(i int) Option {
	return func(opts *Options) {
		opts.MaxConsumer = i
	}
}

type MqttOriMessage struct {
	Duplicate bool
	Qos       byte
	Retained  bool
	Topic     string
	MessageID uint16
}

func (c *Consumer) Consumer(F func(context.Context, mqtt.Message, *MqttMessage) bool) {
	var MessageChannel = make(chan mqtt.Message)
	c.ConnContainer = make(map[string]*MqttClient)
	c.ConsumerContainer = make(map[string]mqtt.Token)
	var err error
	ctx := logger.GetContext(context.Background(), "initMqttConsumer")
	defer func() {
		for cluster, gTopics := range c.Topics {
			c.ConnContainer[cluster].GetMqtt().Unsubscribe(gTopics...)
		}
		for i := range c.ConsumerContainer {
			c.ConnContainer[i].GetMqtt().Disconnect(2500)
		}
		app.Shutdown(ctx)
	}()
	for cluster, gTopics := range c.Topics {
		c.ConnContainer[cluster], err = GetEngine(ctx, cluster, func(ops *mqtt.ClientOptions) {
			ops.SetClientID(config.GetConf().App.Name + "_" + c.Name)
			ops.SetCleanSession(false)
			ops.SetWriteTimeout(time.Second * 10)
		})
		if !c.ConnContainer[cluster].GetMqtt().IsConnected() {
			conn := c.ConnContainer[cluster].GetMqtt().Connect()
			conn.Wait()
		}
		if err != nil {
			logger.AddError(ctx, zap.Error(err))
			logger.WriteErr(ctx)
		}
		tm := make(map[string]byte)
		for i := range gTopics {
			tm[gTopics[i]] = 1
		}
		c.ConsumerContainer[cluster] = c.ConnContainer[cluster].GetMqtt().SubscribeMultiple(tm, func(client mqtt.Client, msg mqtt.Message) {
			MessageChannel <- msg
		})
	}
	ip := netL.GetIp()
	for !c.IsStop {
		select {
		case cm, ok := <-MessageChannel:
			if ok {
				ctx = logger.GetContext(context.Background(), "mqttCmr-"+uuid.New().String())
				start := time.Now()
				logger.SetRefer(ctx, "consumer")
				logger.SetArgs(ctx, string(cm.Payload()))
				logger.SetUrl(ctx, c.Name)
				logger.SetAddr(ctx, ip[0].String(), ip[0].String())
				atomic.AddInt64(&c.Counter, 1)
				gProductChan := new(MqttMessage)
				if c.Options.DecodeFun != nil {
					gProductChan.Topic = cm.Topic()
					err = c.Options.DecodeFun(cm.Payload(), gProductChan)
				} else {
					err = json2.Decode(cm.Payload(), gProductChan)
				}
				if err != nil {
					logger.AddError(ctx, zap.Error(err))
					logger.WriteErr(ctx)
				}
				_ = F(ctx, cm, gProductChan)
				logger.SetExecTime(ctx, int(time.Since(start).Microseconds()))
				logger.WriteLine(ctx)
				cm.Ack()
				if c.Options.MaxConsumer > 0 && c.Counter >= int64(c.Options.MaxConsumer) {
					c.Stop()
				}
			} else {
				logger.AddError(ctx, zap.String("consumer", " chan no ok"))
				logger.WriteErr(ctx)
			}
		case <-time.After(time.Second * 30):
			fmt.Println("30秒没有消息")
		case <-ctx.Done():
			c.IsStop = true
		}
	}
}
func (c *Consumer) Stop() {
	c.IsStop = true
}
