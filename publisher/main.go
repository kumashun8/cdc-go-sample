package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/jackc/pgx/v5/pgconn"
)

const CONN = "postgres://postgres:sample@localhost:5432/postgres?sslmode=disable&replication=database"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	conn, err := pgconn.Connect(ctx, CONN)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	res, err := conn.Exec(ctx, "SELECT * FROM t;").ReadAll()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}
