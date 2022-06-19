package consts

var (
	Brokers = []string{"localhost:9095", "localhost:9096"}
)

type Goods struct {
	IdGoods string `json:"idGoods"`
	Amount  int    `json:"amount"`
}

type AmountWCost struct {
	Amount int `json:"amount"`
	Cost   int `json:"cost"`
}

type Order struct {
	IdOrder int     `json:"idOrder"`
	IdUser  int     `json:"idUser"`
	Details []Goods `json:"details"`
}

type OrderToPayment struct {
	IdOrder int           `json:"idOrder"`
	IdUser  int           `json:"idUser"`
	Details []AmountWCost `json:"details"`
}

type RollbackInfo struct {
	IdOrder int `json:"idOrder"`
}

type Result struct {
	IdOrder   int `json:"idOrder"`
	IdUser    int `json:"idUser"`
	TotalCost int `json:"totalCost"`
}
