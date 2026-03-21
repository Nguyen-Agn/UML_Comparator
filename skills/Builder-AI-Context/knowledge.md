# Knowledge: Builder Module — Kiến Trúc & Interface Pattern

## 1. Phân loại Builder (Future-Proofing)

Hệ thống hỗ trợ nhiều loại Builder ứng với các nguồn dữ liệu khác nhau thông qua `builder.GetBuilder(sourceType string)`:

| Loại | Builder Impl | Định dạng nguồn | Mục đích |
|---|---|---|---|
| **Diagram-based** | `DrawioModelBuilder` | XML (Draw.io) | Parse từ file thiết kế đồ họa Draw.io. |
| **Data-based** | `JsonModelBuilder` (Tương lai) | JSON / YAML | Nhập dữ liệu từ các tool chuyên dụng hoặc API. |
| **Code-based** | `JavaModelBuilder` (Tương lai) | `.java` files | Tự động tạo UMLGraph từ source code (Reverse Engineering). |

**Lưu ý:** Tất cả các builder đều thực thi interface `IModelBuilder` và nhận vào `domain.RawModelData`.

## 2. Phân loại thẻ mxCell (Dành cho Drawio)
- **Node:** `vertex="1"`. Style chứa `shape=umlClass` hoặc `swimlane` (container class). Không có `edge="1"`.
- **Edge:** `edge="1"` + bắt buộc `source="id1"` và `target="id2"`. Style định kiểu mũi tên.
  - `endArrow=block` → Inheritance
  - `endArrow=block` + `dashed=1` → Realization  
  - `endArrow=open/none` → Association
  - `endArrow=diamond` → Aggregation / Composition

## 3. Rủi ro nội dung văn bản (HTML Entities)
Draw.io mã hóa text thành HTML. Ví dụ: `<b>+ id : int</b><br/>+ getName()`.
> **Nhiệm vụ sống còn:** Hàm `htmlSanitizer.clean()` phải decode entities → strip HTML tags → strip stereotypes → tách newlines trước khi phân loại Attr/Method.

## 4. Cấu Trúc DrawioModelBuilder (Interface-Based)

`DrawioModelBuilder` phụ thuộc hoàn toàn vào các internal interfaces để đảm bảo DIP:

| Interface | Vai trò |
|---|---|
| `IXMLParser` | Xử lý cấu trúc đồ thị XML của Draw.io. |
| `ITextSanitizer` | Làm sạch và chuẩn hóa nội dung văn bản. |
| `ITypeDetector` | Nhận diện kiểu Node/Edge dựa trên style và stereotype. |
| `IMemberParser` | Trích xuất Attributes và Methods từ các ô con. |

## 5. Lợi ích của Kiến trúc mới
- **Extensibility**: Có thể thêm builder mới cho nguồn dữ liệu mới (JSON, Code) chỉ bằng cách tạo struct mới thực thi `IModelBuilder`.
- **Maintainability**: Logic của Draw.io được tách biệt hoàn toàn trong `DrawioModelBuilder`.
- **Testability**: Các sub-components có interface riêng giúp dễ dàng mock dữ liệu khi viết unit test.
