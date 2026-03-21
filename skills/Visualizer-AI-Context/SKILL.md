# SKILL: drawio-visualizer
## Purpose
Render và sinh ra hệ file báo cáo (Trực tiếp gán đè vào .drawio XML hoặc xuất ra HTML Mermaid) nhằm tô màu các Entity lỗi để sinh viên xem lại trực quan bài sai điểm nào.

## Execution
Dùng String Replace quét vào chính RawXMLData lúc đầu, thay thế chuỗi thuộc tính 'style="..."' thêm 'fillColor=#f8cecc' (Đỏ) ứng với ID bị báo lỗi từ DiffReport.
Xuất file write ra đĩa.
