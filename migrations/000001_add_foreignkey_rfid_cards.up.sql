ALTER TABLE rfid_cards DROP COLUMN owner_name,
    ADD COLUMN user_id BIGINT UNSIGNED NULL,
    ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id);