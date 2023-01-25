package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

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

	for i := 0; i < 10; i++ {
		_, err = conn.Exec(ctx, "INSERT INTO t (id, name) VALUES(1, 'test');").ReadAll()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}
}
