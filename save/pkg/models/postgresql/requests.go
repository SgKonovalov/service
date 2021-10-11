package postgresql

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"my.service.save/pkg/models"
	"my.service.save/pkg/models/postgresql/cache"
)

/*
Индивидуальные функции, использующиеся для добавления данных в каждую таблицу SQL.

1) Функция InsertNewPayment принимает в качестве аргумента канал типа OrderGet,
осуществляет переборку данных, отбирая необходимые значения полей и вносит данные в БД.
Таблица: payment, модель: Payment.
2) Функция InsertNewItems принимает в качестве аргумента канал типа OrderGet,
осуществляет переборку данных, отбирая необходимые значения полей и вносит данные в БД.
Таблица: items, модель: Items.
3) Функция InsertNewOrder принимает в качестве аргумента канал типа OrderGet,
осуществляет переборку данных, отбирая необходимые значения полей и вносит данные в БД.
Функция записывает данные в кэш и в случае сбоя передачи, организует новое соединение в БД
и вносит данные уже из кэша.
Таблица: order_get, модель: OrderGet.

ВАЖНО: вся работа по добавлению данных в SQL осуществляется на стороне БД посредством хранимых процедур.
Из приложения достаточно вызвать нужную функцию и передать ей необходимые параметры. Такой подход «избавляет»
основной процесс от обработки данных и инкапсулирует работу в БД, непосредственно в самой БД.

Хранимые процедуры:
1) insertnewitem – добавляет данные в таблицу items.
Таблица items – хранит информацию о товарах в заказе;
2) insertnewpayment – добавляет данные в таблицу payment.
Таблица payment – хранит информацию о способе заказа;
3) insertneworder – добавляет данные в таблицу order_get.
Таблица order – хранит информацию заказе.
ВАЖНО: на строне БД, указанная функцию делает выборку из таблиц items и payment по общему OrderUID
и добавляет данные в разделы payment и items таблицы order_get.
Таким образом минимизируем передачу составных данных и срезов, они обрабатываются на стороне БД.

Подробный код хранимых процедур и скрипты таблиц в файле: scripts.sql.
*/

type DbModel struct {
	DB *sql.DB
	sync.RWMutex
	sync.WaitGroup
}

var OrderCache = cache.NewCacheOrderGet(5*time.Minute, 10*time.Minute)

func (m *DbModel) InsertNewPayment(order models.OrderGet) error {

	m.RLock()
	defer m.RUnlock()

	payment := models.Payment{
		OrderUID:     order.OrderUID,
		Transaction:  order.Payment.Transaction,
		Currency:     order.Payment.Currency,
		Provider:     order.Payment.Provider,
		Amount:       order.Payment.Amount,
		PaymentDt:    order.Payment.PaymentDt,
		Bank:         order.Payment.Bank,
		DeliveryCost: order.Payment.DeliveryCost,
		GoodsTotal:   order.Payment.GoodsTotal,
	}

	stmt := "SELECT insertnewpayment ($1, $2, $3, $4, $5, $6, $7, $8)"

	_, err := m.DB.Exec(stmt, payment.OrderUID, payment.Transaction, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost)
	if err != nil {
		return err
	}
	return nil
}

func (m *DbModel) InsertNewItems(order models.OrderGet) error {

	m.RLock()
	defer m.RUnlock()

	items := order.Items

	stmt := "SELECT insertnewitem ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"

	for _, item := range items {

		insertItem := models.Items{
			OrderUID:   order.OrderUID,
			ChrtID:     item.ChrtID,
			Price:      item.Price,
			Rid:        item.Rid,
			Name:       item.Name,
			Sale:       item.Sale,
			Size:       item.Size,
			TotalPrice: item.TotalPrice,
			NmID:       item.NmID,
			Brand:      item.Brand,
		}
		_, err := m.DB.Exec(stmt, insertItem.OrderUID, insertItem.ChrtID, insertItem.Price, insertItem.Rid, insertItem.Name, insertItem.Sale, insertItem.Size, insertItem.TotalPrice, insertItem.NmID, insertItem.Brand)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *DbModel) InsertNewOrder(order models.OrderGet) error {

	m.RLock()
	defer m.RUnlock()

	orderGet := models.OrderGet{
		OrderUID:          order.OrderUID,
		Entry:             order.Entry,
		InternalSignature: order.InternalSignature,
		Payment:           order.Payment,
		Locale:            order.Locale,
		CustomerID:        order.CustomerID,
		TrackNumber:       order.TrackNumber,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmID:              order.SmID,
	}

	m.Add(2)

	go func() {

		defer m.Done()

		OrderCache.SetCacheOrderGet(orderGet.OrderUID, orderGet, 5*time.Minute)
	}()

	go func() {

		defer m.Done()

		stmt := "SELECT insertneworder ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"

		_, err := m.DB.Exec(stmt, orderGet.OrderUID, orderGet.Entry, orderGet.InternalSignature, orderGet.Payment.Transaction, orderGet.Locale, orderGet.CustomerID, orderGet.TrackNumber, orderGet.DeliveryService, orderGet.Shardkey, orderGet.SmID)
		if err != nil {

			log.Println(err)

			m.DB.Exec(stmt, OrderCache.GetCacheOrderGet(orderGet.OrderUID).OrderUID,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).Entry,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).InternalSignature,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).Payment.Transaction,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).Locale,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).CustomerID,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).TrackNumber,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).DeliveryService,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).Shardkey,
				OrderCache.GetCacheOrderGet(orderGet.OrderUID).SmID)
		}
	}()

	m.Wait()

	return nil
}
