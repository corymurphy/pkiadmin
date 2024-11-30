migrateup:
	.\migrate.exe -path ./db/migrations -database sqlite3://pkiadmin.db --verbose up

sqlc-generate:
	.\sqlc.exe generate
