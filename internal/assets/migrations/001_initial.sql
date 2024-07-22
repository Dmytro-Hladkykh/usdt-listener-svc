-- +migrate Up
CREATE TABLE usdt_transfers (
    id BIGSERIAL PRIMARY KEY NOT NULL,
    from_address CHAR(42) NOT NULL,
    to_address CHAR(42) NOT NULL,
    amount NUMERIC NOT NULL,
    transaction_hash CHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

CREATE INDEX usdt_transfers_from_index ON usdt_transfers (from_address);
CREATE INDEX usdt_transfers_to_index ON usdt_transfers (to_address);
CREATE INDEX usdt_transfers_timestamp_index ON usdt_transfers (timestamp);

-- +migrate Down
DROP INDEX IF EXISTS usdt_transfers_timestamp_index;
DROP INDEX IF EXISTS usdt_transfers_to_index;
DROP INDEX IF EXISTS usdt_transfers_from_index;
DROP TABLE IF EXISTS usdt_transfers;