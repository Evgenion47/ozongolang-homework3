package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"golang.org/x/net/context"
	"gorm.io/gorm"
	"log"
	consts "main/const"
	dbCH "main/payment/db"
	"math/rand"
	"time"
)

type Payment struct {
	conn            *gorm.DB
	producer        sarama.SyncProducer
	paymentConsumer *paymentHandler
}

type paymentHandler struct {
	P    sarama.SyncProducer
	conn *gorm.DB
	ctx  context.Context
}

func NewPayment(ctx context.Context) (*Payment, error) {
	db := dbCH.NewConnCH()

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(consts.Brokers, cfg)
	if err != nil {
		return nil, err
	}
	income, err := sarama.NewConsumerGroup(consts.Brokers, "payment", cfg)
	if err != nil {
		return nil, err
	}
	pHandler := &paymentHandler{
		P:    producer,
		conn: db,
		ctx:  ctx,
	}
	go func() {
		for {
			if err := income.Consume(ctx, []string{"income_payments"}, pHandler); err != nil {
				log.Printf("income payment error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	return &Payment{
		conn:            dbCH.NewConnCH(),
		producer:        producer,
		paymentConsumer: pHandler,
	}, nil
}

func (i *paymentHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (i *paymentHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (i *paymentHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var d consts.OrderToPayment

		if err := json.Unmarshal(msg.Value, &d); err != nil {
			log.Print("income data %v: %v", string(msg.Value), err)
			continue
		}

		var totalsum int

		for i := 0; i < len(d.Details); i++ {
			totalsum += d.Details[i].Amount * d.Details[i].Cost
		}

		tmp := rand.Intn(30)
		var successPayment bool
		if tmp%2 == 0 {
			successPayment = true
		}

		if successPayment {
			dbCH.CreateResult(i.conn, int64(d.IdOrder), int64(d.IdUser), totalsum)
			log.Printf("Order %v paid", d.IdOrder)

		} else {
			b, err := json.Marshal(consts.RollbackInfo{IdOrder: d.IdOrder})
			if err != nil {
				log.Printf("wtf? %v", err)
				continue
			}
			par, off, err := i.P.SendMessage(&sarama.ProducerMessage{
				Topic: "reset_orders",
				Key:   sarama.StringEncoder(fmt.Sprintf("%v", d.IdOrder)),
				Value: sarama.ByteEncoder(b),
			})
			log.Printf("order to reset %v -> %v; %v", par, off, err)
			time.Sleep(time.Millisecond * 500)
			log.Printf("Order %v rollbacked", d.IdOrder)
		}
	}
	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if _, err := NewPayment(ctx); err != nil {
		log.Fatalf("NewPayment: %v", err)
	}
	<-ctx.Done()
}
