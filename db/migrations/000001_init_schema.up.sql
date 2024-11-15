
CREATE TABLE "cipher_algorithm" (
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

  FOREIGN KEY(certificate_cryptographic_api_id) REFERENCES certificate_cryptographic_api(id),
  FOREIGN KEY(signing_request_api_id) REFERENCES signing_request_api(id),
  FOREIGN KEY(cipher_algorithm_id) REFERENCES cipher_algorithm(id)
);

INSERT INTO "certificate_cryptographic_api" (name, shortname) VALUES
('CryptoAPI', 'CAPI'),
('CryptoAPI Next-Generation', 'CNG');

INSERT INTO "signing_request_api" (name) VALUES
('Unknown'),
('PKCS10'),
('CMC');
