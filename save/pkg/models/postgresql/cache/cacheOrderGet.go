package cache

import (
	"fmt"
	"sync"
	"time"

	"my.service.save/pkg/models"
)

/*
Реализация in memory кэша, где:
1) Структура ItemOrderGet – это сам объект, который помещаем в кэш, состоит из:
	1.1) Value – сам объект типа models.OrderGet;
	1.2) Created – дата создания (сохранения в кэш);
	1.3) Expiration – истечение срока хранения элемента.
2) Структура CacheOrderGet – сам кэш. Состоит из:
	2.1) sync.RWMutex – зависимость, для возможности использования Mutex’ов;
	2.2) defaultExpiration – срок хранения данных в кэше по умолчанию;
	2.3) cleanUpInterval – время, по истечении которого кэш будет очищен;
	2.4) Item – сами данные, типизированные структурой ItemOrderGet.
3) Конструктор NewCacheOrderGet – создаёт новый кэш;
4) Функция SetCacheOrderGet – добавляет в кэш новое значение, принимая в качестве параметра:
	4.1) key string – ID заказа. Этот ключ уникален, по его значению будет производится поиск и обработка значений;
	4.2) value models.OrderGet – сам объект типа models.OrderGet, который необходимо хранить в кэше;
	4.3) duration time.Duration – устанавливаем время, в течении которого будем хранить данные.
5) Функция GetCacheOrderGet – принимает в качетве аргумента ключ или ID заказа и «выдаёт» нужный объект типа models.OrderGet.
6) Функция DeleteCacheOrderGet– принимает в качетве аргумента ключ или ID заказа и удаляет из кэша нужный объект типа models.OrderGet.
7) Функция expiredKeysCacheOrderGet – принимает срез типа ключей (ID) заказов для их удаления из кэша. ВАЖНО: удаляются только объекты с истекшим сроком хранения.
8) Функция clearItemsCacheOrderGet – принимает срез типа ключей (ID) заказов для их удаления из кэша. ВАЖНО: удаляются все объекты.
9) Функции GCCacheOrderGet и StartGCCacheOrderGet – «сборщик» мусора для кэша. Автоматически удаляет объекты с истекшим сроком хранения.
*/

type ItemOrderGet struct {
	Value      models.OrderGet
	Created    time.Time
	Expiration int64
}

type CacheOrderGet struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanUpInterval   time.Duration
	Item              map[string]ItemOrderGet
}

func NewCacheOrderGet(defaultExpiration, cleanUpInterval time.Duration) *CacheOrderGet {

	items := make(map[string]ItemOrderGet)

	cache := CacheOrderGet{
		defaultExpiration: defaultExpiration,
		cleanUpInterval:   cleanUpInterval,
		Item:              items,
	}
	if cleanUpInterval > 0 {
		cache.StartGCCacheOrderGet()
	}
	return &cache
}

func (c *CacheOrderGet) SetCacheOrderGet(key string, value models.OrderGet, duration time.Duration) {

	var expiration float64

	if duration == 0 {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		expiration = float64(time.Now().Add(duration).UnixNano())
	}
	c.Lock()
	defer c.Unlock()
	c.Item[key] = ItemOrderGet{
		Value:      value,
		Created:    time.Now(),
		Expiration: int64(expiration),
	}
}

func (c *CacheOrderGet) GetCacheOrderGet(key string) (order models.OrderGet) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.Item[key]
	if !found {
		return order
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return order
		}
	}
	return item.Value
}

func (c *CacheOrderGet) DeleteCacheOrderGet(key string) error {
	c.Lock()
	defer c.Unlock()

	c.Lock()
	defer c.Unlock()

	if _, found := c.Item[key]; !found {
		return fmt.Errorf("Keys not found")
	}
	delete(c.Item, key)
	return nil
}

func (c *CacheOrderGet) expiredKeysCacheOrderGet() (keys []string) {

	c.RLock()
	defer c.RUnlock()

	for k, i := range c.Item {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}

	return
}

func (c *CacheOrderGet) clearItemsCacheOrderGet(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		delete(c.Item, k)
	}
}

func (c *CacheOrderGet) GCCacheOrderGet() {
	for {
		<-time.After(c.cleanUpInterval)
		if c.Item == nil {
			return
		}
		if keys := c.expiredKeysCacheOrderGet(); len(keys) != 0 {
			c.clearItemsCacheOrderGet(keys)
		}
	}
}

func (c *CacheOrderGet) StartGCCacheOrderGet() {
	go c.GCCacheOrderGet()
}
