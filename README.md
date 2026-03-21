# UML Comparator — Hệ thống Chấm điểm & Đối sánh Biểu đồ UML

`UML Comparator` là công cụ mạnh mẽ được phát triển bằng **Golang**, thiết kế để tự động hóa việc chấm điểm và so khớp các biểu đồ UML được vẽ từ [draw.io](https://diagrams.net) (dưới dạng tệp `.drawio` hoặc XML) (có thể mở rộng về sau)

Dự án cung cấp khả năng so sánh chi tiết giữa bản vẽ **Đáp án (Solution)** và bài làm của **Sinh viên (Student)**, giúp phát hiện nhanh các sai sót về cấu trúc, thành phần và mối quan hệ.

---

## 🎨 Tính năng chính

- **So sánh trực quan (Visual Side-by-Side):** Hiển thị bảng so sánh hai bên với mã màu ANSI (Xanh/Vàng/Đỏ) để nhận diện nhanh sự khác biệt.
- **Thuật toán Matching thông minh:** Sử dụng *Levenshtein Distance* để nhận diện các Class/Actor ngay cả khi sinh viên gõ sai tên (Typo).
- **Phân tích chi tiết thành phần:** So sánh tỉ mỉ từng Thuộc tính (Attributes) và Phương thức (Methods), bao gồm cả Constructor và Getter/Setter.
- **Kiểm tra mối quan hệ (Edges):** Nhận diện các liên kết (Inheritance, Association, Aggregation...) và phát hiện các lỗi như mũi tên ngược chiều.
- **Báo cáo chuyên sâu (Grading Report):** Xuất báo cáo chi tiết về các lỗi cụ thể để phục vụ việc chấm điểm hoặc phản hồi cho sinh viên.
- **Tính di động cực cao:** Build ra một file thực thi duy nhất (`.exe`), chạy ngay mà không cần cài đặt môi trường Go hay Java.

---

## 🏗️ Kiến trúc Hệ thống

Dự án tuân thủ nghiêm ngặt nguyên tắc **SOLID**, chia thành các module độc lập:

| Module | Chức năng |
| :--- | :--- |
| **Parser** | Đọc và giải mã XML từ tệp `.drawio` (Base64/Inflate). |
| **Builder** | Chuyển đổi dữ liệu XML thô thành đồ thị `UMLGraph` (String-Based). |
| **Pre-Matcher** | Tiền xử lý dữ liệu, phân loại và chuẩn hóa các thực thể (Struct-Based). |
| **Matcher** | So khớp các node giữa hai bản vẽ dựa trên độ tương đồng văn bản (Mapping). |
| **Comparator(💱 hiện tại)** | So sánh chi tiết thuộc tính, phương thức và các liên kết mũi tên. |
| **Grader** | Áp dụng luật chấm điểm (JSON) để tính toán điểm số cuối cùng. |
| **Visualizer** | Xuất báo cáo trực quan dưới dạng CLI Dashboard hoặc hình ảnh màu. |

---

# 🚀 Hướng dẫn sử dụng nhanh

### 1. Sử dụng công cụ CompareUML_CLI (Dành cho người dùng cuối)
Nếu bạn sử dụng Windows, hãy sử dụng tệp đóng gói sẵn:

- **Chế độ tương tác:** 
  + Tải về 2 file .bat và .exe vào cùng 1 thư mục.
  + Click đúp vào file .bat.
  + Chương trình sẽ yêu cầu bạn nhập tên/đường dẫn file Solution và Student để so sánh (file .drawio).
- **Chế độ dòng lệnh (CLI):**
  + Tải về file .exe.
  + Mở CMD hoặc PowerShell tại thư mục chứa file .exe.
  ```bash
  ./CompareUML_CLI.exe <solution.drawio> <student.drawio>
  ```

### 2. Dành cho lập trình viên (Go)
Để chạy từ mã nguồn:
```bash
go run ./cmd/compare/main.go <solution.drawio> <student.drawio>
```

---

## 📂 Cấu trúc thư mục

- `cmd/`: Chứa các tệp `main.go` cho từng công cụ (compare, matcher, parser...).
- `domain/`: Định nghĩa các cấu trúc dữ liệu cốt lõi (`UMLGraph`, `DiffReport`).
- `scheme/`: Tài liệu đặc tả dữ liệu giữa các module (Single Source of Truth).
- `skills/`: Hệ thống **Skill OS** dành cho AI Agent hỗ trợ phát triển dự án.

---

## 📜 Tài liệu tham khảo
- [Quy trình xử lý (Flow)](flow.md)
- [Chiến lược vibe coding (Strategy)](Stratergy.md)
- [Sơ đồ kiến trúc (Mermaid)](architecture.mmd)
- [Hướng dẫn sử dụng chi tiết](USAGE_GUIDE.md)

---
*Dự án được xây dựng với mục tiêu nâng cao tính minh bạch và hiệu quả trong việc giảng dạy & chấm bài môn Thiết kế Hệ thống OOP.*
