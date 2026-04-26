package similarity

import (
	"archive/zip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	ort "github.com/yalue/onnxruntime_go"
)

// ISemanticMatcher defines the interface for our NLP-based matcher.
// Compare returns a similarity score [0.0, 1.0] between two strings
// based on semantic meaning rather than character distance.
type ISemanticMatcher interface {
	Compare(s1, s2 string) float64
	Close() error
}

type MiniLMSemanticMatcher struct {
	session     *ort.DynamicAdvancedSession
	tk          *tokenizer.Tokenizer
	tempDirPath string
	cache       sync.Map // map[string][]float32 — embedding cache
}

// NewMiniLMSemanticMatcher extracts the zip and initializes the model.
// The zip must contain: model.onnx, tokenizer.json, and onnxruntime.dll (on Windows).
func NewMiniLMSemanticMatcher(zipPath string) (*MiniLMSemanticMatcher, error) {
	tempDir, err := os.MkdirTemp("", "uml_nlp_")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	err = extractZip(zipPath, tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to extract zip: %v", err)
	}

	// --- Tokenizer ---
	tkJsonPath := filepath.Join(tempDir, "tokenizer.json")
	if _, err := os.Stat(tkJsonPath); os.IsNotExist(err) {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("tokenizer.json not found in zip")
	}
	tk, err := pretrained.FromFile(tkJsonPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to load tokenizer: %v", err)
	}

	// --- ONNX Runtime (package-level init) ---
	ort.SetSharedLibraryPath(filepath.Join(tempDir, "onnxruntime.dll"))
	err = ort.InitializeEnvironment()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to init ort environment: %v", err)
	}

	modelPath := filepath.Join(tempDir, "model.onnx")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		ort.DestroyEnvironment()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("model.onnx not found in zip")
	}

	opts, err := ort.NewSessionOptions()
	if err != nil {
		ort.DestroyEnvironment()
		os.RemoveAll(tempDir)
		return nil, err
	}
	defer opts.Destroy()

	// MiniLM inputs: input_ids, attention_mask, token_type_ids
	// MiniLM output: last_hidden_state (we do mean-pooling on it)
	session, err := ort.NewDynamicAdvancedSession(
		modelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"last_hidden_state"},
		opts,
	)
	if err != nil {
		ort.DestroyEnvironment()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	return &MiniLMSemanticMatcher{
		session:     session,
		tk:          tk,
		tempDirPath: tempDir,
	}, nil
}

func (m *MiniLMSemanticMatcher) Compare(s1, s2 string) float64 {
	s1 = preprocess(s1)
	s2 = preprocess(s2)

	if s1 == s2 {
		return 1.0
	}

	v1, err := m.getEmbedding(s1)
	if err != nil {
		return 0.0
	}

	v2, err := m.getEmbedding(s2)
	if err != nil {
		return 0.0
	}

	return cosineSimilarity(v1, v2)
}

func (m *MiniLMSemanticMatcher) Close() error {
	var errs []error
	if m.session != nil {
		if err := m.session.Destroy(); err != nil {
			errs = append(errs, err)
		}
	}
	if err := ort.DestroyEnvironment(); err != nil {
		errs = append(errs, err)
	}
	if m.tempDirPath != "" {
		if err := os.RemoveAll(m.tempDirPath); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multiple close errors: %v", errs)
	}
	return nil
}

func (m *MiniLMSemanticMatcher) getEmbedding(text string) ([]float32, error) {
	// Check cache first
	if cached, ok := m.cache.Load(text); ok {
		return cached.([]float32), nil
	}

	enc, err := m.tk.EncodeSingle(text, false)
	if err != nil {
		return nil, err
	}

	seqLen := len(enc.Ids)
	if seqLen == 0 {
		return nil, fmt.Errorf("empty text after tokenization")
	}

	inputIDs := make([]int64, seqLen)
	attentionMask := make([]int64, seqLen)
	tokenTypeIDs := make([]int64, seqLen)

	for i, id := range enc.Ids {
		inputIDs[i] = int64(id)
		attentionMask[i] = 1
		tokenTypeIDs[i] = 0
	}

	shape := ort.NewShape(1, int64(seqLen))

	inTensor1, err := ort.NewTensor(shape, inputIDs)
	if err != nil {
		return nil, err
	}
	defer inTensor1.Destroy()

	inTensor2, err := ort.NewTensor(shape, attentionMask)
	if err != nil {
		return nil, err
	}
	defer inTensor2.Destroy()

	inTensor3, err := ort.NewTensor(shape, tokenTypeIDs)
	if err != nil {
		return nil, err
	}
	defer inTensor3.Destroy()

	// Run expects positional slices: inputs in order, outputs as nil for auto-alloc
	inputs := []ort.Value{inTensor1, inTensor2, inTensor3}
	outputs := []ort.Value{nil} // nil = let ONNX allocate the output

	err = m.session.Run(inputs, outputs)
	if err != nil {
		return nil, err
	}

	// Clean up auto-allocated output
	defer func() {
		if outputs[0] != nil {
			outputs[0].Destroy()
		}
	}()

	lastHiddenState, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("last_hidden_state is not a float32 tensor")
	}

	data := lastHiddenState.GetData()
	outShape := lastHiddenState.GetShape()

	// outShape: [batch_size=1, seq_len, hidden_size]
	if len(outShape) != 3 {
		return nil, fmt.Errorf("unexpected output shape length %d", len(outShape))
	}
	hiddenSize := int(outShape[2])

	// Mean Pooling over token dimension
	embedding := make([]float32, hiddenSize)
	for i := 0; i < seqLen; i++ {
		offset := i * hiddenSize
		for j := 0; j < hiddenSize; j++ {
			embedding[j] += data[offset+j]
		}
	}
	for j := 0; j < hiddenSize; j++ {
		embedding[j] /= float32(seqLen)
	}

	// Store in cache
	m.cache.Store(text, embedding)

	return embedding, nil
}

// preprocess splits camelCase/PascalCase into lowercase words.
func preprocess(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' && s[i-1] >= 'a' && s[i-1] <= 'z' {
			result.WriteRune(' ')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// cosineSimilarity calculates the cosine similarity between two float32 vectors.
func cosineSimilarity(v1, v2 []float32) float64 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0.0
	}
	var dot, mag1, mag2 float64
	for i := 0; i < len(v1); i++ {
		dot += float64(v1[i] * v2[i])
		mag1 += float64(v1[i] * v1[i])
		mag2 += float64(v2[i] * v2[i])
	}
	if mag1 == 0 || mag2 == 0 {
		return 0.0
	}
	return dot / (math.Sqrt(mag1) * math.Sqrt(mag2))
}

// extractZip extracts all files from a zip archive to destDir.
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}
