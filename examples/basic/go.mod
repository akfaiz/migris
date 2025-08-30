module github.com/afkdevs/go-schema/examples/basic

go 1.23.0

require (
	github.com/afkdevs/go-schema v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.24.3
	github.com/urfave/cli/v3 v3.3.8
)

require (
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
)

replace github.com/afkdevs/go-schema => ../..
