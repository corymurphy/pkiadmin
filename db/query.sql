--------------------------------------------
-- certificate_cryptographic_api
--------------------------------------------

-- name: GetCertCryptoApi :one
SELECT id, name, shortname FROM certificate_cryptographic_api
WHERE id = ? LIMIT 1;

-- name: ListCertCryptoApi :many
SELECT id, name, shortname FROM certificate_cryptographic_api
ORDER BY id;

-- name: CreateCertCryptoApi :one
INSERT INTO certificate_cryptographic_api (
  name, shortname
) VALUES (
  ?, ?
)
RETURNING *;

-- name: UpdateCertCryptoApi :exec
UPDATE certificate_cryptographic_api
set name = ?,
shortname = ?
WHERE id = ?
RETURNING *;

-- name: DeleteCertCryptoApi :exec
DELETE FROM certificate_cryptographic_api
WHERE id = ?;



--------------------------------------------
-- signing_request_api
--------------------------------------------

-- name: GetSigningRequestApi :one
SELECT id, name FROM signing_request_api
WHERE id = ? LIMIT 1;

-- name: ListSigningRequestApi :many
SELECT id, name FROM signing_request_api
ORDER BY id;

-- name: CreateSigningRequestApi :one
INSERT INTO signing_request_api (
  name
) VALUES (
  ?
)
RETURNING id, name;

-- name: UpdateSigningRequestApi :exec
UPDATE signing_request_api
set name = ?
WHERE id = ?
RETURNING id, name;

-- name: DeleteSigningRequestApi :exec
DELETE FROM signing_request_api
WHERE id = ?;


--------------------------------------------
-- cipher_algorithm
--------------------------------------------

-- name: GetCipherAlgorithm :one
SELECT id, name, keysize FROM cipher_algorithm
WHERE id = ? LIMIT 1;

-- name: ListCipherAlgorithm :many
SELECT id, name, keysize FROM cipher_algorithm
ORDER BY id;

-- name: CreateCipherAlgorithm :one
INSERT INTO cipher_algorithm (
  name, keysize
) VALUES (
  ?, ?
)
RETURNING id, name;

-- name: UpdateCipherAlgorithm :exec
UPDATE cipher_algorithm
set name = ?, keysize = ?
WHERE id = ?
RETURNING id, name, keysize;

-- name: DeleteCipherAlgorithm :exec
DELETE FROM cipher_algorithm
WHERE id = ?;


--------------------------------------------
-- hash_algorithm
--------------------------------------------

-- name: GetHashAlgorithm :one
SELECT id, name FROM hash_algorithm
WHERE id = ? LIMIT 1;

-- name: ListHashAlgorithm :many
SELECT id, name FROM hash_algorithm
ORDER BY id;

-- name: CreateHashAlgorithm :one
INSERT INTO hash_algorithm (
  name
) VALUES (
  ?
)
RETURNING id, name;

-- name: UpdateHashAlgorithm :exec
UPDATE hash_algorithm
set name = ?
WHERE id = ?
RETURNING id;

-- name: DeleteHashAlgorithm :exec
DELETE FROM hash_algorithm
WHERE id = ?;

--------------------------------------------
-- certificate_requests
--------------------------------------------

-- name: GetCertificateRequest :one
SELECT id,
      display_name,
      signing_algorithm,
      key_length,
      requested_on,
      certificate_cryptographic_api_id,
      signing_request_api_id,
      cipher_algorithm_id,
      hash_algorithm_id
FROM certificate_requests
WHERE id = ? LIMIT 1;

-- name: GetCertificateRequestDetailed :one
SELECT r.id AS id,
      r.display_name AS display_name,
      r.key_length AS key_length,
      r.requested_on AS requested_on,
      h.name AS hash_algorithm,
      c.name AS cipher_algorithm,
      s.name AS signing_request_api,
      capi.name as certificate_cryptographic_api
FROM certificate_requests r
INNER JOIN hash_algorithm h ON r.hash_algorithm_id = h.id
INNER JOIN cipher_algorithm c ON r.cipher_algorithm_id = c.id
INNER JOIN signing_request_api s ON r.signing_request_api_id = s.id
INNER JOIN certificate_cryptographic_api capi ON r.certificate_cryptographic_api_id = capi.id
WHERE r.id = ? LIMIT 1;

-- name: ListCertificateRequest :many
SELECT id,
      display_name,
      signing_algorithm,
      key_length,
      requested_on,
      certificate_cryptographic_api_id,
      signing_request_api_id,
      cipher_algorithm_id,
      hash_algorithm_id
FROM certificate_requests
ORDER BY id;

-- name: CertificateRequestsAndHashAlgorithm :many
SELECT r.id AS id,
      r.display_name AS display_name,
      r.key_length AS key_length,
      r.requested_on AS requested_on,
      h.name AS hash_algorithm,
      c.name AS cipher_algorithm,
      s.name AS signing_request_api,
      capi.name as certificate_cryptographic_api
FROM certificate_requests r
INNER JOIN hash_algorithm h ON r.hash_algorithm_id = h.id
INNER JOIN cipher_algorithm c ON r.cipher_algorithm_id = c.id
INNER JOIN signing_request_api s ON r.signing_request_api_id = s.id
INNER JOIN certificate_cryptographic_api capi ON r.certificate_cryptographic_api_id = capi.id
ORDER BY r.id;

-- name: CreateCertificateRequest :one
INSERT INTO certificate_requests (
      display_name,
      signing_algorithm,
      key_length,
      requested_on,
      certificate_cryptographic_api_id,
      signing_request_api_id,
      cipher_algorithm_id,
      hash_algorithm_id
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: DeleteCertificateRequest :exec
DELETE FROM certificate_requests
WHERE id = ?;
