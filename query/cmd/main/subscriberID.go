package main

import (
	"log"
	"sync"

	nats "github.com/nats-io/nats.go"
)

/*
Функция getSearchedID принимает в качестве аргумента канал типа string для записи в него значения полученного ID заказа, который «ищет» пользователь.
В результате работы возвращает канал с записанным в нём ID.
Процесс работы функции:
1) В качестве слушателя, соединяемся с микросервисом «show»;
2) Получаем от последнего массив байт (ID заказа, введённый пользователем);
3) Переводим полученное значение в строку и записываем в канал;
4) В return, возвращаем канал из п. 3.

Использование NATS вместо NATS streaming обусловлено наличием связи между микросервисами (show и query)
в формате 1:1, а так же их совместной работой в режиме «запрос – ответ» и выдачей только 1-го результата =>
потери сообщений исключены.
*/

func (app *application) getSearchedID(ChanForID chan string) chan string {

	var ID string

	nc, err := nats.Connect("demo.nats.io")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	func(string) {
		if _, err := nc.Subscribe("IDSend", func(m *nats.Msg) {

			ID = string(m.Data)
			defer wg.Done()
		}); err != nil {
			log.Fatal(err)
		}
	}(ID)

	wg.Wait()
	ChanForID <- ID
	return ChanForID
}
