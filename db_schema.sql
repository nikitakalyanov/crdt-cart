CREATE DATABASE crdtcart;

\connect crdtcart

CREATE TABLE cart (id INTEGER, current_cart JSONB);

INSERT INTO cart VALUES (1, NULL);
