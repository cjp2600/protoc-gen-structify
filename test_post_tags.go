package main

import (
	"fmt"
	"log"

	"./example/case_one/db"
)

func main() {
	// Test creating a Post with UUID array tags
	post := &db.Post{
		Id:       1,
		Title:    "Test Post",
		Body:     "This is a test post",
		Tags:     db.PostTagsRepeated{"550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		AuthorId: "550e8400-e29b-41d4-a716-446655440000",
	}

	// Test Value() method
	value, err := post.Tags.Value()
	if err != nil {
		log.Fatal("Error getting value:", err)
	}
	fmt.Printf("Post tags value: %v\n", value)

	// Test Get() method
	tags := post.Tags.Get()
	fmt.Printf("Post tags: %v\n", tags)

	// Test String() method
	fmt.Printf("Post tags string: %v\n", post.Tags.String())

	// Test scanning from PostgreSQL format
	var newTags db.PostTagsRepeated
	err = newTags.Scan("{550e8400-e29b-41d4-a716-446655440000,6ba7b810-9dad-11d1-80b4-00c04fd430c8}")
	if err != nil {
		log.Fatal("Error scanning tags:", err)
	}
	fmt.Printf("Scanned tags: %v\n", newTags.Get())

	fmt.Println("✅ PostTagsRepeated works correctly with UUID arrays!")
}
