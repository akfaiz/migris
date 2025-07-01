module github.com/ahmadfaizk/schema/examples/basic

go 1.23.0

require (
	github.com/ahmadfaizk/schema v0.1.0
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.24.3
	github.com/urfave/cli/v3 v3.3.8
)

replace github.com/ahmadfaizk/schema => ../..

require (
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
)
