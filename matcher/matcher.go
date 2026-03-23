package matcher

import "uml_compare/domain"

// IEntityMatcher defines the contract for matching nodes between the solution and the student's diagram.
type IEntityMatcher interface {
	// Match compares two graphs and produces a dictionary mapping solution node IDs to student node IDs.
	Match(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph) (domain.MappingTable, error)
}

// IArchAnalyzer defines the contract for analyzing and comparing node architectures based on their ArchWeight.
type IArchAnalyzer interface {
	// UnpackArchWeight decomposes a raw uint32 ArchWeight into a readable ArchTraits struct.
	UnpackArchWeight(w uint32) ArchTraits
	// IsArchitectureSimilar checks if two ArchWeights are structurally compatible within a given tolerance.
	IsArchitectureSimilar(solWeight, stuWeight uint32, tolerance float64) bool
	// CalcArchDelta calculates a numerical difference (error score) between two ArchWeights.
	CalcArchDelta(solWeight, stuWeight uint32) float64
	// CalcArchSim calculates a similarity score between 0.0 and 1.0 based on structural differences.
	CalcArchSim(solWeight, stuWeight uint32) float64
}

// ArchTraits represents the decomposed structural characteristics of a UML node.
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
