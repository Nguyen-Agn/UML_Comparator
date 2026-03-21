# Evolution Log — Data Flow Bug History

---

## [2026-03-20] BUG #1: Builder trả về 8 Nodes thay vì 2 Classes — FIXED ✔
- **File phát hiện:** `assignment1.drawio` (Employee + ProductionWorker)
- **Tín hiệu:** Demo trả 8 Nodes thay vì 2.
- **Nguyên nhân:** `isNode()` check `vertex="1"` nhưng không lọc theo `Parent`. Child cells (Attributes, Methods sections) cũng có `vertex="1"` với `parent=CLASS_CELL_ID`.
- **Fix:**
  - `isTopLevelNode()`: Chỉ accept `parent == rootLayerID`
  - `groupChildrenByParent()`: Gom child cells vào class cha
  - `resolveToClassID()`: Edge endpoint đi ngược parent chain về class container

---

## [2026-03-20] BUG #2: Interface nhận diện sai Type = "Class" — FIXED ✔
- **File phát hiện:** `problem1.drawio` (Shape interface)
- **Tín hiệu:** Node `<<interface>>\n\nShape` hiện Type = `Class`, tên chứa newlines.
- **Nguyên nhân:** `detectNodeType()` chỉ xét `style`, không xét text value. `sanitizeHTML()` không collapse consecutive newlines.
- **Fix:**
  - `sanitizeHTML()` thêm `regexp.MustCompile("\n{2,}")` để collapse newlines.
  - `detectNodeType(style, valueHint...)` check thêm `<<interface>>` trong text value.
  - `extractCleanName()` bỏ qua dòng pattern `<<...>>` → trả về tên thực.

---

## [2026-03-20] BUG #3: Multi-line method signature bị split thành fake Attributes — FIXED ✔
- **File phát hiện:** `question5.drawio` (PoliceOffice.issueTicket)
- **Tín hiệu:** `• car: ParkedCar, meter: ParkingMeter` và `• : void` xuất hiện là Attributes.
- **Nguyên nhân:** Draw.io tự wrap method signature dài thành nhiều dòng trong cell. Dòng không có `(` → nhầm là Attribute.
- **Fix:**
  - Đếm `openCount("(")` vs `closeCount(")")` per line.
  - `pendingMethod` buffer: join các dòng tiếp theo khi chưa close đủ `)`.
  - Filter: Bỏ lines bắt đầu bằng `:` (return-type fragment), bỏ `)`, `,` standalone.

---

## [2026-03-20] KNOWN CASE: File rỗng (0 bytes) — HANDLED ✔
- **File:** `problem8.drawio` (0 bytes)
- **Hành vi hiện tại:** Parser trả `unsupported format` error → Grader sẽ báo 0 điểm.
- **Không cần Fix thêm.** Cần bổ sung unit test cho case này trong tương lai.

## [2026-03-20] BUG #4: `<<abstract>> ClassName` inline — tên không được làm sạch — FIXED ✔
- **File phát hiện:** `assignment9.drawio` (BankAccount abstract class)
- **Tín hiệu:** Node hiển thị Name = `<<abstract>> BankAccount` thay vì `BankAccount`, Type = `Class` thay vì `Abstract`.
- **Nguyên nhân:** `extractCleanName()` chỉ skip dòng **hoàn toàn** là `<<...>>`. Khi `<<abstract>>` inline cùng dòng với tên class, pattern `HasPrefix("<<") && HasSuffix(">>")` không khớp.
- **Fix:**
  - `extractCleanName()` dùng `regexp.MustCompile("<<[^>]+>>")` để strip tất cả token `<<...>>` khỏi **bất kỳ dòng nào** (inline hoặc standalone).
  - `detectNodeType(style, valueHint)` thêm nhánh: `strings.Contains(v, "<<abstract>>") → return "Abstract"`.
  - `domain.ValidNodeTypes` bổ sung `"Abstract": true`.

---

## [2026-03-20] BUG #5: Lone visibility marker `-` / `+` thành fake Attribute — FIXED ✔
- **File phát hiện:** `assignment9.drawio` (BankAccount: `• -`, `• +` xuất hiện là Attributes)
- **Tín hiệu:** Attributes list chứa `"-"` và `"+"` riêng lẻ.
- **Nguyên nhân:** Draw.io thêm dòng separator chỉ chứa visibility marker (`-`, `+`) trước một nhóm member. Builder nhận chúng là attribute hợp lệ vì không chứa `(`.
- **Fix:** Bổ sung filter: `if trimmed == "-" || trimmed == "+" || trimmed == "#" || trimmed == "~" { continue }`.

## [2026-03-20] BUG #6: Enum type nhận diện sai — `portConstraint=eastwest` không phải discriminator hợp lệ — FIXED ✔
- **File phát hiện:** `HasEnum.drawio` → Enum đúng. `assignment9.drawio` → `SavingsAccount` nhận sai là `Enum`.
- **Tín hiệu:** Class bình thường có method/attribute bị phân loại thành `Type: Enum`.
- **Nguyên nhân gốc rễ (quan trọng):** `portConstraint=eastwest` có mặt trong style của **TẤT CẢ** child cells trong Draw.io stackLayout swimlane — cả enum constant lẫn class member. Đây KHÔNG PHẢI discriminator hợp lệ.
- **Fix — Dùng content-based discriminator:**
  Enum constant thực sự = dòng plain text, KHÔNG có:
  - Visibility marker đầu dòng (`+`, `-`, `#`, `~`)
  - Dấu ngoặc `(` hoặc `)` (method)
  - Dấu `:` (typed attribute)
  - Hàm `isEnumByChildPattern()` kiểm tra từng dòng sanitized của từng child cell.
  - Separator cells (`style` chứa `"line;"`) bị bỏ qua.
- **Kết quả:**
  - `HasEnum`: Circle/Square/Triangle không có +/-/:/() → `Type: Enum` ✔
  - `SavingsAccount`: `- status: boolean` có `-` và `:` → `Type: Class` ✔

---

### QUY TẮC KHAI BÁO LOG (Dành cho Agent)
1. Phải ghi **File phát hiện** — tên file thực tế kích hoạt bug.
2. Phải ghi **Tín hiệu** — output sai trông như thế nào.
3. Phải ghi **Fix** — hàm / logic cụ thể đã sửa.

---

## [2026-03-20] BUG #7: `&#10;` numeric entity không được decode → tên class thành `>` — FIXED ✔
- **File phát hiện:** `InCorrectUML.drawio` (class `IShow` với value `&lt;&lt; Interface &gt;&gt;&#10;IShow`)
- **Tín hiệu:** Node hiển thị Name = `>` thay vì `IShow`.
- **Nguyên nhân:** `sanitizeHTML()` decode `&lt;` → `<`, `&gt;` → `>` nhưng KHÔNG decode `&#10;` (numeric HTML entity cho ký tự newline). Kết quả: `<< Interface >>&#10;IShow` là một dòng duy nhất. Regex `<<[^>]+>>` strip `<< Interface >>` để lại `&#10;IShow`. `extractCleanName` trả về `&#10;IShow`, terminal hiển thị thành `>`.
- **Fix:** Decode numeric entities **TRƯỚC** khi strip HTML tags:
  ```go
  raw = strings.ReplaceAll(raw, "&#10;", "\n")  // newline
  raw = strings.ReplaceAll(raw, "&#13;", "\r")  // carriage return
  raw = strings.ReplaceAll(raw, "&#xA;", "\n")  // hex newline
  ```
  Cũng thêm double-encoding: `&amp;lt;` → `&lt;` trước bước strip.

---

## [2026-03-20] DESIGN: Validator 2 mức độ ERROR / WARN cho Non-Standard UML ✔ MỚI
- **Lý do:** UML không chuẩn (`InCorrectUML.drawio`) có thể parse được nhưng chứa attribute không đầy đủ (`+ R`, `- GO` không có type). Pipeline không nên dừng nhưng Grader cần biết để trừ điểm.
- **Thiết kế:**
  - `IntegrityError.Severity = "ERROR"` → `FilterErrors()` → dừng pipeline
  - `IntegrityError.Severity = "WARN"` → `FilterWarns()` → tiếp tục, ghi log
- **WARN codes mới:** `SUSPECT_NODE_NAME`, `TRIVIAL_NODE_NAME`, `INCOMPLETE_ATTRIBUTE`, `INCOMPLETE_METHOD`
