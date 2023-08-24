package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cjp2600/protoc-gen-structify/example/case_one/db"
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
	//  FindManyWithPagination
	//

	users, pagination, err := userStorage.FindManyWithPagination(ctx, 5, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("User: %+v, Pagination: %+v", users, pagination))

	for _, user := range users {
		fmt.Println(fmt.Sprintf("Name: %s, Age: %d", user.Name, user.Age))
	}

}
