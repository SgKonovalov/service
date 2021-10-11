package cache

import (
	"fmt"
	"sync"
	"time"

	"my.service.query/pkg/models"
)

/*
Реализация in memory кэша, где:
1) Структура ItemOrderPost – это сам объект, который помещаем в кэш, состоит из:
	1.1) Value – сам объект типа models.OrderPost;
	1.2) Created – дата создания (сохранения в кэш);
	1.3) Expiration – истечение срока хранения элемента.
2) Структура CacheOrderPost – сам кэш. Состоит из:
	2.1) sync.RWMutex – зависимость, для возможности использования Mutex’ов;
	2.2) defaultExpiration – срок хранения данных в кэше по умолчанию;
	2.3) cleanUpInterval – время, по истечении которого кэш будет очищен;
	2.4) Item – сами данные, типизированные структурой ItemOrderPost.
3) Конструктор NewCacheOrderPost – создаёт новый кэш;
4) Функция SetCacheOrderPost – добавляет в кэш новое значение, принимая в качестве параметра:
	4.1) key string – ID заказа. Этот ключ уникален, по его значению будет производится поиск и обработка значений;
	4.2) value models.OrderPost – сам объект типа models.OrderPost, который необходимо хранить в кэше;
	4.3) duration time.Duration – устанавливаем время, в течении которого будем хранить данные.
5) Функция GetCacheOrderPost – принимает в качетве аргумента ключ или ID заказа и «выдаёт» нужный объект
типа models.OrderPost.
6) Функция DeleteCacheOrderPost принимает в качетве аргумента ключ или ID заказа и удаляет из кэша
нужный объект типа models.OrderGet.
7) Функция expiredKeysCacheOrderPost – принимает срез типа ключей (ID) заказов для их удаления из кэша.
ВАЖНО: удаляются только объекты с истекшим сроком хранения.
8) Функция clearItemsCacheOrderPost – принимает срез типа ключей (ID) заказов для их удаления из кэша. ВАЖНО: удаляются все объекты.
9) Функции GCCacheOrderPost и StartGCCacheOrderPost – «сборщик» мусора для кэша.
Автоматически удаляет объекты с истекшим сроком хранения.
*/

type ItemOrderPost struct {
	Value      models.OrderPost
	Created    time.Time
	Expiration int64
}

type CacheOrderPost struct {
	sync.RWMutex
	defaultExpiration time.Duration
	cleanUpInterval   time.Duration
	Item              map[string]ItemOrderPost
}

func NewCacheOrderPost(defaultExpiration, cleanUpInterval time.Duration) *CacheOrderPost {

	items := make(map[string]ItemOrderPost)

	cache := CacheOrderPost{
		defaultExpiration: defaultExpiration,
		cleanUpInterval:   cleanUpInterval,
		Item:              items,
	}
	if cleanUpInterval > 0 {
		cache.StartGCCacheOrderPost()
	}
	return &cache
}

func (c *CacheOrderPost) SetCacheOrderPost(key string, value models.OrderPost, duration time.Duration) {

	var expiration float64

	if duration == 0 {
		duration = c.defaultExpiration
	}
	if duration > 0 {
		expiration = float64(time.Now().Add(duration).UnixNano())
	}
	c.Lock()
	defer c.Unlock()
	c.Item[key] = ItemOrderPost{
		Value:      value,
		Created:    time.Now(),
		Expiration: int64(expiration),
	}
}

func (c *CacheOrderPost) GetCacheOrderPost(key string) (order models.OrderPost, ok bool) {
	c.RLock()
	defer c.RUnlock()

	item, found := c.Item[key]
	if !found {
		return order, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			return order, false
		}
	}
	return item.Value, true
}

func (c *CacheOrderPost) DeleteCacheOrderPost(key string) error {
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

func (c *CacheOrderPost) expiredKeysCacheOrderPost() (keys []string) {

	c.RLock()
	defer c.RUnlock()

	for k, i := range c.Item {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}

	return
}

func (c *CacheOrderPost) clearItemsCacheOrderPost(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		delete(c.Item, k)
	}
}

func (c *CacheOrderPost) GCCacheOrderPost() {
	for {
		<-time.After(c.cleanUpInterval)
		if c.Item == nil {
			return
		}
		if keys := c.expiredKeysCacheOrderPost(); len(keys) != 0 {
			c.clearItemsCacheOrderPost(keys)
		}
	}
}

func (c *CacheOrderPost) StartGCCacheOrderPost() {
	go c.GCCacheOrderPost()
}
