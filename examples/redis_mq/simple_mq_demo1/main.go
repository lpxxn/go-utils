package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis"
	"github.com/lpxxn/go-utils/redis_mq"
)

func main() {
	// 1 use client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 2 or use cluster
	//clusterSlots := func() ([]redis.ClusterSlot, error) {
	//	slots := []redis.ClusterSlot{
	//		{
	//			Start: 0,
	//			End:   16383,
	//			Nodes: []redis.ClusterNode{
	//				{
	//					Addr: "localhost:6379",
	//				}, {
	//					Addr: "localhost:6379",
	//				},
	//			},
	//		},
	//	}
	//	return slots, nil
	//}
	//client := redis.NewClusterClient(&redis.ClusterOptions{
	//	ClusterSlots:  clusterSlots,
	//	RouteRandomly: true,
	//})
	if err := client.Ping().Err(); err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())

	topicName := "testTopic1"
	// normal
	consumer := redis_mq.NewSimpleMQConsumer(ctx, client, topicName)
	// use LBPop
	//consumer := redis_mq.NewSimpleMQConsumer(ctx, client, topicName, redis_mq.UseBLPop(true), redis_mq.NewRateLimitPeriod(time.Millisecond*100))
	consumer.SetHandler(&MyHandler{})

	go func() {
		ticker := time.NewTicker(time.Second / 10)
		producer := redis_mq.NewProducer(client)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stop produce...")
				return
			case <-ticker.C:
				msg := &MyMsg{
					Name: fmt.Sprintf("name_%d", rand.Int()),
					Age:  rand.Intn(20),
				}
				body, _ := json.Marshal(msg)
				if err := producer.Publish(topicName, body); err != nil {
					panic(err)
				}
				if err := producer.PublishDelayMsg(topicName, body, time.Second); err != nil {
					panic(err)
				}
			}

		}
	}()

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)
	<-stopCh
	cancel()
	fmt.Println("stop server")
}

type MyMsg struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type MyHandler struct{}

func (*MyHandler) HandleMessage(m *redis_mq.Message) {
	revMsg := &MyMsg{}
	if err := json.Unmarshal(m.Body, revMsg); err != nil {
		fmt.Printf("handle message error: %#v \n", err)
		return
	}
	fmt.Printf("receive msg: %#v \n", *revMsg)
}
