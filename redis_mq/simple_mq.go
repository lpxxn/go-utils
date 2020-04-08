package redis_mq

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
)

const (
	listSuffix, zsetSuffix = ":list", ":zset"
)

type Message struct {
	ID        string `json:"id"`
	Body      []byte `json:"body"`
	Timestamp int64  `json:"timestamp"`
	DelayTime int64  `json:"delayTime"`
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
		DelayTime: time.Now().Unix(),
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
		s.startGetListMessage()
		s.startGetDelayMessage()
	})
	s.handler = handler
}

func (s *consumer) startGetListMessage() {
	go func() {
		ticker := time.NewTicker(s.options.RateLimitPeriod)
		defer func() {
			log.Println("stop get list message.")
			ticker.Stop()
		}()
		topicName := s.topicName + listSuffix
		for {
			select {
			case <-s.ctx.Done():
				log.Printf("context Done msg: %#v \n", s.ctx.Err())
				return
			case <-ticker.C:
				var revBody []byte
				var err error
				if !s.options.UseBLPop {
					revBody, err = s.redisCmd.LPop(topicName).Bytes()
				} else {
					revs := s.redisCmd.BLPop(time.Second, topicName)
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

func (s *consumer) startGetDelayMessage() {
	go func() {
		ticker := time.NewTicker(s.options.RateLimitPeriod)
		defer func() {
			log.Println("stop get delay message.")
			ticker.Stop()
		}()
		topicName := s.topicName + zsetSuffix
		for {
			currentTime := time.Now().Unix()
			select {
			case <-s.ctx.Done():
				log.Printf("context Done msg: %#v \n", s.ctx.Err())
				return
			case <-ticker.C:
				var valuesCmd *redis.ZSliceCmd
				_, err := s.redisCmd.TxPipelined(func(pip redis.Pipeliner) error {
					valuesCmd = pip.ZRangeWithScores(topicName, 0, currentTime)
					pip.ZRemRangeByScore(topicName, "0", strconv.FormatInt(currentTime, 10))
					return nil
				})
				if err != nil {
					log.Printf("zset pip error: %#v \n", err)
					continue
				}
				rev := valuesCmd.Val()
				for _, revBody := range rev {
					msg := &Message{}
					json.Unmarshal([]byte(revBody.Member.(string)), msg)
					if s.handler != nil {
						s.handler.HandleMessage(msg)
					}
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
	return p.redisCmd.RPush(topicName+listSuffix, string(sendData)).Err()
}

func (p *Producer) PublishDelayMsg(topicName string, body []byte, delay time.Duration) error {
	if delay <= 0 {
		return errors.New("delay need great than zero")
	}
	tm := time.Now().Add(delay)
	msg := NewMessage("", body)
	msg.DelayTime = tm.Unix()

	sendData, _ := json.Marshal(msg)
	return p.redisCmd.ZAdd(topicName+zsetSuffix, redis.Z{Score: float64(tm.Unix()), Member: string(sendData)}).Err()
}
