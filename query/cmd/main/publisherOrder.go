package main

import (
	"encoding/json"
	"log"

	nats "github.com/nats-io/nats.go"
	"my.service.query/pkg/models"
)

/*
Функция PubishOrder принимает в качестве аргумента объект типа models.OrderPost.
Возвращает ошибку (при наличии).
Процесс работы функции:
1) В качестве publisher, соединяемся с микросервисом «show»;
2) Кодируем полученный в качестве аргумента объект типа models.OrderPost в объект JSON;
3) Публикуем полученный JSON в Nats Streaming;
4) В return, возвращаем ошибку, если она возникла при публикации.

Использование NATS вместо NATS streaming обусловлено наличием связи между микросервисами (show и query)
в формате 1:1, а так же их совместной работой в режиме «запрос – ответ» и выдачей только 1-го результата =>
потери сообщений исключены.
*/

func (app *application) PubishOrder(order models.OrderPost) error {

	nc, err := nats.Connect("demo.nats.io")
	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()

	jsonResult, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}

	if err := nc.Publish("OrderSend", []byte(jsonResult)); err != nil {
		log.Fatal(err)
	}
	return err
}
