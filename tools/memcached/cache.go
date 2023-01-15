package memcached

import (
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"main/store/repository/models"
)

type cache struct {
	client *memcache.Client
}

func New() *cache {
	mc := memcache.New("localhost:11211")
	return &cache{mc}
}

func (c *cache) Get(ctx context.Context, ID string) (res models.Cost, err error) {

	span, ctx := opentracing.StartSpanFromContext(ctx, "cache")
	defer span.Finish()

	item, err := c.client.Get(fmt.Sprint(ID))
	if err != nil {
		return
	}

	if err = json.Unmarshal(item.Value, &res); err != nil {
		return
	}

	return
}

func (c *cache) Set(ctx context.Context, p models.Cost) (err error) {

	data, err := json.Marshal(&p)
	if err != nil {
		return
	}

	return c.client.Set(&memcache.Item{
		Key:   fmt.Sprint(p.IdGoods),
		Value: data,
	})
}
