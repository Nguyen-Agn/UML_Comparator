package similarity

import (
	"math"
	"testing"
)

// MockSemanticMatcher implements ISemanticMatcher for testing without ONNX DLLs
type MockSemanticMatcher struct {
	// A hardcoded mock dictionary for testing
	MockEmbeddings map[string][]float32
}

func NewMockSemanticMatcher() *MockSemanticMatcher {
	return &MockSemanticMatcher{
		MockEmbeddings: map[string][]float32{
			// "user" and "nguoi dung" should be similar
			"user":       {0.9, 0.1, 0.1},
			"nguoi dung": {0.85, 0.15, 0.1},

			// "total" and "tong" should be similar
			"total": {0.1, 0.9, 0.1},
			"tong":  {0.15, 0.85, 0.1},

			// Unrelated
			"admin":  {0.1, 0.1, 0.9},
			"create": {0.0, 0.0, 1.0},
		},
	}
}

func (m *MockSemanticMatcher) Compare(s1, s2 string) float64 {
	s1 = preprocess(s1)
	s2 = preprocess(s2)

	if s1 == s2 {
		return 1.0
	}

	v1, ok1 := m.MockEmbeddings[s1]
	v2, ok2 := m.MockEmbeddings[s2]

	if !ok1 || !ok2 {
		return 0.1 // Low similarity if not found
	}

	return cosineSimilarity(v1, v2)
}

func (m *MockSemanticMatcher) Close() error {
	return nil
}

// ---------------------------------------------------------
// Tests
// ---------------------------------------------------------

func TestPreprocess(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NguoiDung", "nguoi dung"},
		{"GetTotalValue", "get total value"},
		{"tinhTong", "tinh tong"},
		{"user", "user"},
	}

	for _, test := range tests {
		actual := preprocess(test.input)
		if actual != test.expected {
			t.Errorf("preprocess(%s) = %s; expected %s", test.input, actual, test.expected)
		}
	}
}

func TestCosineSimilarity(t *testing.T) {
	v1 := []float32{1.0, 0.0, 0.0}
	v2 := []float32{1.0, 0.0, 0.0}
	v3 := []float32{0.0, 1.0, 0.0}

	if math.Abs(cosineSimilarity(v1, v2)-1.0) > 0.001 {
		t.Errorf("Expected 1.0 for identical vectors")
	}
	if math.Abs(cosineSimilarity(v1, v3)-0.0) > 0.001 {
		t.Errorf("Expected 0.0 for orthogonal vectors")
	}
}

func TestMockSemanticMatcher(t *testing.T) {
	matcher := NewMockSemanticMatcher()

	// 1. Synonyms (Vietnamese - English)
	score1 := matcher.Compare("User", "NguoiDung")
	if score1 < 0.8 {
		t.Errorf("Expected high similarity for User/NguoiDung, got %f", score1)
	}
	t.Logf("User vs NguoiDung: %f", score1)

	// 2. Synonyms 2
	matcher.Compare("GetTotal", "TinhTong")

	scoreRoot := matcher.Compare("Total", "Tong")
	if scoreRoot < 0.8 {
		t.Errorf("Expected high similarity for Total/Tong, got %f", scoreRoot)
	}
	t.Logf("Total vs Tong: %f", scoreRoot)

	// 3. Unrelated
	score3 := matcher.Compare("User", "Create")
	if score3 > 0.5 {
		t.Errorf("Expected low similarity for User/Create, got %f", score3)
	}
	t.Logf("User vs Create: %f", score3)
}
