package calculation

import (
	"testing"
)

// Тест построения дерева и генерации кодов Хаффмана
func TestBuildHuffmanTreeAndCodes(t *testing.T) {
	// Тест на строку с разными символами
	content := "aabbbc"

	/*
		Ожидаемая частота:
		'a':2
		'b':3
		'c':1
	*/

	tree := BuildHuffmanTree(content)
	if tree == nil {
		t.Fatal("Tree is nil for non-empty input")
	}
	if tree.Freq != len(content) {
		t.Errorf("Root node frequency should be %d, got %d", len(content), tree.Freq)
	}

	codes := GenerateHuffmanCodes(tree)
	if len(codes) != 3 {
		t.Errorf("Expected 3 codes, got %d", len(codes))
	}

	// Проверка на уникальность и на заполненость
	seen := map[string]bool{}
	for ch, code := range codes {
		if code == "" {
			t.Errorf("Code for '%c' is empty", ch)
		}
		if seen[code] {
			t.Errorf("Duplicate code found: %s", code)
		}
		seen[code] = true
	}

	// Проверка на возвращение нужной длины строки
	encoded, err := Encode(content)
	if err != nil {
		t.Errorf("Encode error: %v", err)
	}
	decLen := 0
	for _, ch := range content {
		decLen += len(codes[ch])
	}
	if len(encoded) != decLen {
		t.Errorf("Encoded length mismatch: got %d, want %d", len(encoded), decLen)
	}
}

// Тест с одним повторяющимся символом
func TestOneChar(t *testing.T) {
	content := "dddddd"
	tree := BuildHuffmanTree(content)
	if tree == nil {
		t.Fatal("Tree is nil for non-empty input")
	}
	codes := GenerateHuffmanCodes(tree)
	if len(codes) != 1 {
		t.Errorf("Should be 1 code, got %d", len(codes))
	}
	code, ok := codes['d']
	if !ok {
		t.Error("No code for 'd'")
	}
	if code != "0" {
		t.Errorf("Expected code '0', got '%s'", code)
	}
	encoded, err := Encode(content)
	if err != nil {
		t.Errorf("Encode error: %v", err)
	}
	if encoded != "000000" {
		t.Errorf("Expected '000000', got '%s'", encoded)
	}
}

// Тест с пустой строкой
func TestEmpty(t *testing.T) {
	content := ""
	tree := BuildHuffmanTree(content)
	if tree != nil {
		t.Error("Tree should be nil for empty string")
	}
	codes := GenerateHuffmanCodes(tree)
	if len(codes) != 0 {
		t.Errorf("Expected 0 codes for empty input, got %d", len(codes))
	}
	encoded, err := Encode(content)
	if err != nil {
		t.Errorf("Encode error for empty input: %v", err)
	}
	if encoded != "" {
		t.Errorf("Expected empty string, got '%s'", encoded)
	}
}
