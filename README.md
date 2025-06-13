# protoc-gen-structify

[![Go Report Card](https://goreportcard.com/badge/github.com/cjp2600/protoc-gen-structify)](https://goreportcard.com/report/github.com/cjp2600/protoc-gen-structify)

`protoc-gen-structify` is a powerful Protocol Buffers (protobuf) plugin that generates Go code for database operations. It provides a seamless way to create type-safe database access layers from your protobuf definitions.

## Features

- **Multiple Database Support**: Currently supports PostgreSQL, SQLite, and ClickHouse
- **Type-Safe Operations**: Generates strongly-typed Go code for database operations
- **CRUD Operations**: Automatically generates Create, Read, Update, Delete operations
- **Relations Support**: Handles one-to-one, one-to-many, and many-to-many relationships
- **Customizable**: Supports custom options for table names, field names, and more
- **Transaction Support**: Built-in transaction management
- **Query Builder**: Integrated with Squirrel query builder for complex queries

## Installation

```bash
go install github.com/cjp2600/protoc-gen-structify@latest
```

Make sure your `GOPATH/bin` is in your `PATH` environment variable.

## Quick Start

1. Define your protobuf messages:

```protobuf
syntax = "proto3";

package example;

import "google/protobuf/timestamp.proto";
import "structify/options.proto";

message User {
  option (structify.table) = "users";
  
  int64 id = 1 [(structify.field).primary_key = true];
  string name = 2;
  string email = 3 [(structify.field).unique = true];
  google.protobuf.Timestamp created_at = 4;
  
  repeated Post posts = 5 [(structify.relation) = {
    field: "user_id",
    reference: "id"
  }];
}

message Post {
  option (structify.table) = "posts";
  
  int64 id = 1 [(structify.field).primary_key = true];
  string title = 2;
  string content = 3;
  int64 user_id = 4 [(structify.field).foreign_key = "users.id"];
}
```

2. Generate the code:

```bash
protoc --go_out=. --structify_out=. user.proto
```

3. Use the generated code:

```go
package main

import (
    "context"
    "database/sql"
    "log"
    
    _ "github.com/lib/pq"
    "your/package/generated"
)

func main() {
    db, err := sql.Open("postgres", "postgres://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Initialize the client
    client := generated.NewUserDatabaseClient(db)
    
    // Create a new user
    user := &generated.User{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    err = client.Create(context.Background(), user)
    if err != nil {
        log.Fatal(err)
    }
    
    // Find user by ID
    found, err := client.GetByID(context.Background(), user.ID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Update user
    found.Name = "John Updated"
    err = client.Update(context.Background(), found)
    if err != nil {
        log.Fatal(err)
    }
    
    // Find users with conditions
    users, err := client.FindMany(context.Background(), generated.NewUserCondition().
        WhereName("John Updated").
        WhereEmail("john@example.com"))
    if err != nil {
        log.Fatal(err)
    }
}
```

## Supported Databases

### PostgreSQL
```protobuf
option (structify.provider) = "postgres";
```

### SQLite
```protobuf
option (structify.provider) = "sqlite";
```

### ClickHouse
```protobuf
option (structify.provider) = "clickhouse";
```

## Field Options

### Primary Key
```protobuf
int64 id = 1 [(structify.field).primary_key = true];
```

### Unique Constraint
```protobuf
string email = 1 [(structify.field).unique = true];
```

### Foreign Key
```protobuf
int64 user_id = 1 [(structify.field).foreign_key = "users.id"];
```

### Nullable Field
```protobuf
string description = 1 [(structify.field).nullable = true];
```

## Relation Options

### One-to-Many
```protobuf
repeated Post posts = 1 [(structify.relation) = {
    field: "user_id",
    reference: "id"
}];
```

### Many-to-One
```protobuf
User user = 1 [(structify.relation) = {
    field: "id",
    reference: "user_id"
}];
```

## Generated Code Structure

The plugin generates the following components:

1. **Database Client**: Main interface for database operations
2. **Storage**: Type-safe storage implementation
3. **Conditions**: Query builder for complex queries
4. **Types**: Go structs matching your protobuf messages
5. **Constants**: Generated constants for field names and table names

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.