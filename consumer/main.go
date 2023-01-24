package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pglogrepl"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
)

const CONN = "postgres://postgres:sample@localhost:5432/postgres?sslmode=disable&replication=database"
const SLOT_NAME = "replication_slot"
const OUTPUT_PLUGIN = "pgoutput"
const INSERT_TEMPLATE = "create table if not exists t (id int, name text);"

var Event = struct {
	Relation string
	Columns  []string
}{}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	conn, err := pgconn.Connect(ctx, CONN)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	// 1. Create table
	if _, err := conn.Exec(ctx, INSERT_TEMPLATE).ReadAll(); err != nil {
		fmt.Println(fmt.Errorf("failed to create table: %v", err))
	}

	// 2. Ensure publication exists
	if _, err := conn.Exec(ctx, "DROP PUBLICATION IF EXISTS pub;").ReadAll(); err != nil {
		fmt.Println(fmt.Errorf("failed to drop publication: %v", err))
	}

	if _, err := conn.Exec(ctx, "CREATE PUBLICATION pub FOR ALL TABLES;").ReadAll(); err != nil {
		fmt.Println(fmt.Errorf("failed to create publication: %v", err))
	}
	// 3. Create temporary replication slot server
	if _, err = pglogrepl.CreateReplicationSlot(ctx, conn, SLOT_NAME, OUTPUT_PLUGIN, pglogrepl.CreateReplicationSlotOptions{Temporary: true}); err != nil {
		fmt.Println(fmt.Errorf("failed to create a replication slot: %v", err))
	}

	// 4. establish connection
	var msgPointer pglogrepl.LSN
	pluginArgs := []string{"proto_version '1'", "publication_names 'pub'"}
	err = pglogrepl.StartReplication(ctx, conn, SLOT_NAME, msgPointer, pglogrepl.StartReplicationOptions{PluginArgs: pluginArgs})
	if err != nil {
		fmt.Println(fmt.Errorf("failed to establish start replication: %v", err))
	}

	var pingTime time.Time
	for ctx.Err() != context.Canceled {
		if time.Now().After(pingTime) {
			if err = pglogrepl.SendStandbyStatusUpdate(ctx, conn, pglogrepl.StandbyStatusUpdate{WALWritePosition: msgPointer}); err != nil {
				fmt.Println(fmt.Errorf("failed to send standby update: %v", err))
			}
			pingTime = time.Now().Add(10 * time.Second)
		}

		ctx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()

		msg, err := conn.ReceiveMessage(ctx)
		if pgconn.Timeout(err) {
			continue
		}
		if err != nil {
			fmt.Println(fmt.Errorf("something went wrong while listening for message: %v", err))
		}

		switch msg := msg.(type) {
		case *pgproto3.CopyData:
			if msg.Data[0] == pglogrepl.XLogDataByteID {
				walLog, err := pglogrepl.ParseXLogData(msg.Data[1:])
				if err != nil {
					fmt.Println(fmt.Errorf("failed to parse logical WAL log: %v", err))
				}
				var msg pglogrepl.Message
				if msg, err = pglogrepl.Parse(walLog.WALData); err != nil {
					fmt.Println(fmt.Errorf("failed to parse logical replication message: %v", err))
				}
				fmt.Println(msg.Type().String())
			}
			switch msg.Data[0] {
			// case pglogrepl.PrimaryKeepaliveMessageByteID:
			// fmt.Println("server: confirmed standby")
			// case pglogrepl.XLogDataByteID:
			// 	walLog, err := pglogrepl.ParseXLogData(msg.Data[1:])
			// 	if err != nil {
			// 		fmt.Println(fmt.Errorf("failed to parse logical WAL log: %v", err))
			// 	}

			// 	var msg pglogrepl.Message
			// 	if msg, err = pglogrepl.Parse(walLog.WALData); err != nil {
			// 		fmt.Println(fmt.Errorf("failed to parse logical replication message: %v", err))
			// 	}
			// 	switch m := msg.(type) {
			// 	case *pglogrepl.RelationMessage:
			// 		Event.Columns = []string{}
			// 		for _, col := range m.Columns {
			// 			Event.Columns = append(Event.Columns, col.Name)
			// 		}
			// 		Event.Relation = m.RelationName
			// 	case *pglogrepl.InsertMessage:
			// 		var sb strings.Builder
			// 		sb.WriteString(fmt.Sprintf("INSERT %s(", Event.Relation))
			// 		for i := 0; i < len(Event.Columns); i++ {
			// 			sb.WriteString(fmt.Sprintf("%s: %s", Event.Columns[i], string(m.Tuple.Columns[i].Data)))
			// 		}
			// 	case *pglogrepl.UpdateMessage:
			// 		var sb strings.Builder
			// 		sb.WriteString(fmt.Sprintf("UPDATE %s(", Event.Relation))
			// 		for i := 0; i < len(Event.Columns); i++ {
			// 			sb.WriteString(fmt.Sprintf("%s: %s ", Event.Columns[i], string(m.NewTuple.Columns[i].Data)))
			// 		}
			// 		sb.WriteString(")")
			// 		fmt.Println(sb.String())
			// 	case *pglogrepl.DeleteMessage:
			// 		var sb strings.Builder
			// 		sb.WriteString(fmt.Sprintf("DELETE %s(", Event.Relation))
			// 		for i := 0; i < len(Event.Columns); i++ {
			// 			sb.WriteString(fmt.Sprintf("%s: %s ", Event.Columns[i], string(m.OldTuple.Columns[i].Data)))
			// 		}
			// 		sb.WriteString(")")
			// 		fmt.Println(sb.String())
			// 	case *pglogrepl.TruncateMessage:
			// 		fmt.Println("ALL GONE (TRUNCATE)")
			// 	}
			}
		default:
			fmt.Printf("recieved unexpected message: %T", msg)
		}
	}
}
