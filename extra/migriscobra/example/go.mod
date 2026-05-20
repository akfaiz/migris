module github.com/akfaiz/migris/extra/migriscobra/example

go 1.23.0

require (
	github.com/akfaiz/migris v0.4.1
	github.com/akfaiz/migris/extra/migriscobra v0.0.0
	github.com/jackc/pgx/v5 v5.7.5
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/pressly/goose/v3 v3.26.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.33.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace github.com/akfaiz/migris => ../../..

replace github.com/akfaiz/migris/extra/migriscobra => ../
