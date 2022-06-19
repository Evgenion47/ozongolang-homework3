package repository

import (
	"errors"
	"golang.org/x/net/context"
	"log"
	"main/store/repository/models"
)

func (r *Repository) CreateOrder(ctx context.Context, ord models.OrderWGoods) (res models.OrderToPayment, err error) {
	const (
		query1 = `
		insert into Orders("IdOrder","IdUser","State")
		values ($1,$2,$3);
`
		query2 = `
		select "AmountOnWH" from Goods
		where "IdGoods" = $1;
`
		query3 = `
		update Goods
		set "AmountOnWH" = $1
		where "IdGoods" = $2;
`
		query4 = `
		insert into Dict("IdOrder","IdGoods","Amount")
		values ($1,$2,$3);
`
		query5 = `
		select "Cost" from Goods
		where "IdGoods" = $1;
`
	)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		log.Println(err)
		err = errors.New("some goes wrong with beginning transaction")
		return res, err
	}

	_, err = tx.Exec(ctx, query1, ord.IdOrder, ord.IdUser, true)
	if err != nil {
		_ = tx.Rollback(ctx)
		log.Println(err)
		err = errors.New("can't create empty order")
		return res, err
	}
	res.IdUser, res.IdOrder = ord.IdUser, ord.IdOrder

	var tmpAmountWCostslc []models.AmountWCost

	for i := 0; i < len(ord.IdGoodsWAmount); i++ {
		row := tx.QueryRow(ctx, query2, ord.IdGoodsWAmount[i].IdGoods)
		var AmOWH int
		err = row.Scan(&AmOWH)
		if err != nil {
			_ = tx.Rollback(ctx)
			log.Println(err)
			err = errors.New("can't read amounts on warehouse(1)")
			return res, err
		}

		if AmOWH > ord.IdGoodsWAmount[i].Amount {
			_, err = tx.Exec(ctx, query3, AmOWH-ord.IdGoodsWAmount[i].Amount, ord.IdGoodsWAmount[i].IdGoods)
			if err != nil {
				_ = tx.Rollback(ctx)
				err = errors.New("can't update amounts on warehouse")
				return res, err
			}
			_, err = tx.Exec(ctx, query4, ord.IdOrder, ord.IdGoodsWAmount[i].IdGoods, ord.IdGoodsWAmount[i].Amount)
			if err != nil {
				_ = tx.Rollback(ctx)
				log.Println(err)
				err = errors.New("can't update Dict")
				return res, err
			}

			row := tx.QueryRow(ctx, query5, ord.IdGoodsWAmount[i].IdGoods)
			var Cost int
			err = row.Scan(&Cost)
			if err != nil {
				_ = tx.Rollback(ctx)
				log.Println(err)
				err = errors.New("can't read amounts on warehouse(1)")
				return res, err
			}
			tmpAmountWCostslc = append(tmpAmountWCostslc, models.AmountWCost{Amount: ord.IdGoodsWAmount[i].Amount, Cost: Cost})

		} else {
			_ = tx.Rollback(ctx)
			err = errors.New("Don't have enought goods on warehouse")
			return res, err
		}
	}
	res.AmountWCost = tmpAmountWCostslc
	err = tx.Commit(ctx)
	return res, err
}
