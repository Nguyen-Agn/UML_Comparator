# Hướng dẫn sử dụng UML Comparator

`UML Comparator` là công cụ so sánh và chấm điểm biểu đồ UML từ file `.drawio`. Công cụ phân tích cấu trúc Class, Attributes, Methods, và Relations, sau đó xuất báo cáo HTML trực quan.

---

## 1. Quick Start — Thư mục `portable/`

Bạn không cần cài đặt bất kỳ môi trường nào. Chỉ cần mở thư mục `portable/` và chọn phiên bản phù hợp:

### A. Phiên bản Tích hợp (`exam_student_uml.exe`) — Khuyên dùng cho thi cử
Phiên bản này đã được giảng viên nén sẵn các file đáp án.
1. **Mở file** `exam_student_uml.exe`.
2. Giao diện sẽ hiển thị danh sách thả xuống ở phần **Embedded Solution**.
3. Chọn câu hỏi cần chấm (ví dụ: Q1, Q2...).
4. Nhấn **CHOOSE ASSIGNMENT** và chọn bài làm của bạn (`.drawio`).
5. Nhấn **RUN ANALYSIS**.

### B. Phiên bản Đồ họa chuẩn (`student_uml.exe`)
Dùng khi bạn có cả file đáp án và file bài làm riêng biệt.
1. **Mở file** `student_uml.exe`.
2. Nhấn **CHOOSE SOLUTION** chọn file đáp án (`.drawio`).
3. Nhấn **CHOOSE ASSIGNMENT** chọn bài làm sinh viên.
4. Nhấn **RUN ANALYSIS**.

### C. Phiên bản CLI (`student_uml_cli.exe`) — Dòng lệnh
1. Chạy thông qua PowerShell hoặc CMD: `./student_uml_cli.exe <solution.drawio> <student.drawio>`
2. Kết quả sẽ in trực tiếp ra Terminal và xuất file HTML.

## 2. Dành cho Giảng viên (Lecturers)

### A. Quản lý Build Tập trung (`test_orchestrator.exe`)
Đây là "bộ não" giúp giảng viên tạo ra các bản build chuyên biệt.
1. Chạy file `test_orchestrator.exe`.
2. Chọn **Option [1] Universal Build**: Để build lại toàn bộ bộ tool chuẩn (CLI, GUI, Batch Mode).
3. Chọn **Option [2] Exam Build**: Để tạo bản build nén sẵn đáp án cho sinh viên.
   - Nhập đường dẫn đến file hoặc folder chứa các tệp `.drawio` đáp án.
   - Bạn có thể nạp nhiều file cùng lúc (ví dụ: `bai1.drawio`, `bai2.drawio`).
   - Kết quả sẽ nằm trong thư mục `portable/exam_student_uml.exe`.

### B. Chấm điểm hàng loạt (`lecture_cli_parallel.exe`)
1. Chạy lệnh: `./lecture_cli_parallel.exe <solution.drawio> <folder_sinh_vien>`
2. Công cụ sẽ quét toàn bộ và xuất ra file `batch_report.csv` cực nhanh nhờ cơ chế song song.

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

## 3. Chạy từ mã nguồn (Go)

```bash
# Chạy bộ điều phối Build (Dành cho Giảng viên)
go run ./cmd/builder_exe/main.go

# Chạy trực tiếp GUI thi cử (Dành cho test)
go run ./cmd/exam_gui/main.go

# Xuất báo cáo HTML chuẩn
go run ./cmd/visualize/main.go <solution.drawio> <student.drawio>
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
