package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"

	"github.com/cjp2600/structify/example/case_one/db"
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

	user, err := userStorage.FindById(ctx, "be0548df-9a62-4698-8cbe-bb96dd496682")
	if err != nil {
		if errors.Is(err, db.ErrRowNotFound) {
			fmt.Println("user not found")
			return
		}
	}

	// lazy load device
	if err := userStorage.LoadDevice(ctx, user); err != nil {
		panic(err)
	}

	// lazy load settings
	if err := userStorage.LoadSettings(ctx, user); err != nil {
		panic(err)
	}

	// lazy load addresses
	if err := userStorage.LoadAddresses(ctx, user); err != nil {
		panic(err)
	}

	fmt.Println(fmt.Sprintf("User: %+v", user))
	fmt.Println(fmt.Sprintf("Device: %+v", user.Device))
	fmt.Println(fmt.Sprintf("Setting: %+v", user.Settings))
	fmt.Println(fmt.Sprintf("Addreses: %+v", user.Addresses))
}
