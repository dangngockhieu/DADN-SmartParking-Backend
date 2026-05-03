ALTER TABLE rfid_cards DROP FOREIGN KEY fk_user_id,
    DROP COLUMN user_id,
    ADD COLUMN owner_name VARCHAR(255) NULL;