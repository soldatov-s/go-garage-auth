-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION production.loginFastSearch(IN login varchar(255)) RETURNS SETOF production."user" PARALLEL SAFE AS $$
DECLARE tmp_login_hash_table varchar(255);
	tmp_user_id bigint;
	tmp_hash bigint;
	login_hash_table varchar(255);
BEGIN 
	tmp_hash := (abs(hashtext(login)) % 1000);
	login_hash_table := 'login_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
    	'SELECT user_id FROM production.%I WHERE user_login=$1',
        login_hash_table
    ) INTO tmp_user_id USING login;
    RETURN QUERY SELECT * FROM production."user"
    	WHERE user_id = tmp_user_id;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS insert_login_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.insertlogin() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    login_hash_table varchar(255);
	last_id bigint;
BEGIN 
	tmp_hash := (abs(hashtext(NEW.user_login)) % 1000);
    login_hash_table := 'login_hash_' || tmp_hash::varchar(255);
	last_id := (SELECT last_value FROM production.user_user_id_seq); -- "user_user_id_seq" it isn't mistake!
	EXECUTE format(
    	'INSERT INTO production.%I (user_id, user_login) VALUES ($1, $2)',
    	login_hash_table
	) USING NEW.user_id, NEW.user_login;
	RETURN NULL;
EXCEPTION
	WHEN unique_violation THEN
		PERFORM setval('production.user_user_id_seq', last_id - 1);
		RAISE;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_login_in_hash
AFTER
INSERT ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.insertlogin();

DROP TRIGGER IF EXISTS delete_login_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.deletelogin() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    login_hash_table varchar(255);
BEGIN 
	tmp_hash := (abs(hashtext(OLD.user_login)) % 1000);
	login_hash_table := 'login_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
		'DELETE FROM production.%I WHERE user_login=$1',
		login_hash_table
	) USING OLD.user_login;
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER delete_login_in_hash
AFTER DELETE ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.deletelogin();

DROP TRIGGER IF EXISTS update_login_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.updatelogin() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    login_hash_table varchar(255);
BEGIN 
	IF NEW.user_login = OLD.user_login THEN 
		RETURN NULL;
	END IF;
	tmp_hash := (abs(hashtext(NEW.user_login)) % 1000);
	login_hash_table := 'login_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
    	'INSERT INTO production.%I (user_id, user_login) VALUES ($1, $2)',
    	login_hash_table
	) USING OLD.user_id, NEW.user_login;
	tmp_hash := (abs(hashtext(OLD.user_login)) % 1000);
	login_hash_table := 'login_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
		'DELETE FROM production.%I WHERE user_login=$1',
		login_hash_table
	) USING OLD.user_login;
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER update_login_in_hash
AFTER UPDATE ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.updatelogin();

DO
$do$
DECLARE
   counter int = 0;
   login_hash_table varchar(255);
BEGIN
LOOP
	login_hash_table := 'login_hash_' || counter::varchar(255);
	EXECUTE format(
		'CREATE TABLE IF NOT EXISTS production.%I  (
				user_id BIGINT PRIMARY KEY,
				user_login character varying(255),
				CONSTRAINT uniq_%I UNIQUE (user_login)
			)',
			login_hash_table,
			login_hash_table
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
DROP TRIGGER IF EXISTS insert_login_in_hash on production."user";
DROP TRIGGER IF EXISTS delete_login_in_hash on production."user";
DROP FUNCTION IF EXISTS production.loginFastSearch;
