package slug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate_BasicString(t *testing.T) {
	input := "Hello World"
	expected := "hello-world"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithSpecialCharacters(t *testing.T) {
	input := "Hello, World! How are you?"
	expected := "hello-world-how-are-you"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithNumbers(t *testing.T) {
	input := "Article 123 Title"
	expected := "article-123-title"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithMultipleSpaces(t *testing.T) {
	input := "Hello    World"
	expected := "hello-world"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithHyphens(t *testing.T) {
	input := "Hello-World-Test"
	expected := "hello-world-test"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithLeadingTrailingSpaces(t *testing.T) {
	input := "  Hello World  "
	expected := "hello-world"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithUnicode(t *testing.T) {
	input := "Café Résumé"
	// Unicode characters are removed
	expected := "caf-rsum"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_OnlySpecialCharacters(t *testing.T) {
	input := "!@#$%^&*()"
	expected := ""

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_EmptyString(t *testing.T) {
	input := ""
	expected := ""

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_AlreadySlug(t *testing.T) {
	input := "already-a-slug"
	expected := "already-a-slug"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_MixedCase(t *testing.T) {
	input := "HeLLo WoRLd"
	expected := "hello-world"

	result := Generate(input)
	assert.Equal(t, expected, result)
}

func TestGenerate_WithUnderscores(t *testing.T) {
	input := "hello_world_test"
	expected := "helloworldtest"

	result := Generate(input)
	assert.Equal(t, expected, result)
}
