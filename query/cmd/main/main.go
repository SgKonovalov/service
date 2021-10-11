package main

import (
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
	"my.service.query/pkg/models"
	"my.service.query/pkg/models/postgresql"
)

/*
Основная часть программы:
1) Структура Application – основная структура программы, управляющая информацией о работе программы и ошибках
+ управляет функциями добавления данных в БД.
2) Функция OpenDB – управляет подключением к БД. В качестве параметра принимает свойства для подключения к БД
и возвращает готовое, проверенное соединение с БД.
3) Каналы ChanForID и ChanForResult - для общения функций и хранения ID и модели соответственно;
4) Указатель на область памяти типа string (ID) - для хранения ID заказа, который ищет пользователь.
5) В функции main:
	5.1) Устанавливаем счётчик WaitGroup;
	5.2) Задаём параметры для информировании о работе приложения и об ошибках в нём (infoLog, errorLog);
	5.3) Получаем получение к БД, создавая объект структуры OpenDB;
	5.4) Получаем функциональность приложения в части возможности вызова функций для добавления информации в БД,
	создавая объект структуры  Application.
	5.5) Запускаем функцию getSearchedID и записываем результат её работы в канал (а это ID заказа, введённый
	пользователем) + присваиваем полученное значение указателю ID.
	5.6) С помощью ID, запускаем функцию GetOriginOrder и получаем нужный заказ;
	5.7) Публикуем этот заказ с помощью функции PubishOrder.
	В качестве параметра передаём созданный с помощью конструктора кэш.
*/

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	orderGet *postgresql.DbModel
}

var Wg sync.WaitGroup

var ChanForID = make(chan string, 1000)

var ChanForResult = make(chan models.OrderPost, 1000)

var ID = new(string)

func main() {
	Wg.Add(2)

	dsn := flag.String("dsn", "user=postgres password=postgres dbname=test sslmode=disable", "Название источника данных")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := OpenDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		orderGet: &postgresql.DbModel{DB: db},
	}

	infoLog.Printf("Запуск приложения. Выдача сведений о заказе при запросе с помощью ID.")

	for {

		asked := <-app.getSearchedID(ChanForID)
		*ID = asked

		searchedOrder := <-app.orderGet.GetOriginOrder(ID, ChanForResult)
		app.PubishOrder(searchedOrder)

	}

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
