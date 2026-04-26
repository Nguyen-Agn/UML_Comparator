# Schema: RawXMLData

**Package:** `domain`  
**Kind:** Primitive Type Alias  
**Pipeline Stage:** Output của `IFileParser`, Input của `IModelBuilder`

---

## Định nghĩa Go
```go
type RawXMLData string
```

## Mô tả
`RawXMLData` là một chuỗi string chứa toàn bộ nội dung XML thuần (plain text) của một biểu đồ Draw.io sau khi đã được giải mã/giải nén từ file `.drawio` gốc.

## Bất Biến (Invariants)
- **Luôn bắt đầu bằng thẻ `<mxGraphModel`** hoặc `<root>`. Không được phép chứa thẻ bao bọc `<mxfile>` hay `<diagram>`.
- **Không được rỗng** (`""`). Nếu file rỗng hoặc không hợp lệ, `IFileParser.Parse()` trả error thay vì giá trị rỗng.
- **Không bị nén** (không chứa chuỗi Base64 mã hóa). Tầng giải mã đã xử lý hết.

## Ví dụ giá trị hợp lệ
```xml
<mxGraphModel dx="1290" dy="687" grid="1" gridSize="10">
  <root>
    <mxCell id="0" />
    <mxCell id="1" parent="0" />
    <mxCell id="2" value="Animal" style="shape=umlClass;..." vertex="1" parent="1">
      <mxGeometry x="100" y="100" width="180" height="90" as="geometry" />
    </mxCell>
  </root>
</mxGraphModel>
```

## Ví dụ giá trị KHÔNG hợp lệ
```xml
<!-- KHÔNG hợp lệ: còn wrapper mxfile -->
<mxfile host="app.diagrams.net">
  <diagram id="abc">...</diagram>
</mxfile>

<!-- KHÔNG hợp lệ: còn dạng nén Base64 -->
jZLBboMwDIafJfctAtRxLSmwXbb...
```
