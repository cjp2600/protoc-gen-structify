# Composite Primary Key Support

This document describes the composite primary key support added to protoc-gen-structify for PostgreSQL.

## Overview

The generator now supports composite (multi-column) primary keys in PostgreSQL. When multiple fields are marked with `primary_key: true`, the generated code will:

1. Create a composite PRIMARY KEY constraint in the CREATE TABLE statement
2. Generate ON CONFLICT clauses that use all primary key columns
3. Maintain backward compatibility with single primary keys

## Usage Example

```protobuf
syntax = "proto3";

import "github.com/cjp2600/protoc-gen-structify/plugin/options/structify.proto";
import "google/protobuf/timestamp.proto";

package example;

option (structify.db) = {
  provider: "postgres"
};

message ActiveTag {
  // Composite primary key using customer_id and tag_name
  string customer_id = 1 [
    (structify.field) = {
      uuid: true,
      primary_key: true
    }
  ];

  // Tag name (part of composite primary key)
  string tag_name = 2 [
    (structify.field) = {
      primary_key: true
    }
  ];

  // Tag ID for reference
  string tag_id = 3;

  // Record creation timestamp
  google.protobuf.Timestamp created_at = 4 [
    (structify.field) = {default: "now()"}
  ];

  // Last update timestamp
  google.protobuf.Timestamp updated_at = 5 [
    (structify.field) = {default: "now()"}
  ];

  option (structify.opts) = {
    table: "active_tags"
  };
}
```

## Generated SQL

### CREATE TABLE Statement

When CRUDSchemas is enabled, the generated CREATE TABLE will use a composite primary key:

```sql
CREATE TABLE IF NOT EXISTS active_tags (
  customer_id UUID NOT NULL,
  tag_name TEXT NOT NULL,
  tag_id TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  PRIMARY KEY (customer_id, tag_name)
);
```

### Upsert Method

The generated Upsert method will use the composite key in the ON CONFLICT clause:

```go
func (t *activeTagStorage) Upsert(ctx context.Context, model *ActiveTag, updateFields []string, opts ...Option) (*string, error) {
    // ... query building ...
    
    // Composite primary key: customer_id + tag_name
    suffixBuilder.WriteString("ON CONFLICT (customer_id, tag_name) DO UPDATE SET ")
    
    // ... rest of the method ...
}
```

This generates SQL like:

```sql
INSERT INTO active_tags (customer_id, tag_name, tag_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (customer_id, tag_name) DO UPDATE SET
  tag_id = EXCLUDED.tag_id,
  updated_at = EXCLUDED.updated_at
RETURNING customer_id;
```

## Backward Compatibility

Single primary keys continue to work as before:

```protobuf
message User {
  string id = 1 [
    (structify.field) = {
      primary_key: true,
      uuid: true
    }
  ];
  string name = 2;
}
```

Generates:

```sql
CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  name TEXT
);
```

And Upsert with:

```sql
ON CONFLICT (id) DO UPDATE SET ...
```

## Implementation Details

### Template Functions Added

1. `getPrimaryKeys()` - Returns all fields marked as primary keys
2. `hasCompositePrimaryKey()` - Returns true if more than one primary key exists

### Files Modified

1. `plugin/provider/postgres/templater/table.go` - Added new template functions
2. `plugin/provider/postgres/tmpl/table.templater.tmpl.go` - Updated templates for:
   - CREATE TABLE (composite PRIMARY KEY constraint)
   - Upsert method (composite ON CONFLICT clause)
   - Create method (composite key support)
   - BatchCreate method (composite key support)

### Key Changes

1. **CREATE TABLE**: When multiple primary keys exist, individual columns are not marked with PRIMARY KEY inline. Instead, a table-level PRIMARY KEY constraint is added:
   ```sql
   PRIMARY KEY (col1, col2, ...)
   ```

2. **ON CONFLICT**: Composite keys are properly listed in the conflict target:
   ```sql
   ON CONFLICT (col1, col2, ...) DO UPDATE SET ...
   ```

3. **RETURNING**: For composite keys, the first primary key column is used for the RETURNING clause (this maintains API compatibility)

## Testing

The implementation has been tested with:
- ✅ Composite primary keys (2+ columns)
- ✅ Single primary keys (backward compatibility)
- ✅ Existing test suite passes
- ✅ Code generation for existing examples works correctly

