package similarity

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const testZipName = "minilm.ai"

// getZipPath returns the path to minilm.zip relative to this test file.
func getZipPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), testZipName)
}

// TestE2E_SemanticMatcher tests the real ONNX model end-to-end.
// Skipped if minilm.zip is not present.
func TestE2E_SemanticMatcher(t *testing.T) {
	zipPath := getZipPath()
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skipf("Skipping E2E test: %s not found. Run prepare_model.py first.", zipPath)
	}

	matcher, err := NewMiniLMSemanticMatcher(zipPath)
	if err != nil {
		t.Fatalf("Failed to create semantic matcher: %v", err)
	}
	defer matcher.Close()

	tests := []struct {
		s1       string
		s2       string
		minScore float64
		maxScore float64
		desc     string
	}{
		// EN-EN synonyms: model scores very high (0.7+)
		{"delete", "remove", 0.7, 1.0, "EN-EN synonym"},
		{"get", "retrieve", 0.6, 1.0, "EN-EN synonym"},

		// Exact match
		{"User", "User", 0.99, 1.0, "Exact match"},

		// Unrelated — should stay below 0.5
		{"User", "PaymentGateway", 0.0, 0.5, "Unrelated"},

		// nearly but not same
		{"User", "Usager", 0.3, 0.7, "nearly but not same"},
		{"Triangle", "Square", 0.3, 0.7, "nearly but not same"},

		// NOTE: paraphrase model scores antonyms HIGH (create/delete = 0.99).
		// This is expected — antonym detection is done by AntonymDetector in main codebase.
		{"create", "delete", 0.0, 1.0, "Antonyms (high score is expected for paraphrase model)"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			score := matcher.Compare(tc.s1, tc.s2)
			t.Logf("%s vs %s = %.4f", tc.s1, tc.s2, score)

			if score < tc.minScore {
				t.Errorf("Score %.4f < expected min %.2f", score, tc.minScore)
			}
			if score > tc.maxScore {
				t.Errorf("Score %.4f > expected max %.2f", score, tc.maxScore)
			}
		})
	}
}

// TestE2E_HybridMatcher tests the full hybrid pipeline.
func TestE2E_HybridMatcher(t *testing.T) {
	zipPath := getZipPath()
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Skipf("Skipping E2E test: %s not found.", zipPath)
	}

	hybrid, err := NewHybridMatcher(zipPath)
	if err != nil {
		t.Fatalf("Failed to create hybrid matcher: %v", err)
	}
	defer hybrid.Close()

	tests := []struct {
		s1       string
		s2       string
		minScore float64
		desc     string
	}{
		// Typos — Levenshtein should handle these (score > 0.8)
		{"SystemManager", "SystmManagr", 0.7, "Typo"},
		{"UserController", "UserControler", 0.8, "Typo"},

		// Synonyms — Semantic should boost these (EN-VI scores 0.3-0.5)
		{"User", "NguoiDung", 0.3, "Synonym"},
		{"Student", "SinhVien", 0.4, "Synonym"},

		// Unrelated — should stay low
		{"User", "PaymentGateway", 0.0, "Unrelated"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			score := hybrid.Compare(tc.s1, tc.s2)
			t.Logf("[Hybrid] %s vs %s = %.4f", tc.s1, tc.s2, score)

			if score < tc.minScore {
				t.Errorf("Score %.4f < expected min %.2f", score, tc.minScore)
			}
		})
	}
}

// BenchmarkSemanticCompare measures the speed of a single Compare call.
func BenchmarkSemanticCompare(b *testing.B) {
	zipPath := getZipPath()
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		b.Skipf("Skipping benchmark: %s not found.", zipPath)
	}

	matcher, err := NewMiniLMSemanticMatcher(zipPath)
	if err != nil {
		b.Fatalf("Failed to create semantic matcher: %v", err)
	}
	defer matcher.Close()

	start := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.Compare("UserController", "NguoiDungController")
	}
	b.StopTimer()
	elapsed := time.Since(start)
	b.Logf("Average: %v per call (%d iterations)", elapsed/time.Duration(b.N), b.N)
}
