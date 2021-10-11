package models

/*
Модели данных:
1) OrderGet – структура, инкапсулирующая сведения о данных поступившего заказа.
Ключевой параметр: OrderUID.
2) Payment – структура, инкапсулирующая сведения об оплате поступившего заказа.
Ключевой параметр: OrderUID.
3) Items – структура, инкапсулирующая сведения о наборе товаров в поступившем заказа.
Ключевой параметр: OrderUID.
ВАЖНО: благодаря внедрению ключевого параметра OrderUID в дальнейшем возможно идентифицировать заказ,
набор товаров из него, а так же способ и порядок оплаты. Это упрощает идентификацию данных, ускоряет их обработку.
*/

type OrderGet struct {
	OrderUID          string  `json:"order_uid"`
	Entry             string  `json:"entry"`
	InternalSignature string  `json:"internal_signature"`
	Payment           Payment `json:"payment"`
	Items             []Items `json:"items"`
	Locale            string  `json:"locale"`
	CustomerID        string  `json:"customer_id"`
	TrackNumber       string  `json:"track_number"`
	DeliveryService   string  `json:"delivery_service"`
	Shardkey          string  `json:"shardkey"`
	SmID              int     `json:"sm_id"`
}

type Payment struct {
	OrderUID     string
	Transaction  string `json:"transaction"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int    `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
}

type Items struct {
	OrderUID   string
	ChrtID     int    `json:"chrt_id"`
	Price      int    `json:"price"`
	Rid        string `json:"rid"`
	Name       string `json:"name"`
	Sale       int    `json:"sale"`
	Size       string `json:"size"`
	TotalPrice int    `json:"total_price"`
	NmID       int    `json:"nm_id"`
	Brand      string `json:"brand"`
}
