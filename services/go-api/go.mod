module github.com/m-spangenberg/demo-monetized-api/services/go-api

go 1.26.2

require (
	github.com/m-spangenberg/demo-monetized-api/pkg/common v0.0.0-00010101000000-000000000000
	github.com/mattn/go-sqlite3 v1.14.44
	github.com/redis/go-redis/v9 v9.19.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
)

replace github.com/m-spangenberg/demo-monetized-api/pkg/common => ../../pkg/common
