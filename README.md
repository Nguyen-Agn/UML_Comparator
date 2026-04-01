# UML Comparator — Hệ thống Chấm điểm & Đối sánh Biểu đồ UML

`UML Comparator` là công cụ mạnh mẽ được phát triển bằng **Golang**, thiết kế để tự động hóa việc chấm điểm và so khớp các biểu đồ UML được vẽ từ [draw.io](https://diagrams.net) (dưới dạng tệp `.drawio` hoặc XML) (có thể mở rộng về sau)

Dự án cung cấp khả năng so sánh chi tiết giữa bản vẽ **Đáp án (Solution)** và bài làm của **Sinh viên (Student)**, giúp phát hiện nhanh các sai sót về cấu trúc, thành phần và mối quan hệ.

---

## 🎨 Tính năng chính

- **So sánh trực quan (Visual Side-by-Side):** Hiển thị bảng so sánh hai bên với mã màu ANSI (Xanh/Vàng/Đỏ) để nhận diện nhanh sự khác biệt.
- **Thuật toán Matching thông minh:** Sử dụng *Levenshtein Distance* để nhận diện các Class/Actor ngay cả khi sinh viên gõ sai tên (Typo).
- **Phân tích chi tiết thành phần:** So sánh tỉ mỉ từng Thuộc tính (Attributes) và Phương thức (Methods), bao gồm cả Constructor và Getter/Setter.
- **Kiểm tra mối quan hệ (Edges):** Nhận diện các liên kết (Inheritance, Association, Aggregation...) và phát hiện các lỗi như mũi tên ngược chiều.
- **Chấm điểm tự động (Grading):** Tính điểm dựa trên số lượng thành phần đúng/sai/thiếu, xuất điểm + phần trăm.
- **Báo cáo HTML trực quan (Visualizer):** Xuất file `.html` self-contained với giao diện dark theme, hiển thị side-by-side Student vs Solution, color-coded từng member, và bảng tổng kết.
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
| **Comparator** | So sánh chi tiết thuộc tính, phương thức và các liên kết mũi tên. |
| **Grader** | Tính điểm dựa trên DiffReport: attribute/method/edge/node scoring. |
| **Visualizer** | Xuất báo cáo HTML self-contained với dark theme và color-coded status. |

---
![USAGE](./USAGE_GUIDE.md)
---

## 📂 Cấu trúc thư mục

- `cmd/`: Chứa các tệp `main.go` cho từng công cụ (`compare`, `visualize`, `match`, `prematch`, `build`, `parse`).
- `domain/`: Định nghĩa các cấu trúc dữ liệu cốt lõi (`UMLGraph`, `DiffReport`, `GradeResult`).
- `grader/`: Module chấm điểm dựa trên DiffReport.
- `visualizer/`: Module xuất báo cáo HTML self-contained.
- `scheme/`: Tài liệu đặc tả dữ liệu giữa các module (Single Source of Truth).
- `skills/`: Hệ thống **Skill OS** dành cho AI Agent hỗ trợ phát triển dự án.

---

## 📜 Tài liệu tham khảo
- [Quy trình xử lý (Flow)](flow.md)
- [Chiến lược vibe coding (Strategy)](Stratergy.md)
- [Sơ đồ kiến trúc (Mermaid)](architecture.mmd)
- [Hướng dẫn sử dụng chi tiết](USAGE_GUIDE.md)

---
# Review 
**Note**: Các file `cmd/*/main.go` là entry point gọi các module và in output. `cmd/visualize` là công cụ chính xuất báo cáo HTML. Các cmd khác (`compare`, `match`, `prematch`...) chủ yếu dùng để debug và verify từng module riêng lẻ.

## Ảnh review CLI

<details>
<summary>Ảnh review CLI</summary>

![assignment 1](./other/assignment1.png)

![assignment 2](./other/assignment2.png)

![compare](./other/compare.png)

</details>

---
*Dự án được xây dựng với mục tiêu nâng cao tính minh bạch và hiệu quả trong việc giảng dạy & chấm bài môn Thiết kế Hệ thống OOP.*
