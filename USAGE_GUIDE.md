# Hướng dẫn sử dụng UML Comparator

`UML Comparator` là công cụ so sánh và chấm điểm biểu đồ UML từ file `.drawio`. Công cụ phân tích cấu trúc Class, Attributes, Methods, và Relations, sau đó xuất báo cáo HTML trực quan.

---

## 1. Quick Start — Thư mục `portable/`

Bạn không cần cài đặt bất kỳ môi trường nào. Chỉ cần mở thư mục `portable/` và chọn phiên bản phù hợp:

### A. Phiên bản Đồ họa (`student_uml.exe`) — Khuyến nghị
1. **Mở file** `student_uml.exe`.
2. Giao diện **Dawn's Berry** sẽ xuất hiện.
3. Nhấn **CHOOSE SOLUTION** và chọn file đáp án (`.drawio`).
4. Nhấn **CHOOSE ASSIGNMENT** và chọn bài làm của sinh viên (`.drawio`).
5. Nhấn **RUN ANALYSIS**. Kết quả sẽ hiển thị ngay lập tức trong ứng dụng.
6. (Tùy chọn) Nhấn **SAVE HTML** để lưu báo cáo ra file riêng.

### B. Phiên bản CLI (`student_uml_cli.exe`) — Cho sinh viên
1. **Mở file** `student_uml_cli.exe`.
2. Một cửa sổ Console xuất hiện với tiêu đề **UML Comparator**.
3. Nhập đường dẫn file **Solution**.
4. Nhập đường dẫn file **Assignment**.
5. Nhấn **RUN ANALYSIS**. Kết quả sẽ hiển thị ngay lập tức trong ứng dụng.
6. (Tùy chọn) Nhấn **SAVE HTML** để lưu báo cáo ra file riêng.

### C. Phiên bản Chấm điểm hàng loạt (`lecture_cli_parallel.exe`) — Cho giáo viên
1. **Mở file** `lecture_cli_parallel.exe`.
2. Một cửa sổ Console xuất hiện với tiêu đề **Lecture Edition (Parallel)**.
3. Nhập đường dẫn file **Solution**.
4. Nhập đường dẫn **Thư mục chứa bài làm** của sinh viên (Ví dụ: `D:\Lop_OOP\Bai_Tap_1`).
5. Công cụ sẽ quét toàn bộ file `.drawio` và chấm điểm **song song (Parallel)** cực nhanh.
6. Kết quả sẽ tự động được lưu vào file `batch_report.csv` và mở lên bằng Excel cho bạn xem ngay.

---

## 2. Sử dụng tham số (Command Line)
Nếu bạn muốn dùng lệnh trong PowerShell hoặc CMD (dành cho automation):

**Dành cho Sinh viên:**
```bash
./student_uml_cli.exe <solution.drawio> <student.drawio> [output.html]
```
> **Output**: tạo ra 1 file html chứa kết quả của sinh viên.

**Dành cho Giáo viên (Đồng loạt):**
```bash
./lecture_cli_parallel.exe <solution.drawio> <student_folder> [report.csv]
```
> **Output**: tạo ra 1 file csv chứa kết quả của tất cả sinh viên.
**Ví dụ:**
```bash
./student_uml_cli.exe [--admin] sol.drawio assignment1.drawio
```
> **Tip:** Thêm tham số `--admin` để xem báo cáo đầy đủ thông tin giải pháp (dành cho giáo viên).

---

## 2. Chạy từ mã nguồn (Go)

```bash
# Xuất báo cáo HTML (full pipeline)
go run ./cmd/visualize/main.go <solution.drawio> <student.drawio> [output.html]

# So sánh CLI (chỉ in ra terminal, không xuất HTML)
go run ./cmd/compare/main.go <solution.drawio> <student.drawio>
```

---

## 3. Các công cụ Debug khác (dành cho dev golang)

| Công cụ | Lệnh | Mục đích |
|:---|:---|:---|
| **Compare CLI** | `go run ./cmd/compare/main.go` | So sánh full pipeline, in ra terminal |
| **Match CLI** | `go run ./cmd/match/main.go` | Kiểm tra Mapping Table |
| **PreMatch CLI** | `go run ./cmd/prematch/main.go` | Xem kết quả tiền xử lý (node Graph) |
| **Build CLI** | `go run ./cmd/build/main.go` | Xem UMLGraph sau khi build (String Graph) |
| **Parse CLI** | `go run ./cmd/parse/main.go` | Xem XML thô đã decode |

---

## 5. Lưu ý
- Đảm bảo các tệp `.drawio` được export từ [diagrams.net](https://diagrams.net) và không bị lỗi định dạng.
- Báo cáo HTML tự chứa (self-contained) — không cần internet, mở bằng bất kỳ trình duyệt nào.
- Màu sắc CLI hiển thị tốt nhất trên Terminal hỗ trợ ANSI (Windows Terminal, PowerShell Core).
