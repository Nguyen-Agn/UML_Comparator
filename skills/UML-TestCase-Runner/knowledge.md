# Knowledge: UML TestCase Error Codes

> **Knowledge Layer** — tài nguyên domain knowledge để Agent chạy và đánh
> giá test case UML đúng cách.

---

## 1. Pipeline Tổng Quan

```
.drawio file
    ↓
Parser (XML parse)
    ↓
Builder.Build(RawXMLData) → *domain.UMLGraph  (hoặc error)
    ↓
domain.ValidateGraph(graph, label) → []IntegrityError
    ↓
FilterErrors() → pipeline STOP nếu có ERROR
FilterWarns()  → log WARN, pipeline CONTINUE
    ↓
Matcher / Comparator
```

---

## 2. Bảng Error Codes Đầy Đủ

### 2.1 ERROR Codes (pipeline MUST STOP)

| Code | Trigger Condition | File Mẫu |
|---|---|---|
| `EMPTY_GRAPH` | `len(graph.Nodes) == 0` sau Build | `err_empty_graph.drawio` |
| `EMPTY_NODE_NAME` | Một node có `Name == ""` | `err_empty_node_name.drawio` |
| `INVALID_NODE_TYPE` | `node.Type` không thuộc `{Class, Interface, Abstract, Actor, Enum}` | `err_invalid_node_type.drawio` |
| `DANGLING_EDGE_SOURCE` | `edge.SourceID` không tồn tại trong `graph.Nodes` | `err_dangling_edge.drawio` |
| `DANGLING_EDGE_TARGET` | `edge.TargetID` không tồn tại trong `graph.Nodes` | `err_dangling_edge.drawio` (nếu cả target cũng sai) |
| `SELF_LOOP_EDGE` | `edge.SourceID == edge.TargetID` | `err_self_loop.drawio` |

### 2.2 WARN Codes (pipeline CONTINUE nhưng flag)

| Code | Trigger Condition | File Mẫu |
|---|---|---|
| `SUSPECT_NODE_NAME` | `node.Name` chứa `>`, `<`, `&`, `&#`, `<<`, `>>` | `warn_suspect_name.drawio` |
| `TRIVIAL_NODE_NAME` | `len([]rune(node.Name)) <= 2` và `node.Name != ""` | `warn_trivial_name.drawio` |
| `INCOMPLETE_ATTRIBUTE` | Attribute không chứa `:` và không rỗng | `warn_incomplete_attr.drawio` |
| `INCOMPLETE_METHOD` | Method kết thúc bằng `(` sau TrimSpace | `warn_incomplete_method.drawio` |

---

## 3. Cách Builder Phân Loại Node Type

Builder (`standard_builder.go`) nhận dạng type dựa trên:

| Điều kiện trong drawio XML | Type được gán |
|---|---|
| `style` chứa `shape=umlClass` HOẶC `swimlane` kết hợp với `value` không có stereotype | `Class` |
| `value` chứa `<<interface>>` hoặc `<<Interface>>` (case-insensitive) | `Interface` |
| `value` chứa `<<abstract>>` hoặc `<<Abstract>>` | `Abstract` |
| `value` chứa `<<actor>>` hoặc có `shape=mxgraph.uml.actor` | `Actor` |
| `value` chứa `<<enum>>` hoặc `<<Enum>>` | `Enum` |
| Không khớp bất kỳ điều kiện nào | `""` → INVALID_NODE_TYPE |

---

## 4. Cấu Trúc File Drawio Hợp Lệ Chuẩn

```xml
<mxfile>
  <diagram>
    <mxGraphModel>
      <root>
        <mxCell id="0"/>              <!-- required root cell -->
        <mxCell id="1" parent="0"/>   <!-- required layer cell -->
        
        <!-- Class node (swimlane container) -->
        <mxCell id="2" value="ClassName" style="swimlane;..." vertex="1" parent="1">
          <mxGeometry .../>
        </mxCell>
        
        <!-- Attribute cell (child of class container) -->
        <mxCell id="3" value="- attr: Type" style="text;..." vertex="1" parent="2">
          <mxGeometry y="26" .../>
        </mxCell>
        
        <!-- Divider line cell -->
        <mxCell id="4" value="" style="line;..." vertex="1" parent="2">
          <mxGeometry y="52" .../>
        </mxCell>
        
        <!-- Method cell (child of class container) -->
        <mxCell id="5" value="+ method(): ReturnType" style="text;..." vertex="1" parent="2">
          <mxGeometry y="60" .../>
        </mxCell>
        
        <!-- Edge cell -->
        <mxCell id="6" value="" style="endArrow=block;..." edge="1" source="2" target="X" parent="1">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>
```

---

## 5. Các Anti-Pattern Thường Gặp

| Anti-Pattern | Lý do sai | Error/Warn trigger |
|---|---|---|
| Swimlane với `value=""` | Tên class rỗng | `EMPTY_NODE_NAME` |
| Dùng shape `rounded=1` thay vì `swimlane` | Builder không nhận ra là UML entity | `INVALID_NODE_TYPE` |
| Edge `source` trỏ vào ID không tồn tại | Kết nối bị đứt (ghost reference) | `DANGLING_EDGE_SOURCE` |
| Edge `source == target` | Self-loop không hợp lệ trong UML class diagram | `SELF_LOOP_EDGE` |
| Stereotype `<<interface>>` không được xóa khỏi Name | HTML decode chưa hoàn chỉnh | `SUSPECT_NODE_NAME` |
| Tên class 1-2 ký tự | Placeholder, chưa điền tên thật | `TRIVIAL_NODE_NAME` |
| Attribute `- fieldName` (không có `:`) | Thiếu type declaration | `INCOMPLETE_ATTRIBUTE` |
| Method `+ method(` (kết thúc `(`) | Signature bị cắt giữa chừng | `INCOMPLETE_METHOD` |

---

## 6. Cách Assert Trong Go Test

```go
// Helper: kiểm tra danh sách errors có chứa code mong muốn không
func assertHasCode(t *testing.T, errs []domain.IntegrityError, code string) {
    t.Helper()
    for _, e := range errs {
        if e.Code == code {
            t.Logf("✔ Got expected error code: %s [%s] — %s", e.Code, e.Severity, e.Message)
            return
        }
    }
    t.Errorf("✘ Expected error code %q not found. Got: %v", code, errs)
}

// Helper: kiểm tra không có code nào trong danh sách
func assertNoCode(t *testing.T, errs []domain.IntegrityError, code string) {
    t.Helper()
    for _, e := range errs {
        if e.Code == code {
            t.Errorf("✘ Unexpected error code %q found: %s", code, e.Message)
            return
        }
    }
    t.Logf("✔ Code %q correctly absent", code)
}
```
