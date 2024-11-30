CREATE TABLE cipher_algorithm (
  id integer PRIMARY KEY,
  keysize integer,
  name varchar
);

CREATE TABLE signing_request_api (
  id integer PRIMARY KEY,
  name varchar
);

CREATE TABLE certificate_cryptographic_api (
  id integer PRIMARY KEY,
  name varchar,
  shortname varchar
);

CREATE TABLE hash_algorithm (
  id integer PRIMARY KEY,
  name varchar
);

CREATE TABLE certificate_requests (
  id integer PRIMARY KEY,
  display_name varchar,
  signing_algorithm varchar,
  key_length integer,
  requested_on timestamp,
  certificate_cryptographic_api_id integer,
  signing_request_api_id integer,
  cipher_algorithm_id integer,
  hash_algorithm_id integer,

  FOREIGN KEY(certificate_cryptographic_api_id) REFERENCES certificate_cryptographic_api(id),
  FOREIGN KEY(signing_request_api_id) REFERENCES signing_request_api(id),
  FOREIGN KEY(hash_algorithm_id) REFERENCES hash_algorithm(id),
  FOREIGN KEY(cipher_algorithm_id) REFERENCES cipher_algorithm(id)
);