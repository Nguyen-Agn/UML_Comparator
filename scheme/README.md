# Catalogue Kiến Trúc Dữ Liệu (Data Schemas)

Thư mục `scheme/` lưu trữ toàn bộ định nghĩa kiểu dữ liệu (Data Models) và các mẫu cấu hình chuẩn được sử dụng xuyên suốt hệ thống **UML Compare**.

## Mục đích
- Là *nguồn sự thật duy nhất (Single Source of Truth)* cho tất cả các cấu trúc dữ liệu.
- Giúp các module giao tiếp với nhau đúng contract mà không cần đọc source code Go.
- Làm tài liệu tham chiếu khi viết test hoặc tạo file cấu hình đầu vào.

## Danh sách Files

| File | Mô tả |
|---|---|
| `raw_xml_data.md` | Mô tả kiểu `RawXMLData` — output của Parser |
| `uml_graph.md` | Mô tả kiểu `UMLGraph`, `UMLNode`, `UMLEdge` — output của Builder |
| `diff_report.md` | Mô tả kiểu `DiffReport` — output của Comparator |
| `grade_result.md` | Mô tả kiểu `GradeResult` — output của Grader |
| `grading_rules.json` | Mẫu file cấu hình điểm (input của Grader) |
