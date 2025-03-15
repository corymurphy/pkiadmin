# General Notes

## Productionize migrations

```go
//go:embed migration/*/*.sql
var migrationFS embed.FS
```
