package matcher

// StandardArchAnalyzer implements IArchAnalyzer using the specific ArchWeight bitfield layout.
type StandardArchAnalyzer struct{}

var _ IArchAnalyzer = (*StandardArchAnalyzer)(nil)

// NewStandardArchAnalyzer creates a new instance of StandardArchAnalyzer.
func NewStandardArchAnalyzer() *StandardArchAnalyzer {
	return &StandardArchAnalyzer{}
}

// UnpackArchWeight decomposes a raw uint32 ArchWeight into a readable ArchTraits struct.
// It extracts node type, inheritance status, and various member counts from the bitfield.
func (a *StandardArchAnalyzer) UnpackArchWeight(w uint32) ArchTraits {
	return ArchTraits{
		ClassType:       (w >> 29) & 0x7,
		HasInheritance:  (w >> 28) & 0x1,
		NumInterfaces:   (w >> 24) & 0xF,
		NumMethods:      (w >> 18) & 0x3F,
		NumAttributes:   (w >> 13) & 0x1F,
		NumDependencies: (w >> 9) & 0xF,
		NumCustomTypes:  (w >> 6) & 0x7,
		NumStaticMems:   (w >> 2) & 0xF,
	}
}

// IsArchitectureSimilar checks if two ArchWeights are structurally compatible within a given tolerance.
// Certain fields (ClassType, Inheritance, Interfaces, CustomTypes) must match exactly.
// Other fields (Methods, Attributes, Dependencies, StaticMems) can vary within the specified percentage.
func (a *StandardArchAnalyzer) IsArchitectureSimilar(solWeight, stuWeight uint32, tolerance float64) bool {
	if solWeight == stuWeight {
		return true
	}
	sol := a.UnpackArchWeight(solWeight)
	stu := a.UnpackArchWeight(stuWeight)

	// Exact match fields
	if sol.ClassType != stu.ClassType ||
		sol.HasInheritance != stu.HasInheritance ||
		sol.NumInterfaces != stu.NumInterfaces ||
		sol.NumCustomTypes != stu.NumCustomTypes {
		return false
	}

	// Dynamic tolerance fields
	if !isTol(sol.NumMethods, stu.NumMethods, tolerance) ||
		!isTol(sol.NumAttributes, stu.NumAttributes, tolerance) ||
		!isTol(sol.NumDependencies, stu.NumDependencies, tolerance) ||
		!isTol(sol.NumStaticMems, stu.NumStaticMems, tolerance) {
		return false
	}

	return true
}

// CalcArchDelta calculates a numerical difference (error score) between two ArchWeights.
// It applies a heavy penalty for mismatched core structures and a linear penalty for member count differences.
func (a *StandardArchAnalyzer) CalcArchDelta(solWeight, stuWeight uint32) float64 {
	sol := a.UnpackArchWeight(solWeight)
	stu := a.UnpackArchWeight(stuWeight)

	var diff float64
	diff += absf(float64(sol.NumMethods) - float64(stu.NumMethods))
	diff += absf(float64(sol.NumAttributes) - float64(stu.NumAttributes))
	diff += absf(float64(sol.NumDependencies) - float64(stu.NumDependencies))
	diff += absf(float64(sol.NumStaticMems) - float64(stu.NumStaticMems))

	// Heavy penalty for non-matching exact fields
	if sol.ClassType != stu.ClassType {
		diff += 1000
	}
	if sol.HasInheritance != stu.HasInheritance {
		diff += 1000
	}
	if sol.NumInterfaces != stu.NumInterfaces {
		diff += 1000
	}
	if sol.NumCustomTypes != stu.NumCustomTypes {
		diff += 1000
	}

	return diff
}

// CalcArchSim calculates a similarity score between 0.0 and 1.0 based on structural differences.
// It normalizes the difference against the total expected elements in the solution node.
func (a *StandardArchAnalyzer) CalcArchSim(solWeight, stuWeight uint32) float64 {
	sol := a.UnpackArchWeight(solWeight)
	stu := a.UnpackArchWeight(stuWeight)

	var totalSol float64
	totalSol += float64(sol.NumMethods + sol.NumAttributes + sol.NumDependencies + sol.NumStaticMems)

	var diff float64
	diff += absf(float64(sol.NumMethods) - float64(stu.NumMethods))
	diff += absf(float64(sol.NumAttributes) - float64(stu.NumAttributes))
	diff += absf(float64(sol.NumDependencies) - float64(stu.NumDependencies))
	diff += absf(float64(sol.NumStaticMems) - float64(stu.NumStaticMems))

	if totalSol == 0 {
		if diff == 0 {
			return 1.0
		}
		return 0.0
	}

	sim := 1.0 - (diff / totalSol)
	if sim < 0 {
		return 0.0
	}
	return sim
}

// isTol is a private helper to check if a value is within a given tolerance percentage.
func isTol(sol, stu uint32, tolerance float64) bool {
	if sol == stu {
		return true
	}
	diff := float64(sol) - float64(stu)
	if diff < 0 {
		diff = -diff
	}
	allowedDiff := ceil(float64(sol) * tolerance)
	return diff <= allowedDiff
}

// ceil is a private helper that performs traditional ceiling on float64.
func ceil(a float64) float64 {
	intA := float64(int(a))
	if intA < a {
		return intA + 1
	}
	return intA
}

// absf returns the absolute value of a float64.
func absf(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}
