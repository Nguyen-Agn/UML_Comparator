# Hướng dẫn sử dụng CompareUML_CLI

`CompareUML_CLI` là công cụ dòng lệnh mạnh mẽ giúp so sánh cấu trúc giữa hai tệp mã nguồn UML dạng `.drawio`. Công cụ này tập trung vào việc hiển thị sự khác biệt về Class, Thuộc tính (Attributes) và Phương thức (Methods).

## 1. Cách chạy nhanh (Quick Start)

Nếu bạn ở trên Windows, hãy **click đúp vào file `CompareUML_CLI.bat`**. 
- Bạn sẽ được yêu cầu nhập tên file **Đáp án (Solution)**.
- Sau đó nhập tên file **Bài nộp (Student)**.
- Nhấn **Enter** để xem kết quả ngay lập tức.
- Công cụ sẽ tự động kiểm tra xem tệp có tồn tại hay không và yêu cầu nhập lại nếu sai.

## 2. Cách chạy qua Command Line

Mở CMD hoặc PowerShell tại thư mục chứa file `.exe` và sử dụng cú pháp:

```bash
CompareUML_CLI.exe <path/to/solution.drawio> <path/to/student.drawio>
```

**Ví dụ:**
```bash
CompareUML_CLI.exe UMLs_testcase/assignment2.drawio UMLs_testcase/assignment1.drawio
```

## 3. Giải thích kết quả

Màn hình sẽ hiển thị 3 phần chính:

### A. Nodes Side-by-Side (Bảng so sánh Class)
Bảng này chia làm 2 cột: **SOLUTION** (Đáp án) và **STUDENT** (Bài làm).
- `✔` (Xanh): Khớp hoàn toàn 100%.
- `≈` (Vàng): Khớp mờ (tên có thể hơi khác nhưng cấu trúc tương đồng).
- `✗` (Đỏ): Tên Class không tìm thấy hoặc bị thiếu.
- **Dấu `+` (Xanh):** Sinh viên làm thừa ra (Extra).
- **Dấu `-` (Đỏ):** Sinh viên bị thiếu (Missing) so với mẫu.

### B. Edges (Relations)
Kiểm tra các mối quan hệ (Inheritance, Association...):
- `✔`: Quan hệ chính xác.
- `✗`: Quan hệ không tìm thấy hoặc sai hướng mũi tên.

### C. Quick Stats
Tóm tắt phần trăm khớp của bài làm.

## 4. Lưu ý
- Đảm bảo các tệp `.drawio` không bị lỗi định dạng (Export từ diagrams.net).
- Màu sắc hiển thị tốt nhất trên Terminal hỗ trợ ANSI (Windows Terminal, PowerShell Core).
