-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION production.emailFastSearch(IN email varchar(255)) RETURNS SETOF production."user" PARALLEL SAFE AS $$
DECLARE tmp_email_hash_table varchar(255);
	tmp_user_id bigint;
	tmp_hash bigint;
	email_hash_table varchar(255);
BEGIN 
	tmp_hash := (abs(hashtext(email)) % 1000);
	email_hash_table := 'email_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
    	'SELECT user_id FROM production.%I WHERE user_email=$1',
        email_hash_table
    ) INTO tmp_user_id USING email;
    RETURN QUERY SELECT * FROM production."user"
    	WHERE user_id = tmp_user_id;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS insert_email_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.insertemail() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    email_hash_table varchar(255);
	last_id bigint;
BEGIN 
	tmp_hash := (abs(hashtext(NEW.user_email)) % 1000);
    email_hash_table := 'email_hash_' || tmp_hash::varchar(255);
	last_id := (SELECT last_value FROM production.user_user_id_seq); -- "user_user_id_seq" it isn't mistake!
	EXECUTE format(
    	'INSERT INTO production.%I (user_id, user_email) VALUES ($1, $2)',
    	email_hash_table
	) USING NEW.user_id, NEW.user_email;
	RETURN NULL;
EXCEPTION
	WHEN unique_violation THEN
		PERFORM setval('production.user_user_id_seq', last_id - 1);
		RAISE;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER insert_email_in_hash
AFTER
INSERT ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.insertemail();

DROP TRIGGER IF EXISTS delete_email_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.deleteemail() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    email_hash_table varchar(255);
BEGIN 
	tmp_hash := (abs(hashtext(OLD.user_email)) % 1000);
	email_hash_table := 'email_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
		'DELETE FROM production.%I WHERE user_email=$1',
		email_hash_table
	) USING OLD.user_email;
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER delete_email_in_hash
AFTER DELETE ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.deleteemail();

DROP TRIGGER IF EXISTS update_email_in_hash on production."user";
CREATE OR REPLACE FUNCTION production.updateemail() RETURNS TRIGGER AS $$
DECLARE 
    tmp_hash bigint;
    email_hash_table varchar(255);
BEGIN 
	IF NEW.user_email = OLD.user_email THEN 
		RETURN NULL;
	END IF;
	tmp_hash := (abs(hashtext(NEW.user_email)) % 1000);
	email_hash_table := 'email_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
    	'INSERT INTO production.%I (user_id, user_email) VALUES ($1, $2)',
    	email_hash_table
	) USING OLD.user_id, NEW.user_email;
	tmp_hash := (abs(hashtext(OLD.user_email)) % 1000);
	email_hash_table := 'email_hash_' || tmp_hash::varchar(255);
	EXECUTE format(
		'DELETE FROM production.%I WHERE user_email=$1',
		email_hash_table
	) USING OLD.user_email;
	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER update_email_in_hash
AFTER UPDATE ON production."user" FOR EACH ROW EXECUTE PROCEDURE production.updateemail();

DO
$do$
DECLARE
   counter int = 0;
   email_hash_table varchar(255);
BEGIN
LOOP
	email_hash_table := 'email_hash_' || counter::varchar(255);
	EXECUTE format(
		'CREATE TABLE IF NOT EXISTS production.%I  (
				user_id BIGINT PRIMARY KEY,
				user_email character varying(255),
				CONSTRAINT uniq_%I UNIQUE (user_email)
			)',
			email_hash_table,
			email_hash_table
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
DROP TRIGGER IF EXISTS insert_email_in_hash on production."user";
DROP TRIGGER IF EXISTS delete_email_in_hash on production."user";
DROP FUNCTION IF EXISTS production.emailFastSearch;
