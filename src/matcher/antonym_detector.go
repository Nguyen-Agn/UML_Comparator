package matcher

// AntonymDetector checks if two tokens represent functionally opposite or mutually exclusive concepts.
// It uses both a hardcoded dictionary and dynamic prefix-based recognition.
type AntonymDetector struct {
	dict map[string]map[string]bool
}

// NewAntonymDetector initializes a detector with standard programming antonyms.
func NewAntonymDetector() *AntonymDetector {
	ad := &AntonymDetector{
		dict: make(map[string]map[string]bool),
	}

	// Hardcoded pairs of mutually exclusive domain concepts
	pairs := [][2]string{
		// Data transformations
		{"encrypt", "decrypt"},
		{"encode", "decode"},
		{"serialize", "deserialize"},
		{"pack", "unpack"},
		{"compress", "decompress"},
		{"assemble", "disassemble"},
		{"wrap", "unwrap"},
		{"marshal", "unmarshal"},

		// Access & Authentication
		{"login", "logout"},
		{"signin", "signout"},
		{"logon", "logoff"},
		{"lock", "unlock"},
		{"grant", "revoke"},
		{"allow", "deny"},

		// Standard operations
		{"start", "stop"},
		{"start", "end"},
		{"open", "close"},
		{"push", "pop"},
		{"get", "set"},
		{"read", "write"},
		{"send", "receive"},
		{"send", "fetch"},
		{"input", "output"},
		{"request", "response"},

		// CRUD & Lifecycles (Mutually exclusive, shouldn't be matched)
		{"create", "destroy"},
		{"create", "drop"},
		{"create", "delete"},
		{"create", "update"},
		{"update", "delete"},
		{"add", "remove"},
		{"add", "delete"},
		{"allocate", "deallocate"},

		// Connections & State
		{"connect", "disconnect"},
		{"attach", "detach"},
		{"mount", "unmount"},
		{"bind", "unbind"},
		{"link", "unlink"},
		{"enable", "disable"},
		{"activate", "deactivate"},
		{"register", "unregister"},
		{"subscribe", "unsubscribe"},
		{"do", "undo"},
		
		// Status / Qualifiers
		{"min", "max"},
		{"minimum", "maximum"},
		{"up", "down"},
		{"import", "export"},
		{"show", "hide"},
		{"visible", "hidden"},
		{"valid", "invalid"},
		{"sync", "async"},
		{"synchronous", "asynchronous"},
	}

	for _, p := range pairs {
		ad.add(p[0], p[1])
	}

	return ad
}

// add inserts a bi-directional antonym relationship into the dictionary
func (ad *AntonymDetector) add(w1, w2 string) {
	if ad.dict[w1] == nil {
		ad.dict[w1] = make(map[string]bool)
	}
	ad.dict[w1][w2] = true

	if ad.dict[w2] == nil {
		ad.dict[w2] = make(map[string]bool)
	}
	ad.dict[w2][w1] = true
}

// IsAntonym returns true if t1 and t2 are semantic opposites
func (ad *AntonymDetector) IsAntonym(t1, t2 string) bool {
	if t1 == t2 {
		return false
	}

	// 1. Exact Dictionary Check
	if targets, ok := ad.dict[t1]; ok {
		if targets[t2] {
			return true
		}
	}

	// 2. Dynamic Prefix Recognition 
	// Handles cases like `marshal` vs `unmarshal`, `blocking` vs `nonblocking` not explicitly in dict
	if checkDynamicPrefix(t1, t2) || checkDynamicPrefix(t2, t1) {
		return true
	}

	return false
}

// checkDynamicPrefix returns true if word == prefix + root
func checkDynamicPrefix(word, root string) bool {
	// Only apply prefix rule if the root word is substantially long enough to avoid false positives (e.g. uncle vs cle)
	if len(root) >= 3 {
		if word == "un"+root || word == "dis"+root || word == "non"+root || word == "de"+root {
			return true
		}
	}
	return false
}
