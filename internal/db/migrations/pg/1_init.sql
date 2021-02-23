-- +goose Up
CREATE TABLE IF NOT EXISTS production."user" (
    user_id BIGSERIAL,
    user_hash text,
    user_login character varying(255),
    user_email character varying(255),
    user_phone character varying(255),
    user_status character varying(255),
    user_role character varying(255),
    user_meta jsonb,
    user_activation_hash text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
) PARTITION BY RANGE (user_id);

CREATE TABLE production."user_1_100000" PARTITION OF production."user" FOR
VALUES FROM (1) TO (100001);
CREATE INDEX user_1_100000_user_id ON production."user_1_100000" (user_id);

CREATE TABLE production."user_100001_200000" PARTITION OF production."user" FOR
VALUES FROM (100001) TO (200001);
CREATE INDEX user_100001_200000_user_id ON production."user_100001_200000" (user_id);

CREATE TABLE production."user_200001_300000" PARTITION OF production."user" FOR
VALUES FROM (200001) TO (300001);
CREATE INDEX user_200001_300000_user_id ON production."user_200001_300000" (user_id);

CREATE TABLE production."user_300001_400000" PARTITION OF production."user" FOR
VALUES FROM (300001) TO (400001);
CREATE INDEX user_300001_400000_user_id ON production."user_300001_400000" (user_id);

CREATE TABLE production."user_400001_500000" PARTITION OF production."user" FOR
VALUES FROM (400001) TO (500001);
CREATE INDEX user_400001_500000_user_id ON production."user_400001_500000" (user_id);

CREATE TABLE production."user_500001_600000" PARTITION OF production."user" FOR
VALUES FROM (500001) TO (600001);
CREATE INDEX user_500001_600000_user_id ON production."user_500001_600000" (user_id);

CREATE TABLE production."user_600001_700000" PARTITION OF production."user" FOR
VALUES FROM (600001) TO (700001);
CREATE INDEX user_600001_700000_user_id ON production."user_600001_700000" (user_id);

CREATE TABLE production."user_700001_800000" PARTITION OF production."user" FOR
VALUES FROM (700001) TO (800001);
CREATE INDEX user_700001_800000_user_id ON production."user_700001_800000" (user_id);

CREATE TABLE production."user_800001_900000" PARTITION OF production."user" FOR
VALUES FROM (800001) TO (900001);
CREATE INDEX user_800001_900000_user_id ON production."user_800001_900000" (user_id);

CREATE TABLE production."user_900001_1000000" PARTITION OF production."user" FOR
VALUES FROM (900001) TO (1000001);
CREATE INDEX user_900001_1000000_user_id ON production."user_900001_1000000" (user_id);

-- +goose Down
DROP TABLE IF EXISTS production."user";