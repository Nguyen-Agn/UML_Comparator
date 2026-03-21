# Schema: UMLGraph, UMLNode, UMLEdge

**Package:** `domain`  
**Pipeline Stage:** Output của `IModelBuilder`, Input của `IEntityMatcher` và `IComparator`

---

## Định nghĩa Go
```go
type UMLGraph struct {
    ID    string
    Nodes []UMLNode
    Edges []UMLEdge
}

type UMLNode struct {
    ID         string
    Name       string
    Type       string   // "Class" | "Interface" | "Actor" | "Enum"
    Attributes []string // VD: ["- id : int", "+ name : String"]
    Methods    []string // VD: ["+ getName() : String", "+ setName(name: String) : void"]
}

type UMLEdge struct {
    SourceID     string
    TargetID     string
    RelationType string // Xem bảng Relation Types bên dưới
    SourceLabel  string // Nhãn multiplicity phía nguồn, VD: "1"
    TargetLabel  string // Nhãn multiplicity phía đích, VD: "0..*"
}
```

---

## Bảng RelationType chuẩn

| Giá trị `RelationType` | Ký hiệu UML | Mô tả |
|---|---|---|
| `"Inheritance"` | Mũi tên tam giác rỗng, nét liền | Class con kế thừa Class cha |
| `"Realization"` | Mũi tên tam giác rỗng, nét đứt | Class implement Interface |
| `"Association"` | Mũi tên thường, nét liền | Liên kết hai chiều hoặc một chiều |
| `"Aggregation"` | Hình thoi rỗng | Quan hệ "có" (whole - part) |
| `"Composition"` | Hình thoi đặc | Quan hệ "có" mạnh (part phụ thuộc whole) |
| `"Dependency"` | Mũi tên, nét đứt | Phụ thuộc tạm thời |

---

## Ví dụ JSON tương đương (minh họa cấu trúc)

```json
{
  "ID": "solution-graph-001",
  "Nodes": [
    {
      "ID": "node-animal",
      "Name": "Animal",
      "Type": "Class",
      "Attributes": ["- name : String"],
      "Methods": ["+ getName() : String"]
    },
    {
      "ID": "node-dog",
      "Name": "Dog",
      "Type": "Class",
      "Attributes": [],
      "Methods": ["+ bark() : void"]
    }
  ],
  "Edges": [
    {
      "SourceID": "node-dog",
      "TargetID": "node-animal",
      "RelationType": "Inheritance",
      "SourceLabel": "",
      "TargetLabel": ""
    }
  ]
}
```

## Bất Biến (Invariants)
- `UMLNode.Name` phải được **chuẩn hóa** (trimmed, không có ký tự HTML) trước khi lưu vào struct.
- `Attributes` và `Methods` là **unordered** — khi compare không được dựa vào index.
- `UMLEdge.SourceID` và `TargetID` tham chiếu tới `UMLNode.ID` **trong cùng một UMLGraph**.
