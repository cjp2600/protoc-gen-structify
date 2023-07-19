package main

import (
	"fmt"
	"github.com/cjp2600/structify/example/db"
)

func main() {
	connect, err := db.DBConnect("localhost", 5432, "test", "test", "testdb", db.WithSSLMode("disable"))
	if err != nil {
		panic(err)
	}
	defer connect.Close()

	client := db.NewBlogDBClient(connect)
	/*
		if err := client.CreateTables(); err != nil {
			panic(err)
		}*/

	/*	// Create a user
		ids, err := client.User().CreateMany(context.Background(), []*db.User{
			{
				Name:  "test",
				Age:   18,
				Email: "test@test.com",
			},
			{
				Name:  "test1",
				Age:   18,
				Email: "test1@test.com",
			},
			{
				Name:  "test2",
				Age:   18,
				Email: "test2@test.com",
			},
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(ids)*/

	/*	user, err := client.User().FindById("4f26dfbe-2f30-4334-89ad-f7b8d4e8078f")
		if err != nil {
			panic(err)
		}

		fmt.Println(user)*/

	users, err := client.User().FindMany(
		db.Or(
			db.UserNameEq("test2"),
			db.UserEmailEq("test@test.com"),
			db.UserAgeGreaterThan(15),
		),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(users)

	/*	count, err := client.User().Count(
					db.Or(
						db.UserAgeGreaterThan(30),
					),
				)
				if err != nil {
					panic(err)
				}
				fmt.Println(count)

		/*	val := "testov"
			err = client.User().Update(context.Background(), "3ef245be-720e-4b84-9243-645aae39058f", &db.UserUpdateRequest{
				LastName: &val,
			})
			if err != nil {
				panic(err)
			}*/
}
