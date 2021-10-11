package main

import (
	"log"
	"sync"

	nats "github.com/nats-io/nats.go"
)

/*
Функция getSearchedOrder.
В результате работы возвращает срез байт – «неразобранный» объект JSON, полученный от микросервиса «query».
Процесс работы функции:
1) В качестве слушателя, соединяемся с микросервисом «query»;
2) Получаем от последнего срез байт (объект JSON, полученный от микросервиса «query»);
3) В return, возвращаем срез из п. 2.
*/

func (app *Application) getSearchedOrder() (order []byte) {

	nc, err := nats.Connect("demo.nats.io")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	func([]byte) {
		if _, err := nc.Subscribe("OrderSend", func(m *nats.Msg) {

			order = m.Data
			defer wg.Done()
		}); err != nil {
			log.Fatal(err)
		}
	}(order)

	wg.Wait()

	return order
}
