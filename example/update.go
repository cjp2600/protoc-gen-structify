package main

import (
	"context"
	"github.com/cjp2600/structify/example/db"
)

func main() {
	connection, err := db.Open(db.Dsn(
		"localhost",
		5432,
		"test",
		"test",
		"testdb",
		"disable",
		0,
	))
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	store := db.NewBlogStorages(connection)
	{
		err := store.CreateTables()
		if err != nil {
			panic(err)
		}
	}

	newName := "Ivan"

	// Create Transaction Manager for store
	err = store.TxManager().ExecFuncWithTx(context.Background(), func(ctx context.Context) error {
		err := store.GetUserStorage().Update(ctx, "d0e628b8-3266-480b-bb65-cfc356121b29", &db.UserUpdate{
			Name: &newName,
		})
		if err != nil {
			return err
		}
		return nil
	})
}
