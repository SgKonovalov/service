package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"my.service.save/pkg/models/postgresql"
	"my.service.save/pkg/models/postgresql/cache"
)

/*
Основная часть программы:
1) Структура Application – основная структура программы, управляющая информацией о работе программы и ошибках
+ управляет функциями добавления данных в БД.
2) Функция OpenDB – управляет подключением к БД. В качестве параметра принимает свойства для подключения к БД
и возвращает готовое, проверенное соединение с БД.
3) В функции main:
	3.1) Устанавливаем счётчик WaitGroup;
	3.2) Задаём параметры для информировании о работе приложения и об ошибках в нём (infoLog, errorLog);
	3.3) Получаем получение к БД, создавая объект структуры OpenDB;
	3.4) Получаем функциональность приложения в части возможности вызова функций для добавления информации в БД,
	создавая объект структуры  Application.
	3.5) Запускаем саму функцию SubAndSave для сохранения данных в БД.
	В качестве параметра передаём созданный с помощью конструктора кэш.
*/

type Application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	orderGet *postgresql.DbModel
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

var Wg sync.WaitGroup

func main() {

	Wg.Add(1000)

	dsn := flag.String("dsn", "user=postgres password=postgres dbname=test sslmode=disable",
		"Название источника данных")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := OpenDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	app := &Application{
		errorLog: errorLog,
		infoLog:  infoLog,
		orderGet: &postgresql.DbModel{DB: db},
	}

	infoLog.Println("Запуск сервера приложения. Получение и обработка новых заказов.")

	for {
		app.SubAndSave(cache.NewCacheOrderGet(5*time.Minute, 10*time.Minute))
	}
}
