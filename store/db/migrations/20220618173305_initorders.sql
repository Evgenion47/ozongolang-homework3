-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
create table Orders(
    "IdOrder" bigint PRIMARY KEY,
    "IdUser" bigint NOT NULL,
    "State" bool
);

create table Goods(
    "IdGoods" VARCHAR(255) PRIMARY KEY,
    "Cost" integer,
    "AmountOnWH" integer
);

create table Dict(
    "IdOrder" bigint REFERENCES Orders,
    "IdGoods" VARCHAR(255) REFERENCES Goods,
    "Amount" integer
)/*PARTITION BY HASH ("IdOrder")*/;

-- CREATE EXTENSION if not exists postgres_fdw;
--
-- CREATE SERVER shard1 FOREIGN DATA WRAPPER postgres_fdw OPTIONS ( host '127.17.0.2', port '5432', dbname 'orders' );
-- CREATE SERVER shard2 FOREIGN DATA WRAPPER postgres_fdw OPTIONS ( host '127.17.0.3', port '5432', dbname 'orders' );
--
-- CREATE FOREIGN TABLE Dict_1 partition of Dict for values with (modulus 2, remainder 0) server shard1;
-- CREATE FOREIGN TABLE Dict_2 partition of Dict for values with (modulus 2, remainder 1) server shard2;
--
-- CREATE USER MAPPING FOR "postgres" SERVER shard1 OPTIONS (user 'postgres', password 'postgres');
-- CREATE USER MAPPING FOR "postgres" SERVER shard2 OPTIONS (user 'postgres', password 'postgres');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
drop table if exists Dict;
drop table if exists Orders;
drop table if exists Goods;
-- drop server if exists shard1;
-- drop server if exists shard2;
-- +goose StatementEnd
