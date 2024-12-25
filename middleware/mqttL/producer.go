package mqttL

import (
	"context"
	"encoding/json"
	"github.com/flyerxp/lib/v2/logger"
	json2 "github.com/flyerxp/lib/v2/utils/json"
	"go.uber.org/zap"
	"time"
)

type MqttMessage struct {
	ProducerTime  int64           `json:"producer_time"`
	ProducerTimeS string          `json:"producer_time_s"`
	From          string          `json:"from"`
	Topic         string          `json:"topic"`
	RequestId     string          `json:"request_id"`
	Content       json.RawMessage `json:"content"`
	DelayTime     time.Duration   `json:"delay_time"`
}

func (p MqttMessage) String() string {
	tmp, _ := json2.Encode(p)
	return string(tmp)
}

type OutMessage struct {
	TopicStr   string            `json:"topic_str"`
	Method     string            `json:"method"`
	Content    any               `json:"content"`
	Properties map[string]string `json:"properties"`
	Key        string            `json:"key"`
	Track      string            `json:"track"`
	Retained   bool              `json:"retained"`
	Qos        byte              `json:"qos"`
}

func getMqttMessage(o *OutMessage, codeStr string, ctx context.Context) *MqttMessage {
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
	return &MqttMessage{
		ProducerTime:  time.Now().UnixMilli(),
		ProducerTimeS: time.Now().Format("2006-01-02 15:04:05"),
		From:          o.Track,
		Topic:         codeStr,
		RequestId:     o.Properties["GlobalRequestId"],
		Content:       BContent,
	}

}
func Flush(ctx context.Context) {
	if producerQue != nil {
		producerQue.Wg.Wait()
	}
}
