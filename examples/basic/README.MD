# Migris Example
This is a basic example of using Migris for database migrations in Go.

```bash
go build -o migris
```


```bash
$ ./migris 
NAME:
   migrate - Migration tool

USAGE:
   migrate [global options] [command [command options]]

COMMANDS:
   create   Create a new migration file
   up       Run all pending migrations
   reset    Rollback all migrations
   down     Rollback the last migration
   status   Show the status of migrations
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

```bash
$ ./migris status
20250904164848_create_users_table.go ................................... Pending
20250904165150_create_posts_table.go ................................... Pending
20250904165317_create_comments_table.go ................................ Pending
20250904165421_insert_data.go .......................................... Pending

$ ./migris up
 INFO  Running migrations.

20250904164848_create_users_table.go ............................... 9.24ms DONE
20250904165150_create_posts_table.go ............................... 6.59ms DONE
20250904165317_create_comments_table.go ............................ 6.90ms DONE
20250904165421_insert_data.go ...................................... 3.10ms DONE

$ ./migris status
20250904164848_create_users_table.go ................................... Applied
20250904165150_create_posts_table.go ................................... Applied
20250904165317_create_comments_table.go ................................ Applied
20250904165421_insert_data.go .......................................... Applied
```