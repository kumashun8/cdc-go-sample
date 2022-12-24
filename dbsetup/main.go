package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	dsn := "postgres://postgres:sample@localhost:5432/postgres?sslmode=disable"
	db := bun.NewDB(
		sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn))),
		pgdialect.New(),
	)

	setupCmds := []string{
		// TODO 冪等にする
		// "pg_create_logical_replication_slot('replication_slot', 'test_decoding');",
		"slot_name, plugin, slot_type FROM pg_replication_slots;",
	}

	for _, v := range setupCmds {
		// TODO どのクエリでもいい感じにScanできるようにする
		output := make([]string, 3)
		if err := db.NewSelect().ColumnExpr(v).Scan(context.Background(), &output[0], &output[1], &output[2]); err != nil {
			log.Fatal(err)
		}
		fmt.Println(output)
	}

}
