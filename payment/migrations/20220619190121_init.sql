-- +goose Up
-- +goose StatementBegin
--SELECT 'up SQL query';
create table payments
(
    IdUser Int64,
    IdOrder Int64,
    TotalCost int
)
ENGINE = MergeTree()
PRIMARY KEY (IdOrder);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
--SELECT 'down SQL query';
drop table if exists payments
-- +goose StatementEnd