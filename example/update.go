package main

import (
	"context"
	"time"

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	store := db.NewBlogStorages(connection)
	userStorage := store.GetUserStorage()

	newName := "Ivan"

	// Create Transaction Manager for store
	err = store.TxManager().ExecFuncWithTx(ctx, func(ctx context.Context) error {
		// Update user name
		err := userStorage.Update(ctx, "d0e628b8-3266-480b-bb65-cfc356121b29", &db.UserUpdate{
			Name: &newName,
		})
		if err != nil {
			return err
		}
		return nil
	})
}
