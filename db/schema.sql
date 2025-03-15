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

CREATE TABLE certificate_contents (
  id integer PRIMARY KEY,
  name varchar NOT NULL,
  encoding varchar NOT NULL,
  content BLOB NOT NULL,
  updated_at timestamp NOT NULL,
  created_at timestamp NOT NULL,
  certificate_request_id integer NOT NULL,

  FOREIGN KEY(certificate_request_id) REFERENCES certificate_requests(id)
);

CREATE TABLE certificate_requests (
  id integer PRIMARY KEY,
  display_name varchar,
  signing_algorithm varchar,
  key_length integer,
  requested_on timestamp,
  status integer,
  status_message varchar,
  certificate_cryptographic_api_id integer,
  signing_request_api_id integer,
  cipher_algorithm_id integer,
  hash_algorithm_id integer,

  FOREIGN KEY(certificate_cryptographic_api_id) REFERENCES certificate_cryptographic_api(id),
  FOREIGN KEY(signing_request_api_id) REFERENCES signing_request_api(id),
  FOREIGN KEY(hash_algorithm_id) REFERENCES hash_algorithm(id),
  FOREIGN KEY(cipher_algorithm_id) REFERENCES cipher_algorithm(id)
);

CREATE TABLE certificate_authorities (
  id integer PRIMARY KEY,
  name varchar NOT NULL,
  server varchar NOT NULL,
  credential_id integer NOT NULL,

  FOREIGN KEY(credential_id) REFERENCES credentials(id)
);

CREATE TABLE credentials (
  id integer PRIMARY KEY,
  username varchar NOT NULL,
  password varchar NOT NULL
);

CREATE TABLE scheduler_scheduledset (
  id uuid PRIMARY KEY ,
  retry boolean NOT NULL,
  retry_count integer NOT NULL,
  created_at timestamp NOT NULL ,
  enqueued_at timestamp NOT NULL,
  perform_at timestamp NOT NULL,
  processor varchar NOT NULL,
  arguments BLOB NOT NULL
);

CREATE TABLE scheduler_inprogressset (
  id uuid PRIMARY KEY,
  retry boolean NOT NULL,
  retry_count integer NOT NULL,
  created_at timestamp NOT NULL,
  enqueued_at timestamp NOT NULL,
  processor varchar NOT NULL,
  arguments BLOB NOT NULL
);

CREATE TABLE scheduler_queue (
  id uuid PRIMARY KEY,
  retry boolean NOT NULL,
  retry_count integer NOT NULL,
  created_at timestamp NOT NULL,
  enqueued_at timestamp NOT NULL,
  processor varchar NOT NULL,
  arguments BLOB NOT NULL
);

CREATE TABLE scheduler_failed (
  id uuid PRIMARY KEY,
  retry boolean NOT NULL,
  retry_count integer NOT NULL,
  created_at timestamp NOT NULL,
  enqueued_at timestamp NOT NULL,
  processor varchar NOT NULL,
  arguments BLOB NOT NULL,
  log varchar NOT NULL
);

CREATE TABLE scheduler_completed (
  id uuid PRIMARY KEY,
  retry boolean NOT NULL,
  retry_count integer NOT NULL,
  created_at timestamp NOT NULL,
  enqueued_at timestamp NOT NULL,
  processor varchar NOT NULL,
  arguments BLOB NOT NULL,
  log varchar NOT NULL
);

CREATE TABLE certificate_requests_timeline (
  id integer PRIMARY KEY NOT NULL,
  certificate_request_id integer NOT NULL,
  event integer NOT NULL,
  status integer NOT NULL,
  updated_at timestamp NOT NULL,
  created_at timestamp NOT NULL,

  FOREIGN KEY(certificate_request_id) REFERENCES certificate_requests(id)
);
