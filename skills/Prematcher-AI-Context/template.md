# Template: Prematcher Implementation Scaffolding

## 1. Concrete Struct Definition
Mọi implementation của `IPreMatcher` nên tuân thủ cấu trúc sau:

```go
package prematcher

import "uml_compare/domain"

type StandardPreMatcher struct {
	// Các internal helpers nếu cần
}

// Kiểm tra tính tuân thủ interface tại thời điểm compile
var _ IPreMatcher = (*StandardPreMatcher)(nil)

func NewStandardPreMatcher() *StandardPreMatcher {
	return &StandardPreMatcher{}
}

func (p *StandardPreMatcher) Process(graph *domain.UMLGraph) (*domain.ProcessedUMLGraph, error) {
	// 1. Khởi tạo ProcessedUMLGraph
	// 2. Map Nodes
	// 3. Map Edges
	// 4. Return
	return nil, nil
}
```

## 2. Helper Method Pattern
Dùng regex hoặc string splitting cho việc parse member:

```go
func parseAttribute(raw string) domain.ProcessedAttribute {
    // Regex: ^(?P<scope>[+\-#]?)\s*(?P<name>[^:]+)\s*:\s*(?P<type>.*)$
    return domain.ProcessedAttribute{}
}
```
