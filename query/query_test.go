package main

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/lib/pq"

	"my.service.query/pkg/models"
	"my.service.query/pkg/models/postgresql"
)

/*
Тестирование функционала по подбору правильно значения из БД (функция GetOrderByID):
1) За основу берём корректную сущность типа models.OrderPost из БД;
2) Далее выполняем тесовое подключение к БД;
3) На основании этого подключения, делаем запрос из БД с помощью функции GetOrderByID, передавая её в качестве аргумента ID образца;
4) Сравниваем результаты: образец и полученный результат – если они сходятся: тест пройден.
*/

func TestGetOrderByID(t *testing.T) {
	var check = models.OrderPost{
		OrderUID:        "1q1",
		Entry:           "WBIL",
		TotalPrice:      7179,
		CustomerID:      "5ea488619943420eaefcbcc402eb8ddc",
		TrackNumber:     "WBIL2817015795SL",
		DeliveryService: "meest",
	}

	testSQLDB, err := sql.Open("postgres", "user=postgres password=1234 dbname=test sslmode=disable")

	if err != nil {
		t.Fatal(err)
	}
	defer testSQLDB.Close()

	testDB := postgresql.DbModel{
		DB: testSQLDB,
	}

	result, err := testDB.GetOrderByID("1q1")

	if err != nil {
		t.Fatalf("Error %v occurated.", err)
	}

	if !reflect.DeepEqual(check, result) {
		t.Fatalf("Incorrect parsing: %v and %v", check, result)
	}
}
