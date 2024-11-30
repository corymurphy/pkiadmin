CREATE TABLE "cipher_algorithm" (
  "id" integer PRIMARY KEY,
  "keysize" integer,
  "name" varchar
);

CREATE TABLE "hash_algorithm" (
  "id" integer PRIMARY KEY,
  "name" varchar
);

CREATE TABLE "signing_request_api" (
  "id" integer PRIMARY KEY,
  "name" varchar
);

CREATE TABLE "certificate_cryptographic_api" (
  "id" integer PRIMARY KEY,
  "name" varchar,
  "shortname" varchar
);

CREATE TABLE "certificate_requests" (
  "id" integer PRIMARY KEY,
  "display_name" varchar,
  "signing_algorithm" varchar,
  "key_length" integer,
  "requested_on" timestamp,
  "certificate_cryptographic_api_id" integer,
  "signing_request_api_id" integer,
  "cipher_algorithm_id" integer,
  "hash_algorithm_id" integer,

  FOREIGN KEY(certificate_cryptographic_api_id) REFERENCES certificate_cryptographic_api(id),
  FOREIGN KEY(signing_request_api_id) REFERENCES signing_request_api(id),
  FOREIGN KEY(hash_algorithm_id) REFERENCES hash_algorithm(id),
  FOREIGN KEY(cipher_algorithm_id) REFERENCES cipher_algorithm(id)
);

INSERT INTO "certificate_cryptographic_api" (name, shortname) VALUES
('CryptoAPI Next-Generation (Recommended)', 'CNG'),
('CryptoAPI (Legacy)', 'CAPI');

INSERT INTO "signing_request_api" (name) VALUES
('unknown'),
('pkcs10'),
('cmc');

INSERT INTO "hash_algorithm" (name) VALUES
('sha1'),
('sha256'),
('sha384'),
('sha512');

INSERT INTO "cipher_algorithm" (name, keysize) VALUES
('rsa_2048', 2048),
('rsa_4096', 4096),
('ecdsa_p256', 256),
('ecdsa_p384', 384),
('ecdsa_p521', 521),
('ecdh_p256', 256),
('ecdh_p384', 384),
('ecdh_p521', 521);

INSERT INTO "certificate_requests" (display_name, key_length, hash_algorithm_id, cipher_algorithm_id, certificate_cryptographic_api_id, signing_request_api_id) VALUES
('example.com', 2048, 2, 1, 2, 2);