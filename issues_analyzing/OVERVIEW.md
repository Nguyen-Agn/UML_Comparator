# 🌟 TỔNG QUAN DỰ ÁN (OVERVIEW)

Tài liệu này tóm tắt gọn gàng các thay đổi mới nhất trong dự án. Bản cập nhật này chủ yếu tập trung vào việc **Cải thiện thuật toán Matching (Pipeline 3 Giai Đoạn)** nhằm giải quyết vấn đề nhận diện sai (false positive) khi các lớp có tên đối lập nhau (VD: `Encrypt` và `Decrypt`) nhưng lại có độ tương đồng chuỗi cao.

---

## 🗂️ 1. Tài Liệu Phân Tích & Giải Pháp (Documentation)
Các tài liệu này giải thích chi tiết lỗi gặp phải và tư duy đằng sau thiết kế xử lý:
- `issues_analyzing/matching_strategy.md`: Bài toán và những lỗ hổng của cơ chế matching cũ.
- `issues_analyzing/solution_ver2.md`: Đề xuất kiến trúc giải pháp 3 giai đoạn (Fuzzy -> Filter -> Rank).
- `issues_analyzing/stage2_matching_issue.md`: Giải thích đào sâu cơ chế hoạt động của giai đoạn 2 (Identity Filter) và xử lý từ trái nghĩa.

## 🛠️ 2. Các Module Mới Cốt Lõi (Core Components)
Các tệp mã nguồn dưới đây được tạo mới hoàn toàn bên trong thư mục `matcher/` để thực hiện nhiệm vụ lọc ngữ nghĩa:
- **Tách từ khóa (`tokenizer.go` & `tokenizer_test.go`):**  
  Module chịu trách nhiệm tách chi tiết chuỗi định danh (từ PascalCase, camelCase) thành danh sách các từ cơ bản nguyên thủy để phân tích.
- **Phát hiện từ trái nghĩa (`antonym_detector.go` & `antonym_detector_test.go`):**  
  Cung cấp logic chặn các từ mang nghĩa phủ định lẫn nhau, kết hợp nhận diện từ vựng cứng và tiền tố cấu tạo ngược (VD: `en-`/`de-`).
- **Trình xác thực Danh Tính (`identity_validator.go` & `identity_validator_test.go`):**  
  Đóng vai trò là "người gác cổng" ở **Giai đoạn 2**, phối hợp Tokenizer và AntonymDetector thành một bộ rào cản vững chắc.

## 🔄 3. Nhúng Tích Hợp Vào Hệ Thống Hiện Tại (Integration Changes)
Các file hệ thống chính đã được chỉnh sửa để đón nhận tính năng trên:
- **Thuật toán cốt lõi (`matcher/standard_entity_matcher.go`):** Gọi vòng kiểm tra `IdentityValidator.IsValid(...)` trước khi cấp phép điểm candidate để triệt tiêu false positive từ trong trứng nước.
- **Kiểm thử tự động (`matcher/standard_entity_matcher_test.go`):** Bổ sung loạt test cho thấy Validator chủ động triệt tiêu các ứng viên rác mà không ảnh hưởng tới kết quả của quá trình đánh vần sai (Typo).
- **Luồng khởi tạo hệ thống (Dependency Injection):** Chỉnh sửa toàn bộ các Entrypoints (điểm chạy chương trình) để tiêm cài đặt Validator vào bên trong Matcher:
  - Giao diện dòng lệnh (CLI): `cmd/compare/main.go`, `cmd/grade_batch/main.go`, `cmd/match/main.go`, `cmd/visualize/main.go`
  - Giao diện người dùng (GUI): `gui/service/uml_processor.go`
- **Cấu hình (`go.mod`):** Cập nhật danh sách phụ thuộc.
- **📦 Biên dịch phần mềm (Rebuild Binaries):** Phải build (compile) lại toàn bộ các ứng dụng chạy nhị phân (`.exe`) bên trong thư mục `portable/` để thuật toán xử lý từ trái nghĩa chính thức được đưa vào ứng dụng đầu cuối. Các tệp đã được cập nhật:
  - `portable/lecture_cli_parallel.exe`
  - `portable/student_uml.exe`
  - `portable/student_uml_cli.exe`
  - `portable/teacher_cipher.exe`

## ✅ 4. Theo Dõi & Kế Hoạch (Issue Tracking)
- **`issue.md`**: Đánh dấu cập nhật trạng thái [OK] với yêu cầu _cải thiện phát hiện và giảm độ nhạy sai phạm cho tên lớp_.