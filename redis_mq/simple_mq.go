package redis_mq

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
)

type Message struct {
	ID        string `json:"id"`
	Body      []byte `json:"body"`
	Timestamp int64  `json:"timestamp"`
	_         struct{}
}

func NewMessage(id string, body []byte) *Message {
	if id == "" {
		id = uuid.NewV4().String()
	}
	return &Message{
		ID:        id,
		Body:      body,
		Timestamp: time.Now().Unix(),
	}
}

type Handler interface {
	HandleMessage(msg *Message)
}

type consumer struct {
	once            sync.Once
	redisCmd        redis.Cmdable
	ctx             context.Context
	topicName       string
	handler         Handler
	rateLimitPeriod time.Duration
	options         ConsumerOptions
	_               struct{}
}

type ConsumerOptions struct {
	RateLimitPeriod time.Duration
	UseBLPop        bool
}

type ConsumerOption func(options *ConsumerOptions)

func NewRateLimitPeriod(d time.Duration) ConsumerOption {
	return func(o *ConsumerOptions) {
		o.RateLimitPeriod = d
	}
}

func UseBLPop(u bool) ConsumerOption {
	return func(o *ConsumerOptions) {
		o.UseBLPop = u
	}
}

type Consumer = *consumer

func NewSimpleMQConsumer(ctx context.Context, redisCmd redis.Cmdable, topicName string, opts ...ConsumerOption) Consumer {
	consumer := &consumer{
		redisCmd:  redisCmd,
		ctx:       ctx,
		topicName: topicName,
	}
	for _, o := range opts {
		o(&consumer.options)
	}
	if consumer.options.RateLimitPeriod == 0 {
		consumer.options.RateLimitPeriod = time.Microsecond * 200
	}
	return consumer
}

func (s *consumer) SetHandler(handler Handler) {
	s.once.Do(func() {
		s.startGetMessage()
	})
	s.handler = handler
}

func (s *consumer) startGetMessage() {
	go func() {
		ticker := time.NewTicker(s.options.RateLimitPeriod)
		defer func() {
			log.Println("stop get message.")
			ticker.Stop()
		}()
		for {
			select {
			case <-s.ctx.Done():
				log.Printf("context Done msg: %#v \n", s.ctx.Err())
				return
			case <-ticker.C:
				var revBody []byte
				var err error
				if !s.options.UseBLPop {
					revBody, err = s.redisCmd.LPop(s.topicName).Bytes()
				} else {
					revs := s.redisCmd.BLPop(time.Second, s.topicName)
					err = revs.Err()
					revValues := revs.Val()
					if len(revValues) >= 2 {
						revBody = []byte(revValues[1])
					}
				}
				if err == redis.Nil {
					continue
				}
				if err != nil {
					log.Printf("LPOP error: %#v \n", err)
					continue
				}

				if len(revBody) == 0 {
					continue
				}
				msg := &Message{}
				json.Unmarshal(revBody, msg)
				if s.handler != nil {
					s.handler.HandleMessage(msg)
				}
			}
		}
	}()
}

type Producer struct {
	redisCmd redis.Cmdable
	_        struct{}
}

func NewProducer(cmd redis.Cmdable) *Producer {
	return &Producer{redisCmd: cmd}
}

func (p *Producer) Publish(topicName string, body []byte) error {
	msg := NewMessage("", body)
	sendData, _ := json.Marshal(msg)
	return p.redisCmd.RPush(topicName, string(sendData)).Err()
}
