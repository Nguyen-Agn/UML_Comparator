# SKILL: batch-grader-reporter
## Purpose
Cung cấp chức năng chấm điểm đồng loạt (Batch Grading) cho nhiều file bài làm UML (`.drawio`) của sinh viên đối chiếu với 1 file đáp án mẫu. 
Kết quả của tất cả các bài nộp được tổng hợp lại thành `BatchGradeResult` và chuyển đền interface `IReporter` để sinh report.

## Execution
1. Cấu trúc: 
   - Load file solution `solution.drawio`.
   - Quét thư mục `student_dir` tìm các file `.drawio`.
   - Với mỗi file: Parse → Build → PreMatch → Match → Compare → Grade.
   - Ghi nhận `domain.GradeResult` vào danh sách.
   - Gọi `IReporter.GenerateReport(batchResult)` để tổng hợp.
2. Interface `IReporter`:
   - Nằm tại `report/reporter.go`.
   - Cho phép cắm (plug-in) nhiều loại định dạng report.
   - Hiện tại support `ConsoleReporter` (in log terminal) và `CSVReporter` (xuất file `.csv` bảng tính).
   - `CSVReporter`: Thiết kế dạng ma trận (Dòng: Tên sinh viên, Cột: Các class/attribute/method/relation mẫu). Mỗi ô mang giá trị 1 (có/đúng) hoặc 0 (sai/thiếu). Cột điểm số và tỷ lệ đã được loại bỏ theo yêu cầu, chỉ chừa cột Student ID, Status và list member.

## CLI Usage
```
go run ./cmd/grade_batch/main.go <solution.drawio> <student_dir>
```
Output: Tạo ra một file `batch_report.csv` trực tiếp tại thư mục làm việc, mã hóa các lỗi dưới dạng 0/1, sẵn sàng cho người chấm mở trên Excel để phân tích nhanh độ phủ bài làm.
