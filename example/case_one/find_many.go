package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/cjp2600/protoc-gen-structify/example/case_one/db"
)

func findManyExample() {
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

	config := &db.Config{
		DB: &db.DB{
			DBRead:  connection,
			DBWrite: connection,
		},
	}

	store, err := db.NewBlogStorages(config)
	if err != nil {
		panic(err)
	}
	userStorage := store.GetUserStorage()

	//
	//  FindMany
	//
	builder := db.FilterBuilder(
		db.UserAgeGT(20),
	) // where age > 20

	users, err := userStorage.FindMany(ctx, builder)
	if err != nil {
		if errors.Is(err, db.ErrRowNotFound) {
			fmt.Println("user not found")
			return
		}
	}

	fmt.Println(fmt.Sprintf("User: %+v", users))

	count, err := userStorage.Count(ctx, builder)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, user := range users {
		fmt.Println(fmt.Sprintf("Age: %d", user.Age))
	}

	fmt.Println(fmt.Sprintf("Count: %+v", count))
}
