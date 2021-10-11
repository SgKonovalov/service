// Скрипт таблицы items
CREATE TABLE items (
    order_uid VARCHAR,
    chrt_id INTEGER,
    price INTEGER,
    rid VARCHAR,
    name VARCHAR,
    sale INTEGER,
    size VARCHAR,
    total_price INTEGER,
    nmID INTEGER,
    brand VARCHAR,
    FOREIGN KEY (order_uid) REFERENCES payment (order_uid) ON DELETE CASCADE
);

// Скрипт таблицы payment
CREATE TABLE payment (
    order_uid VARCHAR PRIMARY KEY,
    transaction VARCHAR,
    currency VARCHAR,
    provider VARCHAR,
    amount INTEGER,
    payment_dt INTEGER,
    bank VARCHAR,
    deliveryCost INTEGER
);

// Скрипт таблицы order_get. Для сохранения данных, полученных ч/з NATS Streaming.
CREATE TABLE order_get (
    order_uid VARCHAR PRIMARY KEY,
    entry VARCHAR,
    internal_signature VARCHAR,
    payment VARCHAR,
    items integer[],
    locale VARCHAR,
    customer_id VARCHAR,
    track_number VARCHAR,
    delivery_service VARCHAR,
    shardkey VARCHAR,
    sm_id INTEGER,
    FOREIGN KEY (order_uid) REFERENCES payment (order_uid) ON DELETE CASCADE
);

// Скрипт таблицы order_post. Для отправления данных клиенту по HTTP.
CREATE TABLE order_post (
    order_uid VARCHAR PRIMARY KEY,
    entry VARCHAR,
    total_price BIGINT,
    customer_id VARCHAR,
    track_number VARCHAR,
    delivery_service VARCHAR,
    FOREIGN KEY (order_uid) REFERENCES order_get (order_uid) ON DELETE CASCADE
);


// Хранимая процедура для добавления нового значения в таблицу items
CREATE OR REPLACE FUNCTION insertnewitem (
    order_uid VARCHAR,
    chrt_id int,
    price int,
    rid varchar,
    name varchar,
    sale int,
    size varchar,
    total_price int,
    nmID int,
    brand varchar
    )
RETURNS VOID AS $$
BEGIN
INSERT INTO items (order_uid, chrt_id, price, rid, name, sale, size, total_price, nmID, brand)
VALUES (order_uid, chrt_id, price, rid, name, sale, size, total_price, nmID, brand);
END;
$$ LANGUAGE plpgsql;

// Хранимая процедура для добавления нового значения в таблицу payment
CREATE OR REPLACE FUNCTION insertnewpayment (
    order_uid VARCHAR,
    transaction varchar,
    currency varchar,
    provider varchar,
    amount int,
    payment_dt int,
    bank varchar,
    deliveryCost int
    )
RETURNS VOID AS $$
BEGIN
INSERT INTO payment (order_uid, transaction, currency, provider, amount, payment_dt, bank, deliveryCost)
VALUES (order_uid, transaction, currency, provider, amount, payment_dt, bank, deliveryCost);
END;
$$ LANGUAGE plpgsql;

// Хранимая процедура для добавления нового значения в таблицу order_get
CREATE OR REPLACE FUNCTION insertneworder (
    order_id varchar,
    entry varchar,
    internal_signature varchar,
    transaction_id varchar,
    locale varchar,
    customer_id varchar,
    track_number varchar,
    delivery_service varchar,
    shardkey varchar,
    sm_id int
    )
RETURNS VOID AS $$
BEGIN
INSERT INTO order_get (order_uid, entry, internal_signature, payment, items, locale, customer_id, track_number, delivery_service, shardkey, sm_id)
VALUES (order_id, entry, internal_signature, (SELECT transaction FROM payment WHERE order_uid = order_id),
(ARRAY (SELECT chrt_id FROM items AS i WHERE i.order_uid = order_id)), locale, customer_id, track_number, delivery_service, shardkey, sm_id);
END;
$$ LANGUAGE plpgsql;

// Хранимая процедура для добавления нового значения в таблицу order_post
CREATE OR REPLACE FUNCTION insertintoorderpost (orid varchar)
RETURNS VOID AS $$
BEGIN
INSERT INTO order_post (order_uid, entry, total_price, customer_id, track_number, delivery_service)
VALUES (
(SELECT order_uid FROM order_get WHERE order_uid = orid), 
(SELECT entry FROM order_get WHERE order_uid = orid), 
((SELECT deliveryCost FROM payment WHERE order_uid = orid)
+
(SELECT SUM (total_price) FROM items WHERE order_uid = orid)),
(SELECT customer_id FROM order_get WHERE order_uid = orid), 
(SELECT track_number FROM order_get WHERE order_uid = orid),
(SELECT delivery_service FROM order_get WHERE order_uid = orid));
END;
$$ LANGUAGE plpgsql;

// Выборка данных из order_post
SELECT order_uid, entry, total_price, customer_id, track_number, delivery_service FROM
order_post WHERE order_uid = orderId;