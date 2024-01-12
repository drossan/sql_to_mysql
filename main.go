package main

import (
	"flag"
	"github.com/drossan/sql_to_mysql/cmd"
)

var schemas string

func setFlags() {
	flag.StringVar(&schemas, "schemas", "no", "Migrate SQL schemas to MySQL")
}

func main() {
	setFlags()
	cmd.StartMigration(schemas)
}
