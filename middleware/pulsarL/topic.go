package pulsarL

import (
	"context"
	"errors"
	"fmt"
	"github.com/flyerxp/lib/v2/config"
	"github.com/flyerxp/lib/v2/logger"
	"github.com/flyerxp/lib/v2/middleware/nacos"
	yaml "github.com/flyerxp/lib/v2/utils/yaml"
	"go.uber.org/zap"
	"os"
	"strconv"
	"sync"
	"time"
)

type TopicS struct {
	Code    int    `yaml:"code" json:"code"`
	CodeStr string `yaml:"code_str" json:"code_str"`
	Delay   int    `yaml:"delay" json:"delay"`
	Cluster string `yaml:"cluster" json:"cluster"`
}
type TopicConfS struct {
	TopicDistribution map[int]string
	Topic             map[string]TopicS
	IsInitEd          bool
	isLoading         bool //是否正在异步重载
}
type yamlTopic struct {
	TopicDistribution map[int]string `yaml:"topic_distribution" json:"topic_distribution"`
	Topic             []TopicS       `yaml:"topic" json:"topic"`
}

func getCluster(ctx context.Context, code int, t map[int]string) string {
	i := int(code / 1000000)
	if c, ok := t[i]; ok {
		return c
	}
	if c, ok := topicConf.TopicDistribution[i]; ok {
		return c
	} else {
		eStr := fmt.Sprintf("topic:%d no find cluster %d/1000000 in cluster %d", code, code, i)
		logger.AddWarn(ctx, zap.Error(errors.New(eStr)))
	}
	return ""
}

var topicConf TopicConfS
var topicOnce sync.Once

func getTopic(ctx context.Context, code any) (*TopicS, bool) {
	if !topicConf.IsInitEd {
		InitTopic(ctx)
	}
	codeStr := ""
	switch code.(type) {
	case int:
		codeStr = strconv.Itoa(code.(int))
	case string:
		codeStr = code.(string)
	default:
		panic(errors.New("topic 必须是数字或字符串"))
	}
	if t, ok := topicConf.Topic[codeStr]; ok {
		if t.CodeStr == "" {
			t.CodeStr = strconv.Itoa(t.Code)
		}
		return &t, true
	} else {
		//如果没找到，自动重新载入配置
		eStr := fmt.Sprintf("topic:%s no find, 5 second reset load", codeStr)
		logger.AddWarn(ctx, zap.Error(errors.New(eStr)))
		topicConf.IsInitEd = false
		if !topicConf.isLoading {
			topicConf.isLoading = true
			time.AfterFunc(time.Second*5, func() {
				InitTopic(ctx)
				topicConf.isLoading = false
			})
		}
		return nil, false
	}
}

func InitTopic(ctx context.Context) {
	if !topicConf.IsInitEd {
		conf, err := topicDistributionF(ctx)
		if err == nil {
			topicConf = conf
		}
		topicConf.IsInitEd = true
	}
}

// 获取topic数据
func topicDistributionF(ctx context.Context) (TopicConfS, error) {
	topicConfTmp := TopicConfS{}
	topicConfTmp.TopicDistribution = make(map[int]string)
	topicConfTmp.Topic = make(map[string]TopicS)
	if getConfFile(ctx) == nil {
		tmpConf := new(yamlTopic)
		err := yaml.DecodeByFile(config.GetConfFile("pulsar.yml"), tmpConf)
		if err != nil {
			logger.AddError(ctx, zap.String("pulsal topic error", "pulsar.yml read error"), zap.Error(err))
			//return topicConfTmp, nil
		}
		getTopicConfig(ctx, &topicConfTmp, tmpConf)
	}
	conf := config.GetConf().TopicNacos
	for _, v := range conf {
		n, e := nacos.GetEngine(ctx, v.Name)
		if e != nil {
			logger.AddError(ctx, zap.Error(e))
		} else {
			b, be := n.GetConfig(context.Background(), v.Did, v.Group, v.Ns)
			if be == nil {
				tmp := new(yamlTopic)
				e = yaml.DecodeByBytes(b, tmp)
				if e != nil {
					logger.AddError(ctx, zap.Error(e))
				} else {
					getTopicConfig(ctx, &topicConfTmp, tmp)
				}
			} else {
				logger.AddError(ctx, zap.Error(e))
				logger.WriteErr(ctx)
				return topicConfTmp, e
			}
		}
	}
	return topicConfTmp, nil
}
func getConfFile(ctx context.Context) error {
	_, errf := os.Stat(config.GetConfFile("pulsar.yml"))
	if errf != nil && os.IsNotExist(errf) {
		return errf
	} else if errf != nil {
		logger.AddError(ctx, zap.String("topic read err", "read pulsar.yml err"), zap.Error(errf))
		return errf
	}
	return nil
}

// 预载Producer
func ProducerPre(ctx context.Context) {
	topicOnce.Do(func() {
		tmpInitTopic := struct {
			Topicinit []string `yaml:"topicinit"`
		}{}
		err := yaml.DecodeByFile(config.GetConfFile("topicinit.yml"), &tmpInitTopic)
		if err == nil {
			if len(tmpInitTopic.Topicinit) > 0 {
				producerPreInit(ctx, tmpInitTopic.Topicinit)
			}
		}
	})
}
func getTopicConfig(ctx context.Context, topicConfTmp *TopicConfS, tmpConf *yamlTopic) {
	for ck, cv := range tmpConf.TopicDistribution {
		topicConfTmp.TopicDistribution[ck] = cv
	}
	for _, cvt := range tmpConf.Topic {
		if cvt.CodeStr != "" {
			topicConfTmp.Topic[cvt.CodeStr] = cvt
		} else if clusterT := getCluster(ctx, cvt.Code, topicConfTmp.TopicDistribution); clusterT != "" {
			cvt.Cluster = clusterT
			topicConfTmp.Topic[strconv.Itoa(cvt.Code)] = cvt
		}
	}
}
