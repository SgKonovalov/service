package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	"my.service.show/pkg/models"
)

/*
Для работы http-сервера, используем 2 хендлера:
1) Home – отображает «стартовую» страницу при запросе по URL "/".
Для UI в данном хендлере, используем срез строк – путь к html-страницам:
	1.1) home.page.html – сама «стартовая» страница;
	1.2) base.layout.html – шаблон страниц для всего сервера.
2) ShowOrder – отображает страницу с найденнм заказом:
	2.1) Считываем ID, введённый пользователем;
	2.2) Публикуем его в Nats Streaming (отправляем микросервису «query» для поиска нужного заказа);
	2.3) Обрабатываем html-страницу: serchbyid.page.html, для вывода информации о заказе;
	2.4) Посредством Nats Streaming, получаем от микросервиса «query» ответ в виде объекта JSON;
	2.5) Переводим декодируем объект JSON в объект типа models.OrderPost;
	2.6) В случае если получили заполненный объект – выводим данные в виде таблицы.
	2.7) В случае, если получили «пустой» объект –
	выводим на экран информацию о неверно введённом ID, пользователем.
*/

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.NotFound(w)
		return
	}

	files := []string{
		"./ui/html/home.page.html",
		"./ui/html/base.layout.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		app.ServerError(w, err)
	}
}

func (app *Application) ShowOrder(w http.ResponseWriter, r *http.Request) {

	searched, ok := r.URL.Query()["id"]
	if !ok || len(searched[0]) < 1 {
		app.NotFound(w)

	}
	orderId := searched[0]
	app.PubishID(orderId)

	files := []string{
		"./ui/html/serchbyid.page.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	order := app.getSearchedOrder()

	showAtUI := models.OrderPost{}
	_ = json.Unmarshal(order, &showAtUI)

	if *&showAtUI.OrderUID != "" {

		err = ts.ExecuteTemplate(w, "order", showAtUI)
		if err != nil {
			app.ServerError(w, err)
		}
	} else {
		showAtUI.OrderUID = "Заказа с указаным ID не существует!"
		showAtUI.Entry = "У несуществующего заказа - нет продавца"
		showAtUI.CustomerID = "ID Клиента не указан"
		showAtUI.TrackNumber = "Невозможно отследить несуществующий заказ"
		showAtUI.DeliveryService = "Этот заказ никто не доставляет"
		err = ts.ExecuteTemplate(w, "order", showAtUI)
		if err != nil {
			app.ServerError(w, err)
		}

	}
}
