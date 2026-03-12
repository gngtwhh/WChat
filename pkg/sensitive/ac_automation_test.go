package sensitive

import (
	"testing"
)

func TestNewACFilter(t *testing.T) {
	ac := NewACFilter()
	if ac == nil {
		t.Fatal("NewACFilter returned nil")
	}
	if ac.root == nil {
		t.Fatal("root node should not be nil")
	}
	if len(ac.root.children) != 0 {
		t.Fatal("root should have no children initially")
	}
}

func TestACFilter_Build(t *testing.T) {
	ac := NewACFilter()
	words := []string{"hello", "world", "test"}
	ac.Build(words)

	if ac.root == nil {
		t.Fatal("root should not be nil after build")
	}
}

func TestACFilter_Filter_Basic(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		text     string
		expected string
	}{
		{
			name:     "single word",
			words:    []string{"bad"},
			text:     "This is a bad word",
			expected: "This is a *** word",
		},
		{
			name:     "multiple words",
			words:    []string{"bad", "evil"},
			text:     "This is bad and evil",
			expected: "This is *** and ****",
		},
		{
			name:     "no match",
			words:    []string{"bad"},
			text:     "This is good",
			expected: "This is good",
		},
		{
			name:     "empty text",
			words:    []string{"bad"},
			text:     "",
			expected: "",
		},
		{
			name:     "empty word list",
			words:    []string{},
			text:     "This is bad",
			expected: "This is bad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewACFilter()
			ac.Build(tt.words)
			result := ac.Filter(tt.text)
			if result != tt.expected {
				t.Errorf("Filter(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestACFilter_Filter_Overlap(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		text     string
		expected string
	}{
		{
			name:     "overlapping words",
			words:    []string{"abc", "bc"},
			text:     "abcd",
			expected: "***d",
		},
		{
			name:     "contained words",
			words:    []string{"he", "hell", "hello"},
			text:     "hello world",
			expected: "***** world",
		},
		{
			name:     "adjacent matches",
			words:    []string{"ab", "cd"},
			text:     "abcd",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewACFilter()
			ac.Build(tt.words)
			result := ac.Filter(tt.text)
			if result != tt.expected {
				t.Errorf("Filter(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestACFilter_Filter_Complex(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		text     string
		expected string
	}{
		{
			name:     "long text with multiple matches",
			words:    []string{"bad", "evil", "wrong"},
			text:     "This is a bad example with evil intentions and wrong decisions",
			expected: "This is a *** example with **** intentions and ***** decisions",
		},
		{
			name:     "case sensitive",
			words:    []string{"bad"},
			text:     "Bad BAD bad",
			expected: "Bad BAD ***",
		},
		{
			name:     "special characters",
			words:    []string{"test@123", "#$%"},
			text:     "My password is test@123 and symbols #$%",
			expected: "My password is ******** and symbols ***",
		},
		{
			name:     "unicode characters",
			words:    []string{"测试", "敏感词"},
			text:     "这是一个测试敏感词的例子",
			expected: "这是一个*****的例子",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewACFilter()
			ac.Build(tt.words)
			result := ac.Filter(tt.text)
			if result != tt.expected {
				t.Errorf("Filter(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestACFilter_Filter_Boundary(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		text     string
		expected string
	}{
		{
			name:     "word at start",
			words:    []string{"bad"},
			text:     "bad start",
			expected: "*** start",
		},
		{
			name:     "word at end",
			words:    []string{"bad"},
			text:     "end bad",
			expected: "end ***",
		},
		{
			name:     "whole text is word",
			words:    []string{"bad"},
			text:     "bad",
			expected: "***",
		},
		{
			name:     "repeated words",
			words:    []string{"bad"},
			text:     "bad bad bad",
			expected: "*** *** ***",
		},
		{
			name:     "spaces in word",
			words:    []string{"bad word"},
			text:     "This is a bad word test",
			expected: "This is a ******** test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewACFilter()
			ac.Build(tt.words)
			result := ac.Filter(tt.text)
			if result != tt.expected {
				t.Errorf("Filter(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestACFilter_Filter_Concurrent(t *testing.T) {
	ac := NewACFilter()
	ac.Build([]string{"test", "bad", "evil"})

	texts := []string{
		"This is a test",
		"Bad things happen",
		"Evil is everywhere",
		"No match here",
	}

	expected := []string{
		"This is a ****",
		"Bad things happen",
		"Evil is everywhere",
		"No match here",
	}

	// 简单并发测试
	done := make(chan bool, len(texts))
	for i, text := range texts {
		go func(idx int, txt string) {
			result := ac.Filter(txt)
			if result != expected[idx] {
				t.Errorf("Concurrent test failed: Filter(%q) = %q, want %q",
					txt, result, expected[idx])
			}
			done <- true
		}(i, text)
	}

	// 等待所有goroutine完成
	for _ = range texts {
		<-done
	}
}

func TestACFilter_Filter_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		words    []string
		text     string
		expected string
	}{
		{
			name:     "empty word",
			words:    []string{""},
			text:     "This is a test",
			expected: "This is a test",
		},
		{
			name:     "multiple spaces",
			words:    []string{"bad"},
			text:     "bad   bad",
			expected: "***   ***",
		},
		{
			name:     "newlines in text",
			words:    []string{"bad"},
			text:     "This is bad\nand also bad",
			expected: "This is ***\nand also ***",
		},
		{
			name:     "tab characters",
			words:    []string{"bad"},
			text:     "bad\tword",
			expected: "***\tword",
		},
		{
			name:     "mixed whitespace",
			words:    []string{"bad"},
			text:     "bad \t\n bad",
			expected: "*** \t\n ***",
		},
		{
			name:     "very long word",
			words:    []string{"verylongsensitiveword"},
			text:     "This is verylongsensitiveword test",
			expected: "This is ********************* test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewACFilter()
			ac.Build(tt.words)
			result := ac.Filter(tt.text)
			if result != tt.expected {
				t.Errorf("Filter(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}
