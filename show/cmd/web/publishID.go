package main

import (
	"log"

	nats "github.com/nats-io/nats.go"
)

/*
Функция PubishID принимает в качестве аргумента строку - сам ID, указанный пользователем.
Возвращает ошибку (при наличии).
Процесс работы функции:
1) В качестве publisher, соединяемся с микросервисом «query»;
2) Публикуем полученную строку в виде среза byte;
3) В return, возвращаем ошибку, если она возникла при публикации.

Использование NATS вместо NATS streaming обусловлено наличием связи между микросервисами (show и query)
в формате 1:1, а так же их совместной работой в режиме «запрос – ответ» и выдачей только 1-го результата =>
потери сообщений исключены.
*/

func (app *Application) PubishID(ID string) error {

	nc, err := nats.Connect("demo.nats.io")
	if err != nil {
		log.Fatal(err)
	}

	defer nc.Close()

	if err := nc.Publish("IDSend", []byte(ID)); err != nil {
		log.Fatal(err)
	}
	return err
}
