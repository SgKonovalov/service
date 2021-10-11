package main

import "net/http"

/*
Функция Routes используется для маршрутизации и обработки запросов пользователя:
1) mux.HandleFunc("/", app.Home) – при запросе по URL «http://localhost:8080/» или «http://localhost:8080»,
«отправляет» пользователя на «стартовую страницу»;
2) mux.HandleFunc("/order", app.ShowOrder) – при запросе по URL «http://localhost:8080/order?id=orderid»,
«отправляет» пользователя на страницу с отображёнными сведениями о заказе, ID которого указал последний.
3) fileServer := http.FileServer(http.Dir("./ui/static/"))
mux.Handle("/static/", http.StripPrefix("/static", fileServer)) – исключает доступ в к статичный файлам,
находящимся на сервере, при использовании маршрутизации.
*/

func (app *Application) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/order", app.ShowOrder)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	return mux
}
