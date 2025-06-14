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
- **Filtering System**: Built-in filtering and querying capabilities
- **Relation Handling**: Handles various types of relations

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

## Filtering System

The generated code provides both convenient, type-safe filter helpers for each field and generic helpers for dynamic cases.

**How to use:**
- Use the generated wrappers for each field and filter type (e.g., `db.UserAgeEq`, `db.UserEmailLike`, etc.).
- For custom or dynamic cases, use generic helpers like `db.Eq("field", value)`.
- Always wrap your filter (or filter composition) with `db.FilterBuilder(...)` when passing to query methods.
- Combine filters with `db.And(...)`, `db.Or(...)`, etc.

**Example:**
```go
// Get users with pagination and complex filter
users, paginator, err := userStorage.FindManyWithPagination(
    ctx,
    10, // limit
    1,  // page
    db.FilterBuilder(
        db.And(
            db.Eq("status", "active"),
            db.UserAgeEq(12),
        ),
    ),
)
```

**Available filter helpers for each field:**
- `db.UserAgeEq(value int32)`
- `db.UserAgeGT(value int32)`
- `db.UserAgeBetween(min, max int32)`
- ...and so on for every field in your struct

**Generic filter helpers:**
- `db.Eq("field", value)`
- `db.NotEq("field", value)`
- `db.In("field", values...)`
- `db.Like("field", pattern)`
- etc.

**Logical composition:**
- `db.And(filter1, filter2, ...)`
- `db.Or(filter1, filter2, ...)`

**Note:**
- Always use `db.FilterBuilder(...)` to pass filters to search/query methods.
- You can mix generic and generated filters inside logical compositions.

## Filtering System — Practical Examples

### 1. Using Generated Field Filters

For each struct field, filter functions are generated, for example:
- `UserAgeEq(value int32)`
- `UserEmailLike(value string)`
- `BotUserIdEq(value string)`
- etc.

**Example:**
```go
users, err := userStorage.FindMany(ctx,
    db.FilterBuilder(db.UserAgeEq(25)),
    db.FilterBuilder(db.UserEmailLike("%@gmail.com")),
)
```

### 2. Using Generic Filters

You can use generic filters for dynamic scenarios:
```go
users, err := userStorage.FindMany(ctx,
    db.FilterBuilder(db.Eq("status", "active")),
    db.FilterBuilder(db.Like("email", "%@gmail.com")),
)
```

### 3. Composing Filters

For complex conditions, use composition:
```go
users, paginator, err := userStorage.FindManyWithPagination(
    ctx,
    10, // limit
    1,  // page
    db.FilterBuilder(
        db.And(
            db.Eq("status", "active"),
            db.UserAgeEq(12),
            db.Or(
                db.UserEmailLike("%@gmail.com"),
                db.UserEmailLike("%@yahoo.com"),
            ),
        ),
    ),
)
```

### 4. Filters for Related Entities

To filter by relations, use the corresponding filters:
```go
bots, err := botStorage.FindMany(ctx,
    db.FilterBuilder(db.BotUserIdEq("user-123")),
)
```
Or for date range:
```go
bots, err := botStorage.FindMany(ctx,
    db.FilterBuilder(db.BotCreatedAtBetween(time1, time2)),
)
```

### 5. Batch Filters

```go
bots, err := botStorage.FindMany(ctx,
    db.FilterBuilder(db.BotUserIdIn("user-1", "user-2", "user-3")),
)
```

### 6. Sorting

```go
users, err := userStorage.FindMany(ctx,
    db.FilterBuilder(db.UserAgeOrderBy(true)), // true = ASC, false = DESC
)
```

---

## Summary

- **Always use `db.FilterBuilder(...)`** to pass filters to query methods.
- You can combine generic filters (`db.Eq`, `db.Like`, ...) and generated filters (`db.UserAgeEq`, `db.BotUserIdEq`, ...).
- For complex conditions, use composition with `db.And`, `db.Or`.
- For batch operations, use filters like `FieldIn`, `FieldNotIn`.
- For sorting, use `FieldOrderBy`.

## Join Operations

The system supports various types of joins:

```go
type JoinType string

const (
    LeftJoin  JoinType = "LEFT"
    InnerJoin JoinType = "INNER"
    RightJoin JoinType = "RIGHT"
)

// Example join
join := Join(
    InnerJoin,
    userTable,
    Eq("users.id", "posts.user_id")
)
```

## Transactions

The generated code includes transaction support:

```go
// Start transaction
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

// Use transaction in context
ctx := WithTx(ctx, tx)

// Perform operations
err = userStorage.Create(ctx, user)
if err != nil {
    return err
}

// Commit transaction
err = tx.Commit()
```

## Error Handling

The generated code uses `fmt.Errorf` for error wrapping:

```go
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

## Relations

Define relations in your protobuf messages:

```protobuf
message Post {
  string id = 1 [(structify.field) = {primary_key: true}];
  string title = 2;
  string content = 3;
  string user_id = 4 [(structify.field) = {index: true}];
  User author = 5 [(structify.field) = {relation: { field: "user_id", reference: "id" }}];
}
```

## Examples

### Basic CRUD Operations

```go
// Create
user := &User{
    Name: "John Doe",
    Age: 30,
    Email: "john@example.com",
}
id, err := userStorage.Create(ctx, user)

// Read
user, err := userStorage.FindByID(ctx, id)

// Update
update := &UserUpdate{
    Name: "John Smith",
}
err = userStorage.Update(ctx, id, update)

// Delete
err = userStorage.DeleteByID(ctx, id)
```

### Querying with Filters

```go
// Find users with age > 18 and active status
users, err := userStorage.FindMany(ctx,
    And(
        Gt("age", 18),
        Eq("status", "active"),
    ),
)

// Find one user by email
user, err := userStorage.FindOne(ctx,
    Eq("email", "john@example.com"),
)
```

### Pagination

```go
// Get users with pagination
users, paginator, err := userStorage.FindManyWithPagination(
    ctx,
    10,  // limit
    1,   // page
    Eq("status", "active"),
)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.