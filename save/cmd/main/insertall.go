package main

import (
	"fmt"

	"my.service.save/pkg/models"
)

/*
Функция InsertAll принимает в качестве аргумента объект типа OrderGet.
В процессе выполнения: поочередно вызывает функции добавления данных в БД.
Именно указанная очерёдность добавления данных - важное положение правильной работы
логики сохранения данных в БД, основанной на работе хранимых процедур в PostgreSQL.
*/

func (app *Application) InsertAll(order models.OrderGet) (err error) {

	defer Wg.Done()

	errAtInsPayment := app.orderGet.InsertNewPayment(order)
	if errAtInsPayment != nil {
		err = errAtInsPayment
	}
	errAtInsItems := app.orderGet.InsertNewItems(order)
	if errAtInsItems != nil {
		err = errAtInsItems
	}
	errAtInsOrder := app.orderGet.InsertNewOrder(order)
	if errAtInsOrder != nil {
		err = errAtInsOrder
	}
	fmt.Printf("Добавлен %s\n", order.OrderUID)
	return err
}
