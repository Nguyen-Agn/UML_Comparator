package matcher

import (
	"math"
	"sort"
	"strings"

	"uml_compare/domain"
)

type StandardEntityMatcher struct {
	fuzzyMatcher     IFuzzyMatcher
	similarityThresh float64
}

var _ IEntityMatcher = (*StandardEntityMatcher)(nil)

// NewStandardEntityMatcher initializes a default Entity Matcher utilizing a given fuzzy Matcher submodule.
func NewStandardEntityMatcher(fz IFuzzyMatcher, threshold float64) *StandardEntityMatcher {
	return &StandardEntityMatcher{
		fuzzyMatcher:     fz,
		similarityThresh: threshold,
	}
}

// Match maps Solution nodes to Student nodes by leveraging Architecture First Sorting -> Fuzzy Match tiebreaker.
// Note: Read-only immutability is adhered to. Neither of the graphs are modified.
func (m *StandardEntityMatcher) Match(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph) (domain.MappingTable, error) {
	mapping := make(domain.MappingTable)
	if solution == nil || student == nil {
		return mapping, nil
	}

	// Keep track of which student nodes have already been mapped to enforce 1:1 mapping constraint
	studentAssigned := make(map[string]bool)

	unmappedSol := make([]int, 0, len(solution.Nodes))
	for i := range solution.Nodes {
		unmappedSol = append(unmappedSol, i)
	}

	// runPass encapsulates the intelligent matching logic with specified thresholds.
	// It returns the list of solution node indices that remain unmapped.
	runPass := func(solIndices []int, archTolerance float64, minSimScore float64) []int {
		var nextUnmapped []int
		for _, solIdx := range solIndices {
			solNode := solution.Nodes[solIdx]

			// Gather unmapped student candidates
			var candidates []studentCandidate
			for _, stuNode := range student.Nodes {
				if !studentAssigned[stuNode.ID] {

					simScore := m.fuzzyMatcher.Compare(solNode.Name, stuNode.Name)
					exactMatch := solNode.Type == stuNode.Type && strings.EqualFold(strings.TrimSpace(solNode.Name), strings.TrimSpace(stuNode.Name))
					if exactMatch {
						simScore = 1.0 // Ensure exact matches always have top score
					}

					candidates = append(candidates, studentCandidate{
						node:       stuNode,
						simScore:   simScore,
						exactMatch: exactMatch,
					})
				}
			}

			if len(candidates) == 0 {
				nextUnmapped = append(nextUnmapped, solIdx)
				continue
			}

			// Sort candidates using the 3-tier algorithm
			sort.Slice(candidates, func(i, j int) bool {
				simI := IsArchitectureSimilar(solNode.ArchWeight, candidates[i].node.ArchWeight, archTolerance)
				simJ := IsArchitectureSimilar(solNode.ArchWeight, candidates[j].node.ArchWeight, archTolerance)

				// Tier 1: Architecture Similarity Priority
				if simI != simJ {
					return simI // true comes before false
				}

				// Tier 2: Fuzzy Score Tie-breaker for Similar Architectures
				if simI { // both are true
					if candidates[i].simScore != candidates[j].simScore {
						return candidates[i].simScore > candidates[j].simScore
					}
				}

				// Tier 3: Delta Fallback
				deltaI := CalcArchDelta(solNode.ArchWeight, candidates[i].node.ArchWeight)
				deltaJ := CalcArchDelta(solNode.ArchWeight, candidates[j].node.ArchWeight)
				if deltaI != deltaJ {
					return deltaI < deltaJ
				}

				return candidates[i].simScore > candidates[j].simScore
			})

			// Test candidates in sorted order. Bind the first one that surpasses threshold or is an exact match.
			mapped := false
			for _, candidate := range candidates {
				if candidate.simScore >= minSimScore || candidate.exactMatch {
					archSim := CalcArchSim(solNode.ArchWeight, candidate.node.ArchWeight)
					finalSim := (archSim * 0.7) + (candidate.simScore * 0.3)
					finalSim = round4(finalSim)

					mapping[solNode.ID] = domain.MappedNode{
						StudentID:  candidate.node.ID,
						Similarity: finalSim,
					}
					studentAssigned[candidate.node.ID] = true
					mapped = true
					break
				}
			}

			if !mapped {
				nextUnmapped = append(nextUnmapped, solIdx)
			}
		}
		return nextUnmapped
	}

	// 1st Pass: Strict Name (>= Threshold), Lenient Architecture (15% Tolerance)
	unmappedSol = runPass(unmappedSol, 0.15, m.similarityThresh)

	// 2nd Pass: Lenient Name (>= 0.4), Strict Architecture (10% Tolerance)
	runPass(unmappedSol, 0.10, 0.4)

	return mapping, nil
}

type studentCandidate struct {
	node       domain.ProcessedNode
	simScore   float64
	exactMatch bool
}

type ArchTraits struct {
	ClassType       uint32
	HasInheritance  uint32
	NumInterfaces   uint32
	NumMethods      uint32
	NumAttributes   uint32
	NumDependencies uint32
	NumCustomTypes  uint32
	NumStaticMems   uint32
}

func UnpackArchWeight(w uint32) ArchTraits {
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

func IsArchitectureSimilar(solWeight, stuWeight uint32, tolerance float64) bool {
	if solWeight == stuWeight {
		return true
	}
	sol := UnpackArchWeight(solWeight)
	stu := UnpackArchWeight(stuWeight)

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

func ceil(a float64) float64 {
	intA := float64(int(a))
	if intA < a {
		return intA + 1
	}
	return intA
}

func CalcArchDelta(solWeight, stuWeight uint32) float64 {
	sol := UnpackArchWeight(solWeight)
	stu := UnpackArchWeight(stuWeight)
	
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

func CalcArchSim(solWeight, stuWeight uint32) float64 {
	sol := UnpackArchWeight(solWeight)
	stu := UnpackArchWeight(stuWeight)

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

func round4(val float64) float64 {
	return math.Round(val*10000) / 10000
}

func absf(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}
