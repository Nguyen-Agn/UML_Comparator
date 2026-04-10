package matcher

// IIdentityValidator is the interface for validating whether two entity names 
// semantically belong to the same core concept without conflicting meanings.
type IIdentityValidator interface {
	IsValid(name1, name2 string) bool
}

// StandardIdentityValidator implements IIdentityValidator.
// It uses Tokenizer to break identifiers down and AntonymDetector to reject opposites.
type StandardIdentityValidator struct {
	antonymDetector *AntonymDetector
}

// Ensure interface implementation at compile time
var _ IIdentityValidator = (*StandardIdentityValidator)(nil)

// NewStandardIdentityValidator creates an identity validator equipped with semantic rules.
func NewStandardIdentityValidator(antonyms *AntonymDetector) *StandardIdentityValidator {
	return &StandardIdentityValidator{
		antonymDetector: antonyms,
	}
}

// IsValid validates whether name1 and name2 can be matched.
// It strictly rejects names containing contradicting tokens (e.g., Encrypt vs Decrypt).
func (v *StandardIdentityValidator) IsValid(name1, name2 string) bool {
	// 1. Exact string match shortcut
	if name1 == name2 {
		return true
	}

	// 2. Tokenize both identifiers into uniform pieces
	tokens1 := TokenizeIdentifier(name1)
	tokens2 := TokenizeIdentifier(name2)

	// Fallback to fuzzy matcher if token arrays are somehow empty
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return true 
	}

	// 3. Cross-check each token looking for Antonyms OR Contradictions
	for _, t1 := range tokens1 {
		for _, t2 := range tokens2 {
			if v.antonymDetector.IsAntonym(t1, t2) {
				return false // Immediate rejection
			}
		}
	}

	// Note: We deliberately DO NOT strictly enforce exact token overlaps here, 
	// because doing so would destroy the Fuzzy Matcher's capability to handle misspellings 
	// (e.g. "Account" -> [account] vs "Acount" -> [acount] have 0 exact token overlap).

	return true // Accepted, pass the judgment down to the Similarity Scorer
}
