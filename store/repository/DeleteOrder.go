package repository

import (
	"errors"
	"golang.org/x/net/context"
	"log"
	"main/store/repository/models"
)

func (r *Repository) DeleteOrder(ctx context.Context, ord models.OrderId) (err error) {
	const (
		query1 = `
		select "IdGoods" from Dict
		where "IdOrder" = $1;
`
		query2 = `
		select "Amount" from Dict
		where "IdOrder" = $1 and "IdGoods" = $2;
`
		query3 = `
		update Goods
		set "AmountOnWH" = "AmountOnWH" + $1
		where "IdGoods" = $2;
`
		query4 = `
		update Orders
		set "State" = false
		where "IdOrder" = $1;
`
	)

	rows, err := r.pool.Query(ctx, query1, ord.IdOrder)
	if err != nil {
		err = errors.New("can't get (you out of my head) IdGoods from order(1)")
		return err
	}
	defer rows.Close()

	var Goods []string

	for rows.Next() {
		var foo string
		if err = rows.Scan(&foo); err != nil {
			err = errors.New("can't get (you out of my head) IdGoods from order(2)")
			return err
		}
		Goods = append(Goods, foo)
	}
	for i := 0; i < len(Goods); i++ {
		row := r.pool.QueryRow(ctx, query2, ord.IdOrder, Goods[i])
		var TmpAmount int
		err = row.Scan(&TmpAmount)
		if err != nil {
			log.Println(err)
			err = errors.New("can't get (you out of my head) Amount from order")
			return err
		}
		_, err = r.pool.Exec(ctx, query3, TmpAmount, Goods[i])
		if err != nil {
			err = errors.New("can't update Amount on warehouse(1)")
			return err
		}
	}
	_, err = r.pool.Exec(ctx, query4, ord.IdOrder)
	if err != nil {
		log.Println(err)
		err = errors.New("can't update state in order")
		return err
	}
	return err
}
