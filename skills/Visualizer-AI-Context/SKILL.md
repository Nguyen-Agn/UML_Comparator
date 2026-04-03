# SKILL: drawio-visualizer
## Purpose
Render kết quả chấm UML ra 2 loại file HTML self-contained:
1. **Grader Report** (`report_*.html`): Full side-by-side Student vs Solution, summary, deduction feedbacks — dành cho người chấm.
2. **Student Feedback** (`feedback_*.html`): Chỉ hiển thị bài làm SV với color-coded status (đúng/sai/thừa). Không lộ đáp án, không hiện deduction details — gửi cho sinh viên xem lại.

## Execution
1. Pipeline: Parse → Build → Validate → PreMatch → Match → Compare → Grade
2. `HTMLVisualizer.ExportHTML(gradeResult, path)` → Grader report
3. `HTMLVisualizer.ExportStudentHTML(gradeResult, path)` → Student feedback
4. Color scheme: Xanh (#d5e8d4) = Correct, Đỏ (#f8cecc) = Missing, Vàng (#ffe6cc) = Wrong
5. Auto-open browser trên Windows/macOS/Linux.

## CLI Usage
```
go run ./cmd/visualize/main.go <solution.drawio> <student.drawio> [output.html]
```
Output mặc định:
- `report_<student_filename>.html` — Grader report
- `feedback_<student_filename>.html` — Student report
