# Hướng dẫn sử dụng UML Comparator

`UML Comparator` là công cụ so sánh và chấm điểm biểu đồ UML từ file `.drawio`. Công cụ phân tích cấu trúc Class, Attributes, Methods, và Relations, sau đó xuất báo cáo HTML trực quan.

---

## 1. Quick Start — VisualizeUML (Khuyến nghị)

### Chế độ tương tác (Windows)
1. Đặt `VisualizeUML.bat` và `visualize.exe` vào cùng thư mục.
2. **Click đúp** vào `VisualizeUML.bat`.
3. Nhập đường dẫn file **Solution** (đáp án) → công cụ tự verify file tồn tại.
4. Nhập đường dẫn file **Student** (bài nộp) → tự verify.
5. Nhập tên file output `.html` hoặc nhấn **Enter** để tự đặt tên.
6. Xác nhận → file HTML được tạo và **tự mở trong trình duyệt**.

### Chế độ dòng lệnh (CLI)
- Chỉ cần file visualize.exe
```bash
visualize.exe <solution.drawio> <student.drawio> [output.html]
```

**Ví dụ:**
```bash
visualize.exe UMLs_testcase/problem1.drawio UMLs_testcase/assignment1.drawio
```
→ Xuất file `report_assignment1.html` và tự mở browser.

> **Lưu ý:** Nếu chạy từ PowerShell, dùng `cmd /c .\VisualizeUML.bat` để đảm bảo interactive input hoạt động.

---

## 2. Chạy từ mã nguồn (Go)

```bash
# Xuất báo cáo HTML (full pipeline)
go run ./cmd/visualize/main.go <solution.drawio> <student.drawio> [output.html]

# So sánh CLI (chỉ in ra terminal, không xuất HTML)
go run ./cmd/compare/main.go <solution.drawio> <student.drawio>
```

---

## 3. Giải thích kết quả HTML Report

### A. Header — Score
Hiển thị điểm tổng, điểm tối đa, và phần trăm chính xác kèm progress bar.

### B. Nodes Comparison (Side-by-Side)
Bảng 2 cột: **Student** (bài nộp) và **Solution** (đáp án).

Mỗi node (Class/Interface/Enum) hiển thị:
- Attributes và Methods với color-coded status:
  - 🟢 **Correct** (xanh): Khớp hoàn toàn.
  - 🟡 **Wrong** (vàng): Tồn tại nhưng sai (sai tên, sai type...).
  - 🔴 **Missing** (đỏ): Thiếu so với đáp án.
  - 🔵 **Extra** (xanh dương): Thừa, không có trong đáp án.

### C. Relations
Liệt kê tất cả mối quan hệ (Inheritance, Association, Aggregation...):
- ✅ Correct: Đúng.
- ⚠️ Wrong: Sai loại hoặc hướng.
- ❌ Missing: Thiếu trong bài sinh viên.
- ➕ Extra: Thừa trong bài sinh viên.

### D. Summary
4 ô thống kê: **Correct** / **Missing** / **Wrong** / **Extra** + danh sách chi tiết trừ điểm.

---

## 4. Các công cụ Debug khác

| Công cụ | Lệnh | Mục đích |
|:---|:---|:---|
| **Compare CLI** | `go run ./cmd/compare/main.go` | So sánh full pipeline, in ra terminal |
| **Match CLI** | `go run ./cmd/match/main.go` | Kiểm tra Mapping Table |
| **PreMatch CLI** | `go run ./cmd/prematch/main.go` | Xem kết quả tiền xử lý |
| **Build CLI** | `go run ./cmd/build/main.go` | Xem UMLGraph sau khi build |
| **Parse CLI** | `go run ./cmd/parse/main.go` | Xem XML thô đã decode |

---

## 5. Lưu ý
- Đảm bảo các tệp `.drawio` được export từ [diagrams.net](https://diagrams.net) và không bị lỗi định dạng.
- Báo cáo HTML tự chứa (self-contained) — không cần internet, mở bằng bất kỳ trình duyệt nào.
- Màu sắc CLI hiển thị tốt nhất trên Terminal hỗ trợ ANSI (Windows Terminal, PowerShell Core).
