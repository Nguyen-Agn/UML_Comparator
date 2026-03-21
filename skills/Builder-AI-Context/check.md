# Verification Checklist — Builder Module

## Code Architecture
- [ ] `StandardModelBuilder` struct fields có type là **interface** (`IXMLParser`, `ITextSanitizer`, `ITypeDetector`, `IMemberParser`), không phải concrete struct
- [ ] Mỗi concrete struct có `var _ IXxx = (*xxxStruct)(nil)` compile-time check
- [ ] `typeDetector.san` field có type `ITextSanitizer` (không phải `*htmlSanitizer`)
- [ ] `memberParser.san` field có type `ITextSanitizer` (không phải `*htmlSanitizer`)

## Test Coverage
- [ ] `go build ./builder/...` — zero errors
- [ ] `go test ./builder/... -v` — 17/17 PASS (9 incorrect + 8 unit tests)
- [ ] Không có `--- FAIL:` trong output

## Output Correctness
- [ ] Node.Name không chứa `<<stereotype>>` thừa
- [ ] Node.Type thuộc `{Class, Interface, Abstract, Actor, Enum}`
- [ ] Edge.SourceID và TargetID resolve về top-level class ID (không phải child cell ID)
- [ ] Multi-line method signature được join thành 1 dòng duy nhất
