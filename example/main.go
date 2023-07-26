package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	store "github.com/cjp2600/structify/example/db"
)

func main() {
	client, err := connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	// migrate database
	if err := migrate(client); err != nil {
		log.Fatalf("failed to migrate database: %s", err)
	}

	var userStore = client.User() // UserStore is a generated struct from the User table

	var val = "Doey"
	id, err := userStore.Create(context.Background(), &store.User{
		Name:     "John",
		LastName: &val,
		Age:      20,
		Email:    "cjp2601@gmail.com",
		Phones:   store.NewUserPhones([]string{"1234567890", "0987654321"}),
		Balls:    store.NewUserBalls([]int32{1, 2, 3, 4, 5, 6}),
		Numrs: &store.JSONUserNumrRepeated{
			{
				State:  21,
				Street: "street 1",
				Zip:    123,
			},
			{
				State:  22,
				Street: "street 3",
				Zip:    223,
			},
			{
				State:  33,
				Street: "street 4",
				Zip:    555,
			},
		},
		NotificationSettings: &store.JSONUserNotificationSetting{
			RegistrationEmail: true,
			OrderEmail:        false,
		},
	})
	if err != nil {
		log.Fatalf("failed to create user: %s", err)
	}
	fmt.Println(id)

	/*	err = userStore.Update(context.Background(), "c68e47a5-56c1-4c54-a7ce-2b401a66f134", &store.UserUpdateRequest{
			NotificationSettings: &store.JSONUserNotificationSettings{
				RegistrationEmail: false,
				OrderEmail:        false,
			},
		})
		if err != nil {
			log.Fatalf("failed to update user: %s", err)
		}*/

	// get all users from the database where age is between 0 and 10 and between 20 and 30
	// and order by created_at in ascending order
	/*	users, err := userStore.FindMany(
			store.Or(
				store.WhereUserAgeBetween(0, 10),
				store.WhereUserAgeBetween(20, 30),
			),
			store.Limit(10),
			store.WhereUserCreatedAtOrderBy(true),
		)
		if err != nil {
			log.Fatalf("failed to find users: %s", err)
		}

		fmt.Println(dump(users))*/
}

// connect is a helper function to connect to the database.
func connect() (*store.BlogDBClient, error) {
	db, err := store.DBConnect("localhost", 5432, "test", "test", "testdb", store.WithSSLMode("disable"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return store.NewBlogDBClient(db), nil
}

// migrate is a helper function to migrate the database.
func migrate(client *store.BlogDBClient) error {
	return client.CreateTables()
}

// dump is a helper function to print structs as JSON.
func dump(s interface{}) string {
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	return string(jsonData)
}
