# Logic So sánh Mũi tên và Mảng (Advanced Comparator)

## 1. TypeMap (Ánh xạ Kiểu Dữ liệu)
Trong UML, kiểu dữ liệu tham số hoặc trả về thường mang tên của một Class khác (vd: `Ship`, `User`). Vì sinh viên có thể đổi tên Class (vd: `Ship` -> `PPShip`), Comparator CẦN xây dựng một `TypeMap` (Từ Solution Name -> Student Name dựa trên `MappingTable`) trước khi bắt đầu so sánh. Khi so sánh kiểu dữ liệu, nếu kiểu của Solution trùng với một key trong TypeMap, phải dịch nó sang tên của Student trước khi so.

## 2. Quy tắc So sánh Type (Generic & Simple)
- **Simple Types**: Phải khớp sau khi mapping qua `TypeMap`.
- **Generic Types (`List<T>`, `Map<K, V>`)**:
  - **Outer Match**: Dùng quy tắc `contains` (case-insensitive). Ví dụ: `List` khớp với `ArrayList`, `Map` khớp với `HashMap`.
  - **Inner Match**: Các tham số bên trong `< >` được so sánh đệ quy sau khi mapping qua `TypeMap`.
  - Để một kiểu Generic được coi là "Match", cả phần Outer và TOÀN BỘ phần Inner đều phải khớp.

## 3. Quy tắc Tiered Matching (Phân tầng khớp)
Để cung cấp feedback chi tiết "Cái gì so với cái gì", Comparator áp dụng 3 tầng khớp cho Attributes và Methods của một Class đã được xác định:
1. **Tầng 1: Perfect Match**: Khớp cả Tên (Exact/Fuzzy) và Kiểu dữ liệu (Sau mapping).
2. **Tầng 2: Signature Match**: Khớp Tên và Số lượng tham số (áp dụng quy tắc +-1 cho Method), dù kiểu dữ liệu có thể khác.
3. **Tầng 3: Name-only Match**: Chỉ cần Tên tương đồng cao (>= 0.8), dùng để bắt cặp các thành phần chắc chắn là cùng một thực thể nhưng sai kiểu dữ liệu.

**Kết quả sau khi bắt cặp (Matched):**
- Nếu khớp hoàn toàn -> `CorrectDetail`.
- Nếu có sai lệch (Sai kiểu dữ liệu, sai Scope, sai Kind) -> `WrongDetail` kèm mô tả chi tiết lỗi.
- Nếu không thể bắt cặp qua cả 3 tầng -> Coi là `MissingDetail` (Thiếu) hoặc `ExtraDetail` (Thừa).

## 4. Reverse Arrow & Edge Errors (Mũi tên ngược & Sai kiểu)
- **Exact match**: SourceID + TargetID + RelationType đều đúng -> `CorrectDetail`.
- **Wrong type**: SourceID + TargetID đúng nhưng RelationType sai -> `WrongDetail` ghi rõ (Solution: X, Student: Y).
- **Reverse arrow**: Source và Target bị đảo ngược, RelationType đúng -> `WrongDetail` ghi `Reverse arrow`.
- **Missing**: Không tìm thấy gì -> `MissingDetail`.

## 5. Output — DiffReport (Object references)
Trả về `*domain.DiffReport` với 4 nhóm phân loại: `MissingDetail`, `WrongDetail`, `ExtraDetail`, `CorrectDetail`.

Mỗi nhóm là `DetailError` struct gồm các slice của các struct Diff sau:
- `NodeDiff`: `{ Sol *ProcessedNode, Stu *ProcessedNode, Description string }`
- `AttributeDiff`: `{ ParentClassName string, Sol *ProcessedAttribute, Stu *ProcessedAttribute, Description string }`
- `MethodDiff`: `{ ParentClassName string, Sol *ProcessedMethod, Stu *ProcessedMethod, Description string }`
- `EdgeDiff`: `{ Sol *ProcessedEdge, Stu *ProcessedEdge, Description string }`

**Lưu ý cho Grader:**
- Nếu `Sol != nil && Stu == nil` -> Item bị thiếu ở bài làm sinh viên.
- Nếu `Sol == nil && Stu != nil` -> Sinh viên viết thừa item.
- Nếu cả hai `!= nil` -> Có sự sai lệch chi tiết (mô tả trong `Description`).
- Grader có thể truy cập `Sol.ArchWeight` hoặc `Stu.Shortcut` trực tiếp để tính điểm thưởng/phạt chính xác.
