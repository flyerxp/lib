package pulsarL

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/flyerxp/lib/v2/app"
	"github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	json2 "github.com/flyerxp/lib/v2/utils/json"
	"github.com/flyerxp/lib/v2/utils/netL"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"sync/atomic"
	"time"
)

type Consumer struct {
	Topics            map[string][]string
	Name              string
	Options           *Options
	ConsumerContainer map[string]pulsar.Consumer
	ConnContainer     map[string]*PulsarClient
	IsStop            bool
	Counter           int64
}
type Options struct {
	Name        int
	Dlq         *pulsar.DLQPolicy
	MaxConsumer int
}
type Option func(opts *Options)

func NewConsumer(ctx context.Context, s []string, subName string, f ...Option) *Consumer {
	c := new(Consumer)
	c.Topics = map[string][]string{}
	for _, v := range s {
		t, _ := getTopic(ctx, v)
		if _, ok := c.Topics[t.Cluster]; ok {
			c.Topics[t.Cluster] = append(c.Topics[t.Cluster], t.CodeStr)
		} else {
			c.Topics[t.Cluster] = []string{t.CodeStr}
		}
	}
	c.Name = subName
	c.Counter = 0
	c.Options = loadOptions(f...)
	return c
}

func GetStringTopics(i []int) []string {
	var s = make([]string, len(i), len(i))
	for ii, v := range i {
		s[ii] = strconv.Itoa(v)
	}
	return s
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

// 死信和失败保存的topic
func WithDlq(policy pulsar.DLQPolicy) Option {
	return func(opts *Options) {
		var p = policy
		opts.Dlq = &p
	}
}
func (c *Consumer) Consumer(F func(context.Context, *pulsar.ConsumerMessage, *PulsarMessage) bool) {
	var MessageChannel = make(chan pulsar.ConsumerMessage)
	c.ConnContainer = make(map[string]*PulsarClient)
	c.ConsumerContainer = make(map[string]pulsar.Consumer)
	var err error
	ackGroup := pulsar.AckGroupingOptions{MaxSize: 1000, MaxTime: 100 * time.Millisecond}
	dlq := new(pulsar.DLQPolicy)
	if c.Options.Dlq == nil {
		dlq.MaxDeliveries = 20
	} else {
		dlq = c.Options.Dlq
	}
	if dlq.DeadLetterTopic == "" {
		dlq.DeadLetterTopic = "dead_letter_topic"
	}
	if dlq.RetryLetterTopic == "" {
		dlq.RetryLetterTopic = "retry_letter_topic"
	}
	ctx := logger.GetContext(context.Background(), "initConsumer")
	for cluster, gTopics := range c.Topics {
		c.ConnContainer[cluster], err = GetEngine(ctx, cluster)
		if err != nil {
			logger.AddError(ctx, zap.Error(err))
			logger.WriteErr(ctx)
		}
		c.ConsumerContainer[cluster], err = c.ConnContainer[cluster].CurrPulsar.Subscribe(pulsar.ConsumerOptions{
			Topics:                 gTopics,
			SubscriptionName:       c.Name,
			Name:                   config.GetConf().App.Name + "_" + c.Name,
			Type:                   pulsar.Shared,
			MessageChannel:         MessageChannel,
			AutoAckIncompleteChunk: true,
			AckGroupingOptions:     &ackGroup,
			DLQ:                    dlq,
			RetryEnable:            true,
		})
		if err != nil {
			logger.AddError(ctx, zap.Error(err))
			logger.WriteErr(ctx)
		}
	}
	defer func() {
		for i := range c.ConsumerContainer {
			c.ConsumerContainer[i].Close()
		}
		app.Shutdown(ctx)
	}()
	for i := range c.ConsumerContainer {
		c.ConsumerContainer[i].Subscription()
	}
	ip := netL.GetIp()
	for !c.IsStop {
		select {
		case cm, ok := <-MessageChannel:
			if ok {
				ctx = logger.GetContext(context.Background(), "cmr-"+uuid.New().String())
				start := time.Now()
				logger.SetRefer(ctx, "consumer")
				logger.SetArgs(ctx, string(cm.Payload()))
				logger.SetUrl(ctx, c.Name)
				logger.SetAddr(ctx, ip[0].String(), ip[0].String())
				atomic.AddInt64(&c.Counter, 1)
				gProductChan := new(PulsarMessage)
				err = json2.Decode(cm.Payload(), gProductChan)
				if err != nil {
					logger.AddError(ctx, zap.Error(err))
					logger.WriteErr(ctx)
				}
				if prop, okRetry := cm.Message.Properties()["RECONSUMETIMES"]; okRetry {
					gProductChan.Properties["RECONSUMETIMES"] = prop
				}
				_ = F(ctx, &cm, gProductChan)
				logger.SetExecTime(ctx, int(time.Since(start).Microseconds()))
				logger.WriteLine(ctx)
				err = cm.Consumer.Ack(cm.Message)
				if err != nil {
					logger.AddError(ctx, zap.Error(err))
				}
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
func RetryAfter(message *pulsar.ConsumerMessage, t time.Duration, m map[string]string) {
	message.Consumer.ReconsumeLaterWithCustomProperties(message, m, time.Second*t)
}
func (c *Consumer) Stop() {
	c.IsStop = true
}
