# Stage 2: Identity Filter - Implementation Plan

Để hiện thực thành công Bộ lọc Danh tính (Identity Filter) cho phép chặn các trường hợp trái nghĩa hoặc không liên quan (nhưng vẫn giữ được tính linh hoạt), chúng ta sẽ cần làm các nhiệm vụ sau đây (chủ yếu tác động vào package `matcher/`):

## 1. Xây dựng tiện ích Tokenizer (Tách từ)
- **File dự kiến:** `matcher/tokenizer.go`
- **Nhiệm vụ:** Viết hàm nhận vào chuỗi tên class/method và tiến hành "chặt" chuỗi đó ra thành mảng các từ đơn lẻ (đã được parse lower case).
- **Phạm vi xử lý:**
  - `PascalCase` / `camelCase` (Ví dụ: `EncryptService` ➔ `["encrypt", "service"]`)
  - `snake_case` (Ví dụ: `user_account_id` ➔ `["user", "account", "id"]`)
  - Loại bỏ các ký tự dấu con hoặc số (tuỳ vào setup).

## 2. Xây dựng "Antonym Engine" (Bộ phát hiện từ trái nghĩa)
- **File dự kiến:** `matcher/antonym_detector.go`
- **Nhiệm vụ:** Trừu tượng hóa cách tìm từ đối nghĩa. Có thể triển khai theo các quy tắc do chúng ta đặt ra:
  1. **Dictionary (Từ điển gán cứng):** Từ điển các cặp từ chuyên ngành phổ biến trái ngược nhau: `encrypt/decrypt`, `encode/decode`, `login/logout`, `open/close`, `min/max`, `import/export`, v.v...
  2. **Prefix-based (So sánh Tiền tố):** Kiểm tra xem 2 từ có bị "đối ngược" vì mang 2 tiền tố khác nhau cho cùng một từ gốc (root) hay không. Ví dụ: `en-` và `de-`.

## 3. Tạo ra cấu trúc Interface `IdentityValidator`
- **File dự kiến:** `matcher/identity_validator.go`
- **Nhiệm vụ:** Cầu nối luân chuyển dữ liệu: Lấy Tên 1 và Tên 2 ➔ Chạy Tokenizer ➔ Run Antonym check ➔ Quyết định Valid hay Invalid.
- **Cấu trúc (Dự kiến):**
  ```go
  type IIdentityValidator interface {
      // Trả về false nếu s1 và s2 bị vi phạm ngữ nghĩa (VD: là từ trái nghĩa)
      IsValid(name1, name2 string) bool
  }

  type StandardIdentityValidator struct { ... }
  ```

## 4. Tích hợp Validator vào luồng hiện tại (`StandardEntityMatcher`)
- **File bị ảnh hưởng:** `matcher/standard_entity_matcher.go`
- **Công việc:**
  - Bổ sung `IIdentityValidator` như một property trong struct `StandardEntityMatcher`.
  - **Sửa constructor**: Nhận thêm tham số validator. *(Lưu ý: Do sửa Constructor, mình sẽ phải sửa ở các file `cmd/` đang gọi constructor này)*
  - **Chèn code vào lõi matching (`runPass`)**: 
    ```go
    if !m.identityValidator.IsValid(solNode.Name, stuNode.Name) {
        continue // Chặn đứng từ "vòng gửi xe"
    }
    ```

## 5. Bổ sung Test Cases
- Tạo mới các file Test độc lập cho các tiện ích nhỏ: `tokenizer_test.go`, `identity_validator_test.go`.
- Sửa đổi file `standard_entity_matcher_test.go` hiện tại:
  - Cập nhật hàm khởi tạo Matcher cho các test case cũ.
  - Thêm một test bài bản: Đưa vào `EncryptService` và `DecryptService` với độ giống cực cao (hoặc `ArchWeight` y xì hệt), để xem nó có thực sự bị **REJECT** như mong đợi hay không.

---
**Nhận xét:** Tổ chức code và viết theo file như trên sẽ giúp module matching của bạn tuân thủ triệt để nguyên tắc **Single Responsibility Principle**. Bạn có xác nhận đi theo thiết kế này để mình bắt tay vào viết hàm `Tokenizer` luôn không?
