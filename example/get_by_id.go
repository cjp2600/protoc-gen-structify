package main

import (
	"context"
	"fmt"
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

	ctx := context.Background()

	store := db.NewBlogStorages(connection)
	{
		err := store.CreateTables(ctx)
		if err != nil {
			panic(err)
		}
	}

	user, err := store.GetUserStorage().GetById(ctx, "d0e628b8-3266-480b-bb65-cfc356121b29")
	if err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("User: %+v", user))
}
