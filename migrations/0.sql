CREATE TABLE schema_changelog (
  id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  version INTEGER NOT NULL UNIQUE,
  date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()
);

CREATE TABLE users (
    id INT GENERATED ALWAYS AS IDENTITY,
    email TEXT NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW()
);



-- last modified function
CREATE OR REPLACE FUNCTION update_last_modified() RETURNS TRIGGER AS $$
BEGIN
    IF(NEW != OLD) THEN
        NEW.last_modified := NOW();
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_last_modified() IS 'Updates the updated_at column with NOW() if any of the row''s columns have changed. Should be used in BEFORE UPDATE triggers.';




CREATE TRIGGER update_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE PROCEDURE update_last_modified();



INSERT INTO schema_changelog (version) VALUES (0);