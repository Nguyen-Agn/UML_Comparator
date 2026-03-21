# Pattern So sánh (Advanced Comparator)
Tách hàm chi tiết để Compare không trở thành hàm God Function:
```go
// 1. Dịch type thông minh
func (c *StandardComparator) translateType(typeName string, typeMap map[string]string) string

// 2. Tách Getter/Setter với Normal/Constructor
func (c *StandardComparator) splitMethods(methods []ProcessedMethod) (gs, normal []ProcessedMethod)

// 3. Match theo quy tắc Scope tuyệt đối và Flexible params
func (c *StandardComparator) matchMethod(sol, stu ProcessedMethod, typeMap map[string]string, isCtor bool, stuClassName string) bool
```
