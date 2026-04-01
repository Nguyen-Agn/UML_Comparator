# SKILL: drawio-visualizer
## Purpose
Render kết quả chấm UML ra file HTML tự chứa (self-contained), hiển thị side-by-side Student vs Solution với color-coded status cho từng node, attribute, method, và relation.

## Execution
1. Pipeline chạy đầy đủ: Parse → Build → Validate → PreMatch → Match → Compare → Grade
2. Gọi `HTMLVisualizer.ExportHTML(gradeResult, outputPath)` — xuất file `.html` self-contained.
3. File HTML sử dụng dark theme, color scheme: Xanh (#d5e8d4) = Correct, Đỏ (#f8cecc) = Missing/Error, Vàng (#ffe6cc) = Wrong.
4. Auto-open browser trên Windows/macOS/Linux.

## CLI Usage
```
go run ./cmd/visualize/main.go <solution.drawio> <student.drawio> [output.html]
```
Nếu không truyền output path, mặc định: `report_<student_filename>.html`
