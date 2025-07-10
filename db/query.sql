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
-- certificates
--------------------------------------------

-- name: GetCertificate :one
SELECT id,
      display_name,
      common_name,
      organization,
      subject_alternative_names,
      key_length,
      status,
      status_message,
      requested_on,
      cipher_algorithm_id,
      hash_algorithm_id
FROM certificates
WHERE id = ? LIMIT 1;

-- name: GetCertificateDetailed :one
SELECT r.id AS id,
      r.display_name AS display_name,
      r.common_name AS common_name,
      r.organization AS organization,
      r.subject_alternative_names AS subject_alternative_names,
      r.key_length AS key_length,
      r.requested_on AS requested_on,
      r.status AS status,
      r.status_message AS status_message,
      h.name AS hash_algorithm,
      c.name AS cipher_algorithm
FROM certificates r
INNER JOIN hash_algorithm h ON r.hash_algorithm_id = h.id
INNER JOIN cipher_algorithm c ON r.cipher_algorithm_id = c.id
WHERE r.id = ? LIMIT 1;

-- name: ListCertificate :many
SELECT id,
      display_name,
      common_name,
      organization,
      subject_alternative_names,
      key_length,
      status,
      status_message,
      requested_on,
      cipher_algorithm_id,
      hash_algorithm_id
FROM certificates
ORDER BY id;

-- name: CertificatesAndHashAlgorithm :many
SELECT r.id AS id,
      r.display_name AS display_name,
      r.common_name AS common_name,
      r.organization AS organization,
      r.subject_alternative_names AS subject_alternative_names,
      r.key_length AS key_length,
      r.requested_on AS requested_on,
      r.status AS status,
      r.status_message AS status_message,
      h.name AS hash_algorithm,
      c.name AS cipher_algorithm
FROM certificates r
INNER JOIN hash_algorithm h ON r.hash_algorithm_id = h.id
INNER JOIN cipher_algorithm c ON r.cipher_algorithm_id = c.id
ORDER BY r.id;

-- name: CertificatesAndHashAlgorithmPaginated :many
SELECT r.id AS id,
      r.display_name AS display_name,
      r.common_name AS common_name,
      r.organization AS organization,
      r.subject_alternative_names AS subject_alternative_names,
      r.key_length AS key_length,
      r.requested_on AS requested_on,
      r.status AS status,
      r.status_message AS status_message,
      h.name AS hash_algorithm,
      c.name AS cipher_algorithm
FROM certificates r
INNER JOIN hash_algorithm h ON r.hash_algorithm_id = h.id
INNER JOIN cipher_algorithm c ON r.cipher_algorithm_id = c.id
ORDER BY r.requested_on DESC
LIMIT ? OFFSET ?;

-- name: GetCertificatesCount :one
SELECT COUNT(*) as count
FROM certificates;

-- name: CreateCertificate :one
INSERT INTO certificates (
      display_name,
      common_name,
      organization,
      subject_alternative_names,
      key_length,
      status,
      status_message,
      requested_on,
      cipher_algorithm_id,
      hash_algorithm_id
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: DeleteCertificate :exec
DELETE FROM certificates
WHERE id = ?;

-- name: UpdateCertificateStatus :exec
UPDATE certificates
set status = ?, status_message = ?
WHERE id = ?;

--------------------------------------------
-- certificate_authorities
--------------------------------------------

-- name: GetCertificateAuthority :one
SELECT ca.id AS id,
      ca.name,
      ca.server,
      c.username AS username,
      c.password AS password,
      c.id AS credential_id
FROM certificate_authorities ca
INNER JOIN credentials c ON ca.credential_id = c.id
WHERE ca.id = ? LIMIT 1;

-- name: CreateCertificateAuthority :one
INSERT INTO certificate_authorities (
  name,
  server,
  credential_id
) VALUES (
  ?, ?, ?
)
RETURNING id;

-- name: ListCertificateAuthorities :many
SELECT ca.id AS id,
      ca.name,
      ca.server,
      c.username AS username,
      c.password AS password,
      c.id AS credential_id
FROM certificate_authorities ca
INNER JOIN credentials c ON ca.credential_id = c.id;

-- name: UpdateCertificateAuthority :exec
UPDATE certificate_authorities
set name = ?, server = ?
WHERE id = ?;

-- name: DeleteCertificateAuthority :exec
DELETE FROM certificate_authorities
WHERE id = ?;

--------------------------------------------
-- credentials
--------------------------------------------
-- name: GetCredential :one
SELECT id, username, password FROM credentials
WHERE id = ? LIMIT 1;

-- name: CreateCredential :one
INSERT INTO credentials (
  username,
  password
) VALUES (
  ?, ?
)
RETURNING id;

-- name: ListCredentials :many
SELECT id, username, password FROM credentials;

-- name: UpdateCredential :exec
UPDATE credentials
set username = ?, password = ?
WHERE id = ?;

-- name: DeleteCredential :exec
DELETE FROM credentials
WHERE id = ?;

--------------------------------------------
-- scheduler_scheduledset
--------------------------------------------
-- name: GetScheduledSet :one
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  perform_at,
  processor,
  arguments
FROM scheduler_scheduledset
WHERE id = ? LIMIT 1;

-- name: CreateScheduledSet :one
INSERT INTO scheduler_scheduledset (
  id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  perform_at,
  processor,
  arguments
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: ListScheduledSet :many
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  perform_at,
  processor,
  arguments
FROM scheduler_scheduledset;

-- name: ListScheduledSetShouldPerform :many
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  perform_at,
  processor,
  arguments
FROM scheduler_scheduledset
WHERE perform_at <= datetime('now');

-- name: DeleteScheduledSet :exec
DELETE FROM scheduler_scheduledset
WHERE id = ?;

-- name: CountScheduledSet :one
SELECT COUNT(*) as count FROM scheduler_scheduledset;

--------------------------------------------
-- scheduler_inprogressset
--------------------------------------------

-- name: GetInProgressSet :one
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_inprogressset
WHERE id = ? LIMIT 1;

-- name: CreateInProgressSet :one
INSERT INTO scheduler_inprogressset (
  id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: ListInProgressSet :many
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_inprogressset;

-- name: DeleteInProgressSet :exec
DELETE FROM scheduler_inprogressset
WHERE id = ?;

-- name: CountInProgressSet :one
SELECT COUNT(*) as count FROM scheduler_inprogressset;

--------------------------------------------
-- scheduler_queue
--------------------------------------------

-- name: GetSchedulerQueueJob :one
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_queue
WHERE id = ? LIMIT 1;

-- name: GetOneSchedulerQueueJob :one
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_queue
ORDER BY id
LIMIT 1;

-- name: CreateSchedulerQueueJob :one
INSERT INTO scheduler_queue (
  id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
) VALUES (
  ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: ListSchedulerQueueJobs :many
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_queue;

-- name: DeleteSchedulerQueueJob :exec
DELETE FROM scheduler_queue
WHERE id = ?;

-- name: CountSchedulerQueueJob :one
SELECT COUNT(*) as count FROM scheduler_queue;

--------------------------------------------
-- scheduler_failed
--------------------------------------------

-- name: GetFailedJob :one
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_failed
WHERE id = ? LIMIT 1;

-- name: CreateFailedJob :one
INSERT INTO scheduler_failed (
  id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments,
  log
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: ListFailedJobs :many
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments,
  log
FROM scheduler_failed;

-- name: DeleteFailedJob :exec
DELETE FROM scheduler_failed
WHERE id = ?;

-- name: CountFailedJob :one
SELECT COUNT(*) as count FROM scheduler_failed;

--------------------------------------------
-- scheduler_completed
--------------------------------------------

-- name: GetCompletedJob :one
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments
FROM scheduler_completed
WHERE id = ? LIMIT 1;

-- name: CreateCompletedJob :one
INSERT INTO scheduler_completed (
  id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments,
  log
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING id;

-- name: ListCompletedJobsPaginated :many
SELECT id,
  retry,
  retry_count,
  created_at,
  enqueued_at,
  processor,
  arguments,
  log
FROM scheduler_completed
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: DeleteCompletedJob :exec
DELETE FROM scheduler_completed
WHERE id = ?;

-- name: CountCompletedJob :one
SELECT COUNT(*) as count FROM scheduler_completed;

--------------------------------------------
-- certificate_contents
--------------------------------------------

-- name: GetCertificateContent :one
SELECT id, name, encoding, content, updated_at, created_at, certificate_id
FROM certificate_contents
WHERE id = ? LIMIT 1;

-- name: GetCertificateContentByNameEncodingRequestID :one
SELECT id, name, encoding, content, updated_at, created_at, certificate_id
FROM certificate_contents
WHERE certificate_id = ? AND name = ? AND encoding = ? LIMIT 1;

-- name: ListCertificateContent :many
SELECT id, name, encoding, content, updated_at, created_at, certificate_id
FROM certificate_contents
WHERE certificate_id = ?;

-- name: CreateCertificateContent :one
INSERT INTO certificate_contents (
  name,
  encoding,
  content,
  certificate_id,
  updated_at,
  created_at
) VALUES (
  ?, ?, ?, ?, ?, ?
)
RETURNING id;

--------------------------------------------
-- certificate_timeline
--------------------------------------------

-- name: GetCertificateTimeline :one
SELECT id, certificate_id, status, event, created_at, updated_at
FROM certificate_timeline
WHERE id = ? LIMIT 1;

-- name: ListCertificateTimeline :many
SELECT id, certificate_id, status, event, created_at, updated_at
FROM certificate_timeline
WHERE certificate_id = ?
ORDER BY id;

-- name: CreateCertificateTimeline :one
INSERT INTO certificate_timeline (
  certificate_id,
  status,
  event,
  created_at,
  updated_at
) VALUES (
  ?, ?, ?, ?, ?
)
RETURNING id;

-- name: GetCertificateTimelineByRequest :one
SELECT id, certificate_id, status, event, created_at, updated_at
FROM certificate_timeline
WHERE certificate_id = ? AND event = ? LIMIT 1;

-- name: UpdateCertificateTimeline :exec
UPDATE certificate_timeline
SET status = ?, event = ?, updated_at = ?
WHERE id = ?;

-- name: UpdateCertificateTimelineByRequest :exec
UPDATE certificate_timeline
SET status = ?, updated_at = ?
WHERE certificate_id = ? AND event = ?;

-- name: DeleteCertificateTimeline :exec
DELETE FROM certificate_timeline
WHERE id = ?;

-- name: DeleteCertificateTimelines :exec
DELETE FROM certificate_timeline
WHERE certificate_id = ?;

--------------------------------------------
-- certificate_request_authority
--------------------------------------------

-- name: GetCertificateRequestAuthority :one
SELECT certificate_id, certificate_authority_id, template_name
FROM certificate_request_authority
WHERE certificate_id = ? LIMIT 1;

-- name: ListCertificateRequestAuthority :many
SELECT certificate_id, certificate_authority_id, template_name
FROM certificate_request_authority
WHERE certificate_id = ?;

-- name: CreateCertificateRequestAuthority :one
INSERT INTO certificate_request_authority (
  certificate_id,
  certificate_authority_id,
  template_name
) VALUES (
  ?, ?, ?
)
RETURNING certificate_id;

-- name: DeleteCertificateRequestAuthority :exec
DELETE FROM certificate_request_authority
WHERE certificate_id = ?;
