package models

/*
Модель данных:
OrderPost – структура, инкапсулирующая сведения о заказе для выдачи по запросу пользователя.
ВАЖНО: благодаря внедрению ключевого параметра OrderUID в дальнейшем возможно идентифицировать заказ,
набор товаров из него, а так же способ и порядок оплаты. Это упрощает идентификацию данных, ускоряет их обработку.
*/

type OrderPost struct {
	OrderUID        string `json:"order_uid"`
	Entry           string `json:"entry"`
	TotalPrice      int    `json:"total_price"`
	CustomerID      string `json:"customer_id"`
	TrackNumber     string `json:"track_number"`
	DeliveryService string `json:"delivery_service"`
}
