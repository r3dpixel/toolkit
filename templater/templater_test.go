package templater

import (
	"fmt"
	"testing"
)

// Test with a User struct
type User struct {
	Name  string
	Email string
	Age   int
}

// Test with a Product struct
type Product struct {
	ID    string
	Name  string
	Price float64
	Stock int
}

func TestTemplaterWithUser(t *testing.T) {
	// Define tokens for User
	tokens := []Token[User]{
		&BasicToken[User]{
			Key: "{{name}}",
			Extractor: func(u User) string {
				return u.Name
			},
		},
		&BasicToken[User]{
			Key: "{{email}}",
			Extractor: func(u User) string {
				return u.Email
			},
		},
		&BasicToken[User]{
			Key: "{{age}}",
			Extractor: func(u User) string {
				return string(rune(u.Age + '0'))
			},
		},
	}

	templater := New(tokens...)

	tests := []struct {
		name     string
		template string
		user     User
		expected string
	}{
		{
			name:     "simple name replacement",
			template: "Hello, {{name}}!",
			user:     User{Name: "Alice", Email: "alice@example.com", Age: 5},
			expected: "Hello, Alice!",
		},
		{
			name:     "multiple tokens",
			template: "User: {{name}} ({{email}})",
			user:     User{Name: "Bob", Email: "bob@test.com", Age: 3},
			expected: "User: Bob (bob@test.com)",
		},
		{
			name:     "repeated tokens",
			template: "{{name}} loves {{name}}!",
			user:     User{Name: "Charlie", Email: "charlie@test.com", Age: 7},
			expected: "Charlie loves Charlie!",
		},
		{
			name:     "no tokens",
			template: "Just plain text",
			user:     User{Name: "Dave", Email: "dave@test.com", Age: 4},
			expected: "Just plain text",
		},
		{
			name:     "all tokens",
			template: "{{name}}, {{email}}, {{age}}",
			user:     User{Name: "Eve", Email: "eve@test.com", Age: 2},
			expected: "Eve, eve@test.com, 2",
		},
		{
			name:     "tokens at start and end",
			template: "{{name}} is the name {{email}}",
			user:     User{Name: "Frank", Email: "frank@test.com", Age: 6},
			expected: "Frank is the name frank@test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := templater.Execute(tt.template, tt.user)
			if result != tt.expected {
				t.Errorf("Execute() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTemplaterWithProduct(t *testing.T) {
	// Define tokens for Product
	tokens := []Token[Product]{
		&BasicToken[Product]{
			Key: "{{id}}",
			Extractor: func(p Product) string {
				return p.ID
			},
		},
		&BasicToken[Product]{
			Key: "{{name}}",
			Extractor: func(p Product) string {
				return p.Name
			},
		},
		&BasicToken[Product]{
			Key: "{{price}}",
			Extractor: func(p Product) string {
				return fmt.Sprintf("$%.2f", p.Price)
			},
		},
		&BasicToken[Product]{
			Key: "{{stock}}",
			Extractor: func(p Product) string {
				if p.Stock > 0 {
					return "In Stock"
				}
				return "Out of Stock"
			},
		},
	}

	templater := New[Product](tokens...)

	tests := []struct {
		name     string
		template string
		product  Product
		expected string
	}{
		{
			name:     "product listing",
			template: "Product: {{name}} ({{id}})",
			product:  Product{ID: "P001", Name: "Laptop", Price: 999.99, Stock: 5},
			expected: "Product: Laptop (P001)",
		},
		{
			name:     "full product details",
			template: "{{id}}: {{name}} - {{price}}, {{stock}}",
			product:  Product{ID: "P002", Name: "Mouse", Price: 29.99, Stock: 0},
			expected: "P002: Mouse - $29.99, Out of Stock",
		},
		{
			name:     "mixed text and tokens",
			template: "The {{name}} costs {{price}} and is {{stock}}",
			product:  Product{ID: "P003", Name: "Keyboard", Price: 79.99, Stock: 10},
			expected: "The Keyboard costs $79.99 and is In Stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := templater.Execute(tt.template, tt.product)
			if result != tt.expected {
				t.Errorf("Execute() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLongestMatchGreedy(t *testing.T) {
	// Test that longest match is preferred
	tokens := []Token[User]{
		&BasicToken[User]{
			Key: "{{n}}",
			Extractor: func(u User) string {
				return "SHORT"
			},
		},
		&BasicToken[User]{
			Key: "{{name}}",
			Extractor: func(u User) string {
				return u.Name
			},
		},
	}

	templater := New(tokens...)
	user := User{Name: "Alice", Email: "alice@test.com", Age: 5}

	// Should match {{name}}, not {{n}}
	result := templater.Execute("Hello {{name}}!", user)
	expected := "Hello Alice!"

	if result != expected {
		t.Errorf("Execute() = %q, want %q (should prefer longest match)", result, expected)
	}
}

func TestCompileOnce(t *testing.T) {
	// Test that compiling once and executing multiple times works
	tokens := []Token[User]{
		&BasicToken[User]{
			Key: "{{name}}",
			Extractor: func(u User) string {
				return u.Name
			},
		},
	}

	templater := New(tokens...)
	compiled := templater.Compile("Hello, {{name}}!")

	users := []User{
		{Name: "Alice", Email: "alice@test.com", Age: 5},
		{Name: "Bob", Email: "bob@test.com", Age: 3},
		{Name: "Charlie", Email: "charlie@test.com", Age: 7},
	}

	expected := []string{
		"Hello, Alice!",
		"Hello, Bob!",
		"Hello, Charlie!",
	}

	for i, user := range users {
		result := compiled.Execute(user)
		if result != expected[i] {
			t.Errorf("Execute() = %q, want %q", result, expected[i])
		}
	}
}

func TestUnicodeSupport(t *testing.T) {
	// Test Unicode characters in tokens and template
	tokens := []Token[User]{
		&BasicToken[User]{
			Key: "{{名前}}",
			Extractor: func(u User) string {
				return u.Name
			},
		},
		&BasicToken[User]{
			Key: "{{メール}}",
			Extractor: func(u User) string {
				return u.Email
			},
		},
	}

	templater := New(tokens...)
	user := User{Name: "太郎", Email: "taro@test.jp", Age: 5}

	result := templater.Execute("ユーザー: {{名前}} ({{メール}})", user)
	expected := "ユーザー: 太郎 (taro@test.jp)"

	if result != expected {
		t.Errorf("Execute() = %q, want %q", result, expected)
	}
}

func TestEmptyTemplate(t *testing.T) {
	tokens := []Token[User]{
		&BasicToken[User]{
			Key: "{{name}}",
			Extractor: func(u User) string {
				return u.Name
			},
		},
	}

	templater := New(tokens...)
	user := User{Name: "Alice", Email: "alice@test.com", Age: 5}

	result := templater.Execute("", user)
	expected := ""

	if result != expected {
		t.Errorf("Execute() = %q, want %q", result, expected)
	}
}

func TestNoMatchingTokens(t *testing.T) {
	tokens := []Token[User]{
		&BasicToken[User]{
			Key: "{{name}}",
			Extractor: func(u User) string {
				return u.Name
			},
		},
	}

	templater := New(tokens...)
	user := User{Name: "Alice", Email: "alice@test.com", Age: 5}

	// Template with no valid tokens
	result := templater.Execute("Hello {{unknown}} world!", user)
	expected := "Hello {{unknown}} world!"

	if result != expected {
		t.Errorf("Execute() = %q, want %q", result, expected)
	}
}
