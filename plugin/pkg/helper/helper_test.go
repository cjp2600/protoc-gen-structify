package helper

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"

	"github.com/stretchr/testify/assert"
)

func TestGoTypeToPostgresType(t *testing.T) {
	tests := []struct {
		goType       string
		postgresType string
	}{
		{"string", "TEXT"},
		{"*string", "TEXT"},
		{"bool", "BOOLEAN"},
		{"int", "INTEGER"},
		{"int32", "INTEGER"},
		{"int64", "BIGINT"},
		{"float32", "REAL"},
		{"float64", "DOUBLE PRECISION"},
		{"time.Time", "TIMESTAMP"},
		{"[]byte", "BYTEA"},
		{"CustomType", "TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.goType, func(t *testing.T) {
			result := GoTypeToPostgresType(tt.goType)
			assert.Equal(t, tt.postgresType, result)
		})
	}
}

func TestExecuteTemplate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpl := "Hello, {{.Name}}!"
		data := map[string]string{"Name": "World"}
		result, err := ExecuteTemplate(tmpl, nil, data)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("With Included Templates", func(t *testing.T) {
		tmpl := "Hello, {{template \"name\" .}}!"
		include := IncludeTemplate{Name: "name", Body: "{{.Name}}"}
		data := map[string]string{"Name": "World"}
		result, err := ExecuteTemplate(tmpl, nil, data, include)
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", result)
	})

	t.Run("Parse Error", func(t *testing.T) {
		tmpl := "Hello, {{.Name}!"
		data := map[string]string{"Name": "World"}
		_, err := ExecuteTemplate(tmpl, nil, data)
		assert.Error(t, err)
	})
}

func TestClearPointer(t *testing.T) {
	// Test cases representing different scenarios
	tests := []struct {
		input  string
		output string
	}{
		{"*string", "string"},       // Test with single pointer
		{"[]int", "int"},            // Test with slice notation
		{"[]*bool", "bool"},         // Test with both slice notation and pointer
		{"time.Time", "time.Time"},  // Test with no pointer or slice notation
		{"[][]*float64", "float64"}, // Test with multiple slice notations and pointer
	}

	// Iterate through each test case
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			// Call ClearPointer function with the input string
			result := ClearPointer(test.input)

			// Check that the result matches the expected output
			assert.Equal(t, test.output, result)
		})
	}
}

// TestConvertType tests the ConvertType function
func TestConvertType(t *testing.T) {
	// Test cases representing different field types
	tm := "google.protobuf.Timestamp"
	tests := []struct {
		field    *descriptor.FieldDescriptorProto // Input field descriptor
		expected string                           // Expected converted type
	}{
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_DOUBLE.Enum(),
			},
			expected: "float64",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_FLOAT.Enum(),
			},
			expected: "float32",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_INT32.Enum(),
			},
			expected: "int32",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type:  descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label: descriptor.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			},
			expected: "[]string",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type:  descriptor.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label: descriptor.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			},
			expected: "[]*",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type:  descriptor.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				Label: descriptor.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			expected: "*",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_INT32.Enum(),
			},
			expected: "int32",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_INT64.Enum(),
			},
			expected: "int64",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_UINT32.Enum(),
			},
			expected: "uint32",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_UINT64.Enum(),
			},
			expected: "uint64",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_BOOL.Enum(),
			},
			expected: "bool",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
			},
			expected: "string",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_GROUP.Enum(),
			},
			expected: "error",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_BYTES.Enum(),
			},
			expected: "[]byte",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type:     descriptor.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
				TypeName: &tm,
			},
			expected: "time.Time",
		},
	}

	// Iterate through each test case
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			// Call ConvertType function with the field descriptor
			result := ConvertType(test.field)

			// Check that the result matches the expected output
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTypePrefix(t *testing.T) {
	// Test cases representing different scenarios
	tests := []struct {
		field    *descriptor.FieldDescriptorProto // Input field descriptor
		typeName string                           // Input type name
		expected string                           // Expected result with prefixes
	}{
		// Test with non-repeated, non-optional field
		{
			field: &descriptor.FieldDescriptorProto{
				Type: descriptor.FieldDescriptorProto_TYPE_INT32.Enum(),
			},
			typeName: "int32",
			expected: "int32",
		},
		// Test with repeated field
		{
			field: &descriptor.FieldDescriptorProto{
				Type:  descriptor.FieldDescriptorProto_TYPE_STRING.Enum(),
				Label: descriptor.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			},
			typeName: "string",
			expected: "[]string",
		},
		// Test with optional field
		{
			field: &descriptor.FieldDescriptorProto{
				Type:  descriptor.FieldDescriptorProto_TYPE_BOOL.Enum(),
				Label: descriptor.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
			},
			typeName: "bool",
			expected: "bool",
		},
		{
			field: &descriptor.FieldDescriptorProto{
				Type:  descriptor.FieldDescriptorProto_TYPE_FLOAT.Enum(),
				Label: descriptor.FieldDescriptorProto_LABEL_REPEATED.Enum(),
			},
			typeName: "float32",
			expected: "[]float32",
		},
	}

	// Iterate through each test case
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			// Call TypePrefix function with the field descriptor and type name
			result := TypePrefix(test.field, test.typeName)

			// Check that the result matches the expected output
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestFirstLetterLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      error
	}{
		// Test with an empty string
		{
			input:    "",
			expected: "",
			err:      fmt.Errorf("string is empty"),
		},
		// Test with a lowercase starting letter
		{
			input:    "apple",
			expected: "a",
			err:      nil,
		},
		// Test with an uppercase starting letter
		{
			input:    "Banana",
			expected: "b",
			err:      nil,
		},
		// Test with a number
		{
			input:    "1fruit",
			expected: "1",
			err:      nil,
		},
		// Test with a special character
		{
			input:    "@home",
			expected: "@",
			err:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			output, err := FirstLetterLower(tt.input)
			assert.Equal(t, tt.expected, output)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestSliceToString(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		// Test with an empty slice
		{
			input:    []string{},
			expected: "[]string{}",
		},
		// Test with a single element
		{
			input:    []string{"apple"},
			expected: `[]string{"apple"}`,
		},
		// Test with multiple elements
		{
			input:    []string{"apple", "banana", "cherry"},
			expected: `[]string{"apple", "banana", "cherry"}`,
		},
		// Test with special characters
		{
			input:    []string{"a\"pple", "b@nana"},
			expected: `[]string{"a\"pple", "b@nana"}`,
		},
		// Test with mixed types of strings
		{
			input:    []string{"123", "true", "apple"},
			expected: `[]string{"123", "true", "apple"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			output := SliceToString(tt.input)
			assert.Equal(t, tt.expected, output)
		})
	}
}
