package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"golang.org/x/net/context"
	"log"
	consts "main/const"
	"main/store/db"
	"main/store/repository"
	"main/store/repository/models"
	"main/tools/memcached"
	"time"
)

type Store struct {
	repository     *repository.Repository
	producer       sarama.SyncProducer
	incomeConsumer *IncomeHandler
	resetConsumer  *ResetHandler
}

type ResetHandler struct {
	P          sarama.SyncProducer
	repository *repository.Repository
	ctx        context.Context
}

type IncomeHandler struct {
	P          sarama.SyncProducer
	repository *repository.Repository
	ctx        context.Context
}

func NewStore(ctx context.Context) (*Store, error) {
	adp, err := db.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	cache := memcached.New()
	repo := repository.New(adp, cache)

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(consts.Brokers, cfg)
	if err != nil {
		return nil, err
	}
	income, err := sarama.NewConsumerGroup(consts.Brokers, "store", cfg)
	if err != nil {
		return nil, err
	}
	iHandler := &IncomeHandler{
		P:          producer,
		repository: repo,
		ctx:        ctx,
	}
	go func() {
		for {
			if err := income.Consume(ctx, []string{"income_orders"}, iHandler); err != nil {
				log.Printf("income consumer error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()

	reset, err := sarama.NewConsumerGroup(consts.Brokers, "storeReset", cfg)
	if err != nil {
		return nil, err
	}
	rHandler := &ResetHandler{
		P:          producer,
		repository: repo,
		ctx:        ctx,
	}
	go func() {
		for {
			if err := reset.Consume(ctx, []string{"reset_orders"}, rHandler); err != nil {
				log.Printf("reset consumer error: %v", err)
				time.Sleep(time.Second * 5)
			}
		}
	}()
	return &Store{
		repository:     repository.New(adp, cache),
		producer:       producer,
		incomeConsumer: iHandler,
		resetConsumer:  rHandler,
	}, nil
}

func (i *IncomeHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (i *IncomeHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (i *IncomeHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var d consts.Order

		if err := json.Unmarshal(msg.Value, &d); err != nil {
			log.Print("income data %v: %v", string(msg.Value), err)
			continue
		}
		//здесь должна быть запись заказа
		var tmpIdGoodsWAmount []models.IdGoodsWAmount

		for i := 0; i < len(d.Details); i++ {
			tmpIdGoodsWAmount = append(tmpIdGoodsWAmount, models.IdGoodsWAmount{d.Details[i].IdGoods, d.Details[i].Amount})
		}

		resp, err := i.repository.CreateOrder(i.ctx, models.OrderWGoods{IdOrder: d.IdOrder, IdUser: d.IdUser, IdGoodsWAmount: tmpIdGoodsWAmount})
		if err != nil {
			log.Println(err)
		} else {
			var d consts.OrderToPayment
			d.IdOrder, d.IdUser = resp.IdOrder, resp.IdUser
			var tmpAmountsWCost []consts.AmountWCost

			for i := 0; i < len(resp.AmountWCost); i++ {
				tmpAmountsWCost = append(tmpAmountsWCost, consts.AmountWCost{Cost: resp.AmountWCost[i].Cost, Amount: resp.AmountWCost[i].Amount})
			}
			d.Details = tmpAmountsWCost

			b, err := json.Marshal(d)
			if err != nil {
				log.Printf("wtf? %v", err)
				continue
			}
			par, off, err := i.P.SendMessage(&sarama.ProducerMessage{
				Topic: "income_payments",
				Key:   sarama.StringEncoder(fmt.Sprintf("%v", d.IdOrder)),
				Value: sarama.ByteEncoder(b),
			})
			log.Printf("order to payment %v -> %v; %v", par, off, err)
			time.Sleep(time.Millisecond * 500)
			log.Printf("Order %v reserved", d.IdOrder)
		}
	}
	return nil
}

func (r *ResetHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (r *ResetHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (r *ResetHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var d consts.RollbackInfo

		if err := json.Unmarshal(msg.Value, &d); err != nil {
			log.Print("reset data %v: %v", string(msg.Value), err)
			continue
		}
		//здесь должно быть удаление заказа
		err := r.repository.DeleteOrder(r.ctx, models.OrderId{IdOrder: d.IdOrder})
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Order %v deleted", d.IdOrder)
	}
	return nil
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if _, err := NewStore(ctx); err != nil {
		log.Fatalf("NewStore: %v", err)
	}
	<-ctx.Done()
}
