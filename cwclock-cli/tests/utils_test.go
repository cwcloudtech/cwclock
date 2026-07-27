package test

import (
	"bytes"
	"cwclock/utils"
	"encoding/json"
	"fmt"
	"testing"
)

type TestStruct struct {
	Name  string
	Value int
}

func TestIf(t *testing.T) {
	tests := []struct {
		name     string
		cond     bool
		vtrue    interface{}
		vfalse   interface{}
		expected interface{}
	}{
		{"true condition string", true, "yes", "no", "yes"},
		{"false condition string", false, "yes", "no", "no"},
		{"true condition int", true, 1, 0, 1},
		{"false condition int", false, 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.If(tt.cond, tt.vtrue, tt.vfalse)
			if result != tt.expected {
				t.Errorf("If(%v, %v, %v) = %v; want %v", tt.cond, tt.vtrue, tt.vfalse, result, tt.expected)
			}
		})
	}
}

func TestIsNotBlank(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{utils.EMPTY, false},
		{" ", false},
		{"  ", false},
		{"hello", true},
		{" hello ", true},
	}

	for _, tt := range tests {
		if got := utils.IsNotBlank(tt.input); got != tt.expected {
			t.Errorf("IsNotBlank(%q) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", utils.EMPTY, true},
		{"non-empty string", "hello", false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"hello"}, false},
		{"zero int", 0, true},
		{"non-zero int", 1, false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"empty struct", TestStruct{}, true},
		{"non-empty struct", TestStruct{Name: "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.IsEmpty(tt.input); got != tt.expected {
				t.Errorf("IsEmpty(%v) = %v; want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	slice := []string{"apple", "banana", "orange"}
	tests := []struct {
		input    string
		expected bool
	}{
		{"apple", true},
		{"banana", true},
		{"grape", false},
		{utils.EMPTY, false},
	}

	for _, tt := range tests {
		if got := utils.StringInSlice(tt.input, slice); got != tt.expected {
			t.Errorf("StringInSlice(%q, %v) = %v; want %v", tt.input, slice, got, tt.expected)
		}
	}
}

func TestJsonPrettyPrint(t *testing.T) {
	input := `{"name":"test","value":123}`
	var expectedFormatted bytes.Buffer
	json.Indent(&expectedFormatted, []byte(input), utils.EMPTY, "\t")

	if got := utils.JsonPrettyPrint(input); got != expectedFormatted.String() {
		t.Errorf("JsonPrettyPrint() formatting mismatch:\ngot:\n%s\nwant:\n%s", got, expectedFormatted.String())
	}
}

func TestShortName(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected string
	}{
		{"deployment-abc123", "abc123", "deployment"},
		{"deployment-abc123", utils.EMPTY, "deployment"},
		{"deployment", "abc123", "deployment"},
		{utils.EMPTY, "abc123", utils.EMPTY},
		{"deployment", utils.EMPTY, "deployment"},
		{"my-app-abc123", "abc123", "my-app"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("name=%s,hash=%s", tt.name, tt.hash), func(t *testing.T) {
			if got := utils.ShortName(tt.name, tt.hash); got != tt.expected {
				t.Errorf("ShortName(%q, %q) = %q; want %q", tt.name, tt.hash, got, tt.expected)
			}
		})
	}
}
