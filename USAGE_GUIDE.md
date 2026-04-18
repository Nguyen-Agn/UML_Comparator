# Hướng dẫn sử dụng UML Comparator

`UML Comparator` là công cụ so sánh và chấm điểm biểu đồ UML từ file `.drawio`. Công cụ phân tích cấu trúc Class, Attributes, Methods, và Relations, sau đó xuất báo cáo HTML trực quan.

---

## 1. Quick Start — Thư mục `portable/`

Bạn không cần cài đặt bất kỳ môi trường nào. Chỉ cần mở thư mục `portable/` và chọn phiên bản phù hợp:

### A. Phiên bản Tích hợp (`exam_student_uml.exe`) — Khuyên dùng cho nén file solution
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

Từ phiên bản mới nhất, tất cả các công cụ dành cho giảng viên đã được tích hợp vào một giao diện duy nhất: **Instructor Suite**.

### A. Sử dụng Instructor Suite (`instructor_suite.exe`)
1. **Mở file** `instructor_suite.exe` trong thư mục `portable/`.
2. Giao diện bao gồm 4 chức năng chính (Tab):
   - **Live Compare**: So sánh trực tiếp biểu đồ Solution và Student với đầy đủ thông số kỹ thuật (Admin mode).
   - **Batch Grader**: Chọn 1 file đáp án và 1 thư mục chứa bài làm sinh viên để tự động chấm điểm toàn bộ và xuất ra file `batch_result.csv`.
   - **Solution Encrypt**: Mã hóa file `.drawio` thành file `.solution` để chia sẻ an toàn cho sinh viên.
   - **Exam Builder**: Nén các file đáp án vào một bản build GUI `exam_student_uml.exe` duy nhất dành riêng cho kỳ thi.

### B. Quản lý Build Cốt lõi (`test_orchestrator.exe`)
Nếu bạn muốn tự compile lại hệ thống:
1. Chạy file `test_orchestrator.exe`.
2. Chọn **Option [1] Universal Build**: Để build lại toàn bộ bộ tool chuẩn (CLI, GUI, Batch Mode).
3. (Hoặc bạn có thể dùng Tab **Exam Builder** bên trong `instructor_suite.exe` thay vì dùng Orchestrator).

---

## 2. Sử dụng tham số (Command Line)
Nếu bạn muốn dùng lệnh trong PowerShell hoặc CMD (dành cho automation):

**Dành cho Sinh viên:**
```bash
./student_uml_cli.exe <solution.drawio> <student.drawio> [output.html]
```
> **Output**: tạo ra 1 file html chứa kết quả của sinh viên.

**Dành cho Sinh viên (CLI Standalone):**
```bash
./student_uml_cli.exe <solution.drawio> <student.drawio> [output.html]
```
> **Output**: tạo ra 1 file html chứa kết quả của sinh viên.

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


# Build Exe

## Instructor Suite
```bash
go build -ldflags="-H windowsgui" -o portable\instructor_suite.exe .\cmd\instructor\main.go
```

## Student GUI
```bash
go build -ldflags="-H windowsgui" -o portable\student_uml.exe .\gui\main.go
```

## Student CLI
```bash
go build -o portable\student_uml_cli.exe .\cmd\visualize\main.go .\cmd\visualize\interactive.go
```