-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION production.phoneFastSearch(IN phone varchar(255)) RETURNS SETOF production."user" PARALLEL SAFE AS $$
DECLARE tmp_phone_hash_table varchar(255);
	tmp_user_id bigint;
	tmp_hash bigint;
	phone_hash_table varchar(255);
BEGIN 
	tmp_hash := (abs(hashtext(phone)) % 1000);
	phone_hash_table := 'phone_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
    	'SELECT user_id FROM production.%I WHERE user_phone=$1',
        phone_hash_table
    ) INTO tmp_user_id USING phone;
    RETURN QUERY SELECT * FROM production."user"
    	WHERE user_id = tmp_user_id;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS insert_phone_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.insertphone() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    phone_hash_table varchar(255);
	last_id bigint;
BEGIN 
	tmp_hash := (abs(hashtext(NEW.user_phone)) % 1000);
    phone_hash_table := 'phone_hash_' || tmp_hash::varchar(255);
	last_id := (SELECT last_value FROM production.user_user_id_seq); -- "user_user_id_seq" it isn't mistake!
	EXECUTE format(
    	'INSERT INTO production.%I (user_id, user_phone) VALUES ($1, $2)',
    	phone_hash_table
	) USING NEW.user_id, NEW.user_phone;
	RETURN NULL;
EXCEPTION
	WHEN unique_violation THEN
		PERFORM setval('production.user_user_id_seq', last_id - 1);
		RAISE;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_phone_in_hash
AFTER
INSERT ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.insertphone();

DROP TRIGGER IF EXISTS delete_phone_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.deletephone() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    phone_hash_table varchar(255);
BEGIN 
	tmp_hash := (abs(hashtext(OLD.user_phone)) % 1000);
	phone_hash_table := 'phone_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
		'DELETE FROM production.%I WHERE user_phone=$1',
		phone_hash_table
	) USING OLD.user_phone;
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER delete_phone_in_hash
AFTER DELETE ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.deletephone();

DROP TRIGGER IF EXISTS update_phone_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.updatephone() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    phone_hash_table varchar(255);
BEGIN 
	IF NEW.user_phone = OLD.user_phone THEN 
		RETURN NULL;
	END IF;
	tmp_hash := (abs(hashtext(NEW.user_phone)) % 1000);
	phone_hash_table := 'phone_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
    	'INSERT INTO production.%I (user_id, user_phone) VALUES ($1, $2)',
    	phone_hash_table
	) USING OLD.user_id, NEW.user_phone;
	tmp_hash := (abs(hashtext(OLD.user_phone)) % 1000);
	phone_hash_table := 'phone_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
		'DELETE FROM production.%I WHERE user_phone=$1',
		phone_hash_table
	) USING OLD.user_phone;
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER update_phone_in_hash
AFTER UPDATE ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.updatephone();

DO
$do$
DECLARE
   counter int = 0;
   phone_hash_table varchar(255);
BEGIN
LOOP
	phone_hash_table := 'phone_hash_' || counter::varchar(255);
	EXECUTE format(
		'CREATE TABLE IF NOT EXISTS production.%I  (
				user_id BIGINT PRIMARY KEY,
				user_phone character varying(255),
				CONSTRAINT uniq_%I UNIQUE (user_phone)
			)',
			phone_hash_table,
			phone_hash_table
		);
	counter := counter + 1;
    IF counter > 999 THEN
        EXIT;
    END IF;
END LOOP;
END;
$do$;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS insert_phone_in_hash on production."user";
DROP TRIGGER IF EXISTS delete_phone_in_hash on production."user";
DROP FUNCTION IF EXISTS production.phoneFastSearch;
