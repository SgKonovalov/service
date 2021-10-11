package postgresql

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"my.service.query/pkg/models"
	"my.service.query/pkg/models/postgresql/cache"
)

/*
Структура DbModel - управляет доступом к БД.

Индивидуальные функции, использующиеся для добавления данных в каждую таблицу SQL.

1) Функция GetOrderByID принимает в качестве аргумента ID заказа,
указанное пользователем для поиска.
В результате генерирует структуру из БД для последующей выдачи её по HTTP. Для этого:
	1.1) Делает запрос в БД для формирования новой строки в таблицу с хранящимися в ней результатами запроса,
	осуществляя поиск по ID заказа;
	1.2) Если запись уже есть в БД – выдаёт существующую;
	1.3) Записывает полученные данные в кэш.
Таблица: order_post, модель: OrderPost.

2) Функция GetOriginOrder
принимает в качестве аргументов:
	2.1.1) Указатель на строку (ID заказа, указанное пользователем);
	2.1.2) Канал типа models.OrderPost для записи в него результата;
По окончании работы, возвращает канал с нужной структурой типа models.OrderPost
для последующей выдачи его данных по http. Порядок работы функции:
	2.2.1) Используя аргумент из п. 2.1.1 – получаем данные из кэша;
	2.2.2) В случае возникновения ошибки, вызываем функцию GetOrderByID;
	2.2.3) И в том и в другом случае записываем полученный результат в канал и возвращаем его в return.

ВАЖНО: вся работа по добавлению и поиску данных в SQL осуществляется на стороне БД посредством хранимых процедур.
Из приложения достаточно вызвать нужную функцию и передать ей необходимые параметры. Такой подход «избавляет»
основной процесс от обработки данных и инкапсулирует работу в БД, непосредственно в самой БД.

Хранимая процедура:
insertintoorderpost – добавляет данные в таблицу order_post.
Таблица order_post – хранит информацию о заказах, которые искали пользователи;

Подробный код хранимых процедур и скрипты таблиц в файле: scripts.sql.
*/

type DbModel struct {
	DB *sql.DB
	sync.RWMutex
}

var OrderCache = cache.NewCacheOrderPost(5*time.Minute, 10*time.Minute)

func (m *DbModel) GetOrderByID(orderId string) (result models.OrderPost, err error) {

	create := "SELECT insertintoorderpost ($1)"

	_, err = m.DB.Exec(create, orderId)
	if err != nil {
		goto showingExisting
	}

showingExisting:
	query := "SELECT order_uid, entry, total_price, customer_id, track_number, delivery_service FROM order_post WHERE order_uid = $1"

	row, err := m.DB.Query(query, orderId)
	if err != nil {
		return result, err
	}

	for row.Next() {
		err = row.Scan(&result.OrderUID, &result.Entry, &result.TotalPrice, &result.CustomerID, &result.TrackNumber, &result.DeliveryService)
	}
	row.Close()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, err
		} else {
			return result, err
		}
	}

	OrderCache.SetCacheOrderPost(result.OrderUID, result, 5*time.Minute)

	return result, nil
}

func (m *DbModel) GetOriginOrder(orderId *string, ChanForResult chan models.OrderPost) chan models.OrderPost {

	result, ok := OrderCache.GetCacheOrderPost(*orderId)

	if !ok {
		result, err := m.GetOrderByID(*orderId)
		if err != nil {
			return nil
		}
		ChanForResult <- result
		return ChanForResult
	}

	ChanForResult <- result
	return ChanForResult
}
