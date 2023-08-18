package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
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

	//
	//  FindById
	//

	user, err := userStorage.FindById(ctx, "d0e628b8-3266-480b-bb65-cfc356121b28")
	if err != nil {
		if errors.Is(err, db.ErrRowNotFound) {
			fmt.Println("user not found")
			return
		}
	}

	fmt.Println(fmt.Sprintf("User: %+v", user))
}
