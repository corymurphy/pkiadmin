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

CREATE TABLE "certificate_contents" (
  "id" integer PRIMARY KEY,
  "name" varchar NOT NULL,
  "encoding" varchar NOT NULL,
  "content" BLOB NOT NULL,
  "updated_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL,
  "certificate_id" integer NOT NULL,

  FOREIGN KEY(certificate_id) REFERENCES certificates(id)
);

CREATE TABLE "certificates" (
  "id" integer PRIMARY KEY,
  "display_name" varchar NOT NULL,
  "common_name" varchar NOT NULL,
  "organization" varchar,
  "subject_alternative_names" varchar,
  "key_length" integer NOT NULL,
  "status" integer NOT NULL,
  "status_message" varchar,
  "requested_on" timestamp NOT NULL,
  "cipher_algorithm_id" integer NOT NULL,
  "hash_algorithm_id" integer NOT NULL,
  "certificate_contents_id" integer,

  FOREIGN KEY(hash_algorithm_id) REFERENCES hash_algorithm(id),
  FOREIGN KEY(cipher_algorithm_id) REFERENCES cipher_algorithm(id),
  FOREIGN KEY(certificate_contents_id) REFERENCES certificate_contents(id)
);

CREATE TABLE "certificate_timeline" (
  "id" integer PRIMARY KEY NOT NULL,
  "certificate_id" integer NOT NULL,
  "event" integer NOT NULL,
  "status" integer NOT NULL,
  "updated_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL,

  FOREIGN KEY(certificate_id) REFERENCES certificates(id)
);

CREATE TABLE "certificate_authorities" (
  "id" integer PRIMARY KEY,
  "name" varchar NOT NULL,
  "server" varchar NOT NULL,
  "credential_id" integer NOT NULL,

  FOREIGN KEY(credential_id) REFERENCES credentials(id)
);

CREATE TABLE "certificate_request_authority" (
  "id" integer PRIMARY KEY,
  "certificate_id" integer NOT NULL,
  "certificate_authority_id" integer NOT NULL,
  "template_name" varchar NOT NULL,

  FOREIGN KEY(certificate_id) REFERENCES certificates(id),
  FOREIGN KEY(certificate_authority_id) REFERENCES certificate_authorities(id)
);

CREATE TABLE "credentials" (
  "id" integer PRIMARY KEY,
  "username" varchar NOT NULL,
  "password" varchar NOT NULL
);

CREATE TABLE "scheduler_scheduledset" (
  "id" BLOB PRIMARY KEY,
  "retry" boolean NOT NULL,
  "retry_count" integer NOT NULL,
  "created_at" timestamp NOT NULL,
  "enqueued_at" timestamp NOT NULL,
  "perform_at" timestamp NOT NULL,
  "processor" varchar NOT NULL,
  "arguments" BLOB
);

CREATE TABLE "scheduler_inprogressset" (
  "id" BLOB PRIMARY KEY,
  "retry" boolean,
  "retry_count" integer NOT NULL,
  "created_at" timestamp NOT NULL,
  "enqueued_at" timestamp NOT NULL,
  "processor" varchar NOT NULL,
  "arguments" BLOB
);

CREATE TABLE "scheduler_queue" (
  "id" BLOB PRIMARY KEY,
  "retry" boolean NOT NULL,
  "retry_count" integer NOT NULL,
  "created_at" timestamp NOT NULL,
  "enqueued_at" timestamp NOT NULL,
  "processor" varchar NOT NULL,
  "arguments" BLOB
);

CREATE TABLE "scheduler_failed" (
  "id" BLOB PRIMARY KEY,
  "retry" boolean,
  "retry_count" integer NOT NULL,
  "created_at" timestamp NOT NULL,
  "enqueued_at" timestamp NOT NULL,
  "processor" varchar NOT NULL,
  "arguments" BLOB,
  "log" varchar NOT NULL
);

CREATE TABLE "scheduler_completed" (
  "id" BLOB PRIMARY KEY,
  "retry" boolean,
  "retry_count" integer NOT NULL,
  "created_at" timestamp NOT NULL,
  "enqueued_at" timestamp NOT NULL,
  "processor" varchar NOT NULL,
  "arguments" BLOB,
  "log" varchar NOT NULL
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

INSERT INTO "certificates" (display_name, common_name, key_length, status, hash_algorithm_id, cipher_algorithm_id, requested_on) VALUES
('example.com', 'example.com', 2048, 1, 2, 1, '2019-01-01 00:00:00');
