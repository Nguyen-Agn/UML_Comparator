# SKILL: uml-model-builder
## Purpose
Chuyển đổi chuỗi RawXMLData thành cấu trúc domain.UMLGraph thống nhất (bao gồm trích xuất Nodes, Edges, Attributes, Methods).

## Required Inputs
- domain.RawXMLData (chuỗi XML thô từ DrawioParser)

## Expected Output
- *domain.UMLGraph (Mô hình OOP hoàn chỉnh)

## Architecture (Interface-Based — DIP)
`StandardModelBuilder` là orchestrator thuần tuý, **chỉ phụ thuộc vào 4 internal interfaces**:

| Interface | File impl | Trách nhiệm |
|---|---|---|
| `IXMLParser` | `xml_parser.go` | Parse XML → `[]mxCell`, structural queries |
| `ITextSanitizer` | `html_sanitizer.go` | Decode HTML entities, strip tags/stereotypes |
| `ITypeDetector` | `type_detector.go` | Phân loại Class/Interface/Abstract/Actor/Enum + RelationType |
| `IMemberParser` | `member_parser.go` | Tách Attributes vs Methods từ child cells |

Xem `builder/interfaces.go` để biết contract đầy đủ.
Xem `builder/builder.mmd` để thấy sơ đồ class diagram.

## Execution Approach
1. `cellParser.parse()` → `[]mxCell` từ XML string
2. Index cells (rootLayerID, cellMap, childrenByParent)
3. Với mỗi top-level node: `typeDetector.nodeType()` + `memberParser.parseChildren()`
4. Với mỗi edge cell: `typeDetector.relationType()` + resolve endpoint IDs

## Compile-time Guarantees
Mỗi concrete struct có `var _ Interface = (*struct)(nil)` để Go compiler bắt lỗi ngay nếu interface không được implement đầy đủ.
