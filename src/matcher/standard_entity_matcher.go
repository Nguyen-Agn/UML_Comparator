package matcher

import (
	"math"
	"sort"
	"strings"

	"uml_compare/domain"
)

type StandardEntityMatcher struct {
	fuzzyMatcher     IFuzzyMatcher
	archAnalyzer     IArchAnalyzer
	identValidator   IIdentityValidator
	similarityThresh float64
}

var _ IEntityMatcher = (*StandardEntityMatcher)(nil)

// NewStandardEntityMatcher initializes a default Entity Matcher utilizing a given fuzzy Matcher, architecture analyzer, and identity validator.
func NewStandardEntityMatcher(threshold float64) *StandardEntityMatcher {
	return &StandardEntityMatcher{
		fuzzyMatcher:     NewLevenshteinMatcher(),
		archAnalyzer:     NewStandardArchAnalyzer(),
		identValidator:   NewStandardIdentityValidator(NewAntonymDetector()),
		similarityThresh: threshold,
	}
}

// Match maps Solution nodes (OR-aware SolutionProcessedUMLGraph) to Student nodes (ProcessedUMLGraph).
// Matching is driven by the 3-tier algorithm: Architecture → Fuzzy Name → Arch Delta.
// Note: Read-only immutability is adhered to. Neither graph is modified.
func (m *StandardEntityMatcher) Match(solution *domain.SolutionProcessedUMLGraph, student *domain.ProcessedUMLGraph) (domain.MappingTable, error) {
	mapping := make(domain.MappingTable)
	if solution == nil || student == nil {
		return mapping, nil
	}

	// Keep track of which student nodes have already been mapped to enforce 1:1 constraint
	studentAssigned := make(map[string]bool)

	unmappedSol := make([]int, 0, len(solution.Nodes))
	for i := range solution.Nodes {
		unmappedSol = append(unmappedSol, i)
	}

	// runPass encapsulates the intelligent matching logic with specified thresholds.
	// solNode is a SolutionProcessedNode — we use its primary name (Names[0]) for fuzzy matching
	// and its ArchWeight for architecture comparison. OR-alternatives in Names are checked as
	// bonus comparators: if any alternative fully matches, simScore becomes 1.0.
	runPass := func(solIndices []int, archTolerance float64, minSimScore float64) []int {
		var nextUnmapped []int
		for _, solIdx := range solIndices {
			solNode := solution.Nodes[solIdx]

			// Gather unmapped student candidates
			var candidates []studentCandidate
			for _, stuNode := range student.Nodes {
				if !studentAssigned[stuNode.ID] {

					// Semantic Identity Check (Filter out antonyms early)
					if m.identValidator != nil && !m.identValidator.IsValid(solNode.Name, stuNode.Name) {
						continue
					}

					simScore := m.fuzzyMatcher.Compare(solNode.Name, stuNode.Name)
					exactMatch := solNode.Type == stuNode.Type &&
						strings.EqualFold(strings.TrimSpace(solNode.Name), strings.TrimSpace(stuNode.Name))

					if exactMatch {
						simScore = 1.0
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
				simI := m.archAnalyzer.IsArchitectureSimilar(solNode.ArchWeight, candidates[i].node.ArchWeight, archTolerance)
				simJ := m.archAnalyzer.IsArchitectureSimilar(solNode.ArchWeight, candidates[j].node.ArchWeight, archTolerance)

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
				deltaI := m.archAnalyzer.CalcArchDelta(solNode.ArchWeight, candidates[i].node.ArchWeight)
				deltaJ := m.archAnalyzer.CalcArchDelta(solNode.ArchWeight, candidates[j].node.ArchWeight)
				if deltaI != deltaJ {
					return deltaI < deltaJ
				}

				return candidates[i].simScore > candidates[j].simScore
			})

			// Test candidates in sorted order. Bind the first one that surpasses threshold or is an exact match.
			mapped := false
			for _, candidate := range candidates {
				if candidate.simScore >= minSimScore || candidate.exactMatch {
					archSim := m.archAnalyzer.CalcArchSim(solNode.ArchWeight, candidate.node.ArchWeight)
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

func round4(val float64) float64 {
	return math.Round(val*10000) / 10000
}
