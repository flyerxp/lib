package mqttL

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/flyerxp/lib/logger"
	"github.com/flyerxp/lib/middleware/mqttL"
	"testing"
	"time"
)

func TestProd(T *testing.T) {
	//return
	ctx := logger.GetContext(context.Background(), "test")
	time.Sleep(time.Second * 1)
	fmt.Println("开始发10条消息")
	t := time.Now()
	fmt.Println("============")
	client, err := mqttL.GetEngine(ctx, "pubMqtt")
	fmt.Println("============", err, client, t)
	/*for i := 0; i < 1; i++ {
		_ = client.Producer(ctx, &mqttL.OutMessage{
			TopicStr:   "test",
			Content:    map[string]string{"a": "b", "test": "==============test======" + strconv.Itoa(i) + "============"},
			Properties: map[string]string{"prop": "prop"},
			Delay:      0,
		})
	}
	fmt.Println(time.Since(t).Milliseconds(), "总耗时！")
	mqttL.Flush(ctx)
	logger.WriteLine(ctx)*/
}
func TestConsum(T *testing.T) {
	//return
	topics := make(map[string][]string)
	topics["pubMqtt"] = []string{"test"}
	c := mqttL.NewConsumer(logger.GetContext(context.Background(), "init"), topics, "testConsume", mqttL.WithMaxConsumerCnt(100000))
	count := 0
	c.Consumer(func(ctx context.Context, message mqtt.Message, message2 *mqttL.MqttMessage) bool {
		//c.Stop()

		fmt.Println(message2.String())
		fmt.Println(message2.Topic)
		count++
		if count == 9 {

			return true
		}
		return true
	})
}
