package repository

import (
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/net/context"
	"log"
	"main/store/repository/models"
)

func (r *Repository) GetValueOfProduct(ctx context.Context, ord models.IdGoods) (res models.Cost, err error) {
	const (
		query1 = `
		select "Cost" from Goods
		where "IdGoods" = $1;
`
	)
	var cost models.Cost
	if cost, err = r.cache.Get(ctx, ord.IdGoods); errors.Is(err, memcache.ErrCacheMiss) {
		row := r.pool.QueryRow(ctx, query1, ord.IdGoods)
		err = row.Scan(&cost)
		if err != nil {
			log.Println(err)
			err = errors.New("can't read cost of product")
			return res, err
		}
		if err = r.cache.Set(ctx, cost); err != nil {
			return res, err
		}
	}
	return res, err
}
