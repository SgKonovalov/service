package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

/*
Основная часть программы:
1) Структура Application – основная структура программы, управляющая информацией о работе программы и ошибках
2) В функции main:
	2.1) Устанавливаем счётчик WaitGroup;
	2.2) Указываем адрес веб-сервера;
	2.3) Задаём параметры для информировании о работе приложения и об ошибках в нём (infoLog, errorLog);
	2.4) Получаем функциональность приложения в части информирования о работе программы и сбоях в ней, создавая объект структуры  Application.
	2.5) Инициализируем http.Server, передавая:
		2.5.1) Адрес веб-сервера из п. 2.2;
		2.5.2) Лог ошибок из п. 2.3;
		2.5.3) В качестве хендлера – функцию Routes, отвечающую за маршрутизацию запросов.
	2.6) Подключаемся и обслуживаем созданный сервер.
*/

type Application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

var Wg sync.WaitGroup

func main() {
	Wg.Add(2)
	addr := flag.String("addr", ":8080", "Сетевой адрес веб-сервера")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &Application{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.Routes(),
	}

	infoLog.Printf("Запуск сервера на %s", *addr)

	err := srv.ListenAndServe()
	errorLog.Fatal(err)
}
