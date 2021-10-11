package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	stan "github.com/nats-io/stan.go"
	"my.service.save/pkg/models"
	"my.service.save/pkg/models/postgresql/cache"
)

/*
Функция SubAndSave принимает в качестве аргумента разыменновынный адрес области памяти с in memory кэшем.
Этот кэш будет использован, в случае сбоя при добавлении данных в БД.
Сама функция выполняет следующие действия:
1) Подключается к Nats Streaming в качестве слушателя и ассинхронно принимает от него сообщения
в виде среза byte или JSON-нотации;
2) Переводит срез байт в строку и проверяет наличие всех обязательных атрибутов структуры models.OrderGet.
В случае отсутствия хотя бы одного параметра, выходит из метода.
3) «Демаршалирует» полученные данные из JSON в структуру типа models.OrderGet;
4) Сохраняет эту структуру в кэш;
5) Вызывает фнукцию InsertAll, передавая в качестве параметра демаршалированный из JSON объект типа models.OrderGet.
В случае в работе функции InsertAll, вызывает её повторно, но уже отправляя данные из кэша.
В случае сбоя этого варианта – выходит из программы с log.Fatal.
*/

func (app *Application) SubAndSave(inMemoryCache *cache.CacheOrderGet) {

	var order models.OrderGet

	sc, err := stan.Connect("world-nats-stage", "SK", stan.NatsURL("wbx-world-nats-stage.dp.wb.ru"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	wg := sync.WaitGroup{}

	wg.Add(1000)

	if _, err := sc.Subscribe("go.test", func(m *stan.Msg) {

		conOrdeID := strings.Contains(string(m.Data), "order_uid")
		conEntry := strings.Contains(string(m.Data), "entry")
		conInSig := strings.Contains(string(m.Data), "internal_signature")
		conPayment := strings.Contains(string(m.Data), "payment")
		conItems := strings.Contains(string(m.Data), "items")
		conLocale := strings.Contains(string(m.Data), "locale")
		conCustID := strings.Contains(string(m.Data), "customer_id")
		conTNum := strings.Contains(string(m.Data), "track_number")
		conDService := strings.Contains(string(m.Data), "delivery_service")
		conShKey := strings.Contains(string(m.Data), "shardkey")

		if !conOrdeID || !conEntry || !conInSig || !conPayment || !conItems || !conLocale || !conCustID || !conTNum ||
			!conDService || !conShKey {
			fmt.Printf("Не добавлен %s\n", order.OrderUID)
			return
		}

		err := json.Unmarshal(m.Data, &order)

		if err != nil {
			log.Fatal(err)
		}

		inMemoryCache.SetCacheOrderGet(order.OrderUID, order, 5*time.Minute)
		errInsert := app.InsertAll(order)
		if errInsert != nil {
			errCache := app.InsertAll(inMemoryCache.GetCacheOrderGet(order.OrderUID))
			if errCache != nil {
				log.Fatal(errCache)
			}
			log.Fatal(err)
		}

		defer wg.Done()

	}); err != nil {
		log.Fatal(err)
	}

	wg.Wait()

}
