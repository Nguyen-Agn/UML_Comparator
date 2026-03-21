# Kiến thức Toàn vẹn Dữ liệu (Data Integrity Knowledge)
> Cập nhật: 2026-03-20 — dựa trên kiểm thử `assignment1`, `problem1`, `question5`, `assignment9`, `HasEnum`, `InCorrectUML`.

---

## 1. Swimlane Container Pattern ✔ ĐÃ FIX

Draw.io UML Class dùng **Swimlane Container**: Container class cell chứa nhiều thẻ con. Mỗi nhóm (Attributes, Methods) là một cell con với `parent=CLASS_CELL_ID`.

**Quy tắc:**
- Top-level class: `vertex="1"` **và** `parent == rootLayerID`
- Child cells: `parent == CLASS_CELL_ID` → gom qua `groupChildrenByParent()`
- Edge endpoints đôi khi trỏ tới child cell → dùng `resolveToClassID()` đi ngược parent chain

---

## 2. Interface / Abstract / Enum Stereotype Detection ✔ ĐÃ FIX

| Stereotype | Cách detect |
|---|---|
| `<<interface>>` | `detectNodeType(style, valueHint)` → check text value + style `"ellipse"` |
| `<<abstract>>` | `detectNodeType` check text value `<<abstract>>` |
| `<<enum>>` | `detectNodeType` check text value + style `"enumeration"` + `isEnumByChildPattern()` |

`extractCleanName()` dùng regex `<<[^>]+>>` để strip **tất cả** stereotype token (standalone hoặc inline).

---

## 3. Multi-line Method Signature ✔ ĐÃ FIX

Draw.io wrap method signature dài | `pendingMethod` buffer, đếm `openCount("(")` vs `closeCount(")")`.
Lines bắt đầu bằng `:` không có `(` → bỏ qua (return-type fragment).

---

## 4. HTML Entity Decode — ĐẦY ĐỦ ✔ ĐÃ FIX

`sanitizeHTML()` phải decode theo thứ tự sau:
1. **Double-encoded entities FIRST:** `&amp;lt;` → `&lt;` → `<`
2. **Numeric newline entities:** `&#10;` → `\n`, `&#13;` → `\r`, `&#xA;` → `\n`
3. **Strip HTML tags** (`<[^>]+>` → `\n`)
4. **Named entities:** `&lt;` `&gt;` `&amp;` `&nbsp;` `&quot;`
5. **Collapse** `\n{2,}` → `\n`

**Lỗi nếu thiếu `&#10;` decode:** Value `&lt;&lt; Interface &gt;&gt;&#10;IShow` trở thành `<< Interface >>&#10;IShow` (một dòng), regex strip `<< Interface >>` để lại `&#10;IShow` → tên hiển thị thành `>`.

---

## 5. Enum Detection — Content-Based ✔ ĐÃ FIX

`portConstraint=eastwest` **KHÔNG PHẢI** discriminator — tất cả swimlane child cells dùng nó.
Dùng `isEnumByChildPattern()`: Enum constant = KHÔNG có `+/-/#/~`, KHÔNG có `(`, KHÔNG có `:`.

---

## 6. File Rỗng / Không hợp lệ

| Trường hợp | Hành động |
|---|---|
| File 0 bytes | Parser error → Grader: 0 điểm |
| Không có `<mxGraphModel>` | Parser error `unsupported format` |
| Graph 0 Node | `EMPTY_GRAPH` ERROR → dừng pipeline |

---

## 7. Non-Standard / Incorrect UML — WARN Level

**Nguồn:** `InCorrectUML.drawio` — UML không chuẩn nhưng file hợp lệ.

Validator dùng 2 mức độ:
- **ERROR** → dừng pipeline, data không dùng được
- **WARN** → tiếp tục nhưng log lại, Grader có thể trừ điểm

| WARN Code | Nguyên nhân | Ví dụ |
|---|---|---|
| `SUSPECT_NODE_NAME` | Tên chứa `>`, `<`, `&`, `&#` → HTML decode sót | Node `>` |
| `TRIVIAL_NODE_NAME` | Tên ≤ 2 ký tự → placeholder hoặc typo | Node `A` |
| `INCOMPLETE_ATTRIBUTE` | Attribute không có `:` (thiếu kiểu) | `+ R`, `- GO` |
| `INCOMPLETE_METHOD` | Method signature kết thúc bằng `(` | `issueTicket(` |

API: `domain.FilterErrors(errs)` → dừng pipeline. `domain.FilterWarns(errs)` → log/penalty.


---

## 1. Vấn đề Swimlane Container Pattern của Draw.io ✔ ĐÃ FIX

Draw.io UML Class dùng **Swimlane Container**: Container class cell chứa nhiều thẻ con. Mỗi nhóm (Attributes, Methods) là một cell con với `parent=CLASS_CELL_ID`.

**Cấu trúc thực tế:**
```xml
<mxCell id="2"  parent="1"  value="Employee"         vertex="1"/> <!-- TOP-LEVEL class -->
<mxCell id="3"  parent="2"  value="- name : String"  vertex="1"/> <!-- child: Attributes -->
<mxCell id="4"  parent="2"  value=""                 vertex="1"/> <!-- child: separator  -->
<mxCell id="5"  parent="2"  value="+ Employee(...)+" vertex="1"/> <!-- child: Methods    -->
<mxCell id="7"  edge="1"    source="X"  target="2"/>              <!-- edge → container  -->
```

**Quy tắc:**
- Top-level class: `vertex="1"` **và** `parent == rootLayerID`
- Child cells: `parent == CLASS_CELL_ID` → gom vào class cha qua `groupChildrenByParent()`
- Edge endpoints đôi khi trỏ tới child cell → dùng `resolveToClassID()` để đi ngược parent chain

---

## 2. Interface Stereotype Detection ✔ ĐÃ FIX

Draw.io có hai cách đánh dấu Interface:
- **Bằng style:** `style` chứa `"ellipse"` hoặc keyword tương đương
- **Bằng text:** Value cell chứa `"<<interface>>"` dạng stereotype trên dòng đầu

**Xử lý:**
- `extractCleanName()` bỏ qua các dòng có pattern `<<...>>` để lấy tên thực
- `detectNodeType(style, valueHint)` kiểm tra cả style **và** text value

---

## 3. Multi-line Method Signature ✔ ĐÃ FIX

Draw.io đôi khi tự xuống dòng khi method signature quá dài trong một cell:

**Ví dụ thực tế (question5.drawio):**
```
cell value = "+ issueTicket(\n"
             "car: ParkedCar, meter: ParkingMeter\n"
             "): ParkingTicket"
```

**Lỗi gây ra nếu không handle:**
- `"car: ParkedCar, meter: ParkingMeter"` → nhận nhầm là **Attribute**
- `"): ParkingTicket"` → nhận nhầm là **Attribute** (bắt đầu bằng `:`)

**Fix:** Đếm `openCount("(")` và `closeCount(")")` — nếu chưa bằng nhau, buffer các dòng tiếp theo. Lines bắt đầu bằng `:` mà không có `(` → bỏ qua.

---

## 4. File Rỗng / Không hợp lệ

| Trường hợp | Hành động đúng |
|---|---|
| File 0 bytes (`problem8.drawio`) | Parser trả error → Grader cho 0 điểm ngay, không crash |
| File không có `<mxGraphModel>` | Parser trả error `unsupported format` |
| Graph build ra 0 Node | `ValidateGraph()` trả `EMPTY_GRAPH` → dừng pipeline |

---

## 5. Nguyên tắc kiểm tra trước khi so sánh

Mọi `Edge.SourceID` và `Edge.TargetID` **phải tồn tại** trong `graph.Nodes` (không được trỏ tới child cell ID của swimlane con). Validator: `domain.ValidateGraph(*UMLGraph, label)`.
