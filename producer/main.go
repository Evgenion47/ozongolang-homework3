package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	consts "main/const"
	"math/rand"
	"time"
)

func main() {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(consts.Brokers, cfg)
	if err != nil {
		log.Fatalf("sync kafka: %v", err)
	}

	for {
		d := consts.Order{
			IdOrder: int(time.Now().UnixNano()),
			IdUser:  rand.Intn(320),
			Details: []consts.Goods{
				{
					IdGoods: "example.com",
					Amount:  3,
				},
				{
					IdGoods: "1example.com",
					Amount:  4,
				},
			},
		}
		b, err := json.Marshal(d)
		if err != nil {
			log.Printf("wtf? %v", err)
			continue
		}
		par, off, err := syncProducer.SendMessage(&sarama.ProducerMessage{
			Topic: "income_orders",
			Key:   sarama.StringEncoder(fmt.Sprintf("%v", d.IdOrder)),
			Value: sarama.ByteEncoder(b),
		})
		log.Printf("order %v -> %v; %v", par, off, err)
		time.Sleep(time.Millisecond * 5000)
	}
}
