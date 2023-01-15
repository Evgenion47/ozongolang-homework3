package models

type OrderId struct {
	IdOrder int
}

type IdGoods struct {
	IdGoods string
}

type IdGoodsWAmount struct {
	IdGoods string
	Amount  int
}

type OrderWGoods struct {
	IdOrder        int
	IdUser         int
	IdGoodsWAmount []IdGoodsWAmount
}

type AmountWCost struct {
	Amount int
	Cost   int
}

type Cost struct {
	IdGoods string
	Cost    int
}

type OrderToPayment struct {
	IdOrder     int
	IdUser      int
	AmountWCost []AmountWCost
}
