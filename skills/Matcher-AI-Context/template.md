# Template: Matcher Implementation Scaffolding

## 1. Concrete Struct Definition
Tuân thủ giao tiếp Interface `IEntityMatcher` với input là `ProcessedUMLGraph`.

```go
package matcher

import (
	"uml_compare/domain"
)

type IFuzzyMatcher interface {
	Compare(s1, s2 string) float64
}

type StandardEntityMatcher struct {
	fuzzyMatcher IFuzzyMatcher
}

// Kiểm tra tính tuân thủ interface tại thời điểm compile
var _ IEntityMatcher = (*StandardEntityMatcher)(nil)

func NewStandardEntityMatcher(fz IFuzzyMatcher) *StandardEntityMatcher {
	return &StandardEntityMatcher{
		fuzzyMatcher: fz,
	}
}

func (m *StandardEntityMatcher) Match(solution *domain.ProcessedUMLGraph, student *domain.ProcessedUMLGraph) (domain.MappingTable, error) {
	mapping := make(domain.MappingTable)
	
	// 1. Phân loại Exact Match
	// 2. Với các node chưa map:
	//    a. Duyệt quaừng Node Solution
	//    b. Sort các Node Student theo độ chênh lệch abs(int64(ArchWeight) - int64(ArchWeight))
	//    c. Duyệt list đã sort, chạy Fuzzy Match (fuzzyMatcher.Compare) để kết luận (nếu Threshold > 0.8)
	
	return mapping, nil
}
```

## 2. Helper Method Pattern (Abs/Delta tính theo int64)
```go
func absDeltaU32(a, b uint32) int64 {
	diff := int64(a) - int64(b)
	if diff < 0 {
		return -diff
	}
	return diff
}
```
