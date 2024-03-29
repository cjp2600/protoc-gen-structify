package main

import (
	"context"
	"fmt"
	"github.com/cjp2600/protoc-gen-structify/example/case_two/db"
	"github.com/pkg/errors"
)

func main() {
	connection, err := db.Open(db.Dsn("case_two", "", "", 3))
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	ctx := context.Background()

	store := db.NewBlogStorages(connection)
	userStorage := store.GetUserStorage()

	//
	//  FindById
	//

	user, err := userStorage.FindById(ctx, "3b9e312a-ed8d-11ee-8f3f-acde48001122")
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
