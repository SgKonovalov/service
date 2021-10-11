package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

/*
Содержимое файла helpers, использует для обработки ошибок, возникающих на строне сервера:
1) ServerError - записывает сообщение об ошибке в errorLog и отправляет пользователю ответ 500 "Внутренняя ошибка сервера";
2) ClientError - отправляет определенный код состояния и соответствующее описание пользователю.
Используется, если есть проблема с пользовательским запросом;
3) NotFound  - оболочка вокруг ClientError, которая отправляет пользователю ответ "404 Страница не найдена"
*/

func (app *Application) ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) NotFound(w http.ResponseWriter) {
	app.ClientError(w, http.StatusNotFound)
}
