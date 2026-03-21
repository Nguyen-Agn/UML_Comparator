# SKILL: data-flow-comparator

## Purpose (Mục đích)
Kiểm tra tính toàn vẹn luồng dữ liệu (Data Flow Integrity) trước khi triển khai module Comparator thực sự. Khi dữ liệu từ Builder ra bị biến dạng (như lỗi swimlane vừa phát hiện), Comparator sẽ chấm điểm sai hoàn toàn. Skill này định nghĩa cách xác minh rằng pipeline Parser→Builder đang chạy chính xác trước khi so sánh 2 Graph.

## Use When
- Vừa sửa xong `IModelBuilder` và cần kiểm tra output chưa bị méo mó trước khi code Matcher/Comparator.
- Muốn phân tích độ tin cậy của dữ liệu đầu vào cho Comparator.
- Có nghi ngờ rằng mô hình dữ liệu đang sai (ví dụ: count Nodes sai, Edge trỏ sai ID).

## Required Inputs
- `*domain.UMLGraph` (đã Build từ file đáp án mẫu)
- `*domain.UMLGraph` (đã Build từ bài sinh viên)

## Expected Output — Báo cáo Integrity Check:
1. **Node Count Match:** Cả 2 graph phải có số Node hợp lý (> 0)
2. **Edge Integrity:** Mọi `Edge.SourceID` và `Edge.TargetID` phải tồn tại trong danh sách `graph.Nodes` của chính graph đó
3. **Name Non-Empty:** Mọi `UMLNode.Name` phải không rỗng (nếu rỗng = bug ở Builder)
4. **Type Valid:** Mọi `UMLNode.Type` phải thuộc tập hợp hợp lệ: `["Class", "Interface", "Actor"]`

## Execution Approach
1. Chạy hàm `ValidateGraph(*domain.UMLGraph)` trên cả 2 graph.
2. Phân loại lỗi theo 4 tiêu chí trên.
3. Nếu có lỗi → Log vào DiffReport với prefix `[DATA_INTEGRITY_ERROR]` và dừng pipeline.
4. Nếu pass → Pipeline tiếp tục sang Matcher.

## Nguyên tắc vàng
Nếu luồng dữ liệu bị biến dạng (data corruption) ở tầng Builder → Comparator dùng data đó sẽ sinh ra kết quả sai lệch 100%. Luôn chạy validate trước khi so sánh.

## Edge Cases
- File sinh viên nộp biểu đồ RỖNG (0 class, 0 edge) → Phải báo 0 điểm ngay lập tức thay vì crash.
- Sinh viên dùng shape khác (không phải `umlClass` style) → Node bị nhận sai Type → Cần fallback detection.

## References
- `scheme/uml_graph.md`: Bảng RelationType hợp lệ và bất biến của UMLGraph
- `Skills/Builder-AI-Context/recorrect.md`: Log lỗi swimlane đã phát hiện và sửa 2026-03-20
