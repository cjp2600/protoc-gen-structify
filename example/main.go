package main

import (
	"encoding/json"
	"fmt"
	"log"

	store "github.com/cjp2600/structify/example/db"
)

func main() {
	db, err := store.DBConnect("localhost", 5432, "test", "test", "testdb", store.WithSSLMode("disable"))
	if err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}
	defer db.Close()

	client := store.NewBlogDBClient(db)
	userStore := client.User()

	users, err := userStore.FindOne()
	if err != nil {
		log.Fatalf("failed to find users: %s", err)
	}

	fmt.Println(dump(users))
}

// dump is a helper function to print structs as JSON.
func dump(s interface{}) string {
	jsonData, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshaling failed: %s", err)
	}
	return string(jsonData)
}
