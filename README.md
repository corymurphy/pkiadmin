# PKI Admin

**NOTE** This project is currently under very early development and is not an indication of the final state. It is an attempt to rewrite [CertificateManager](https://github.com/corymurphy/CertificateManager) in Go/HTMX.

Web app for managing Active Directory Certificate Services infrastructure and a key escrow system

## Development

### Migrations

```shell
# create new migration
migrate create -ext sql -dir db/migrations -seq sequence-name
```

### Database Code

```shell
# generate golang from query.sql
docker run --rm -v $(pwd):/src -w /src sqlc/sqlc generate
```

## TODO

* [ ] implement auditing
* [ ] implement auth
    - [ ] saml
    - [ ] oidc
    - [ ] local
* [ ] install certificates on iis
    - [ ] build powershell remote execution system
* [x] show templates under certificate authorities
* [ ] separate certs/requests?
* [ ] fix the status badge to be less of a mess of css
* [ ] support group managed service accounts
* [ ] support kerberos for ca auth
* [ ] implement search
* [x] build job scheduling system
* [x] issue certs using adcs from linux
* [ ] create installation docs
* [ ] create ci
* [ ] encrypt secrets
* [ ] refactor template rendering
* [x] paginaton for requests
* [ ] paginaton for scheduler
* [ ] generate pfx
* [ ] generate encrypted private key
* [ ] fix ca template slow warmup
* [ ] add simple logging to replace fmt.println usage
