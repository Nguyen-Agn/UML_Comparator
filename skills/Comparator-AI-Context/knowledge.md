# Logic So sánh Mũi tên và Mảng (Advanced Comparator)

## 1. TypeMap (Ánh xạ Kiểu Dữ liệu)
Trong UML, kiểu dữ liệu tham số hoặc trả về thường mang tên của một Class khác (vd: `Ship`, `User`). Vì sinh viên có thể đổi tên Class (vd: `Ship` -> `PPShip`), Comparator CẦN xây dựng một `TypeMap` (Từ Solution Name -> Student Name dựa trên `MappingTable`) trước khi bắt đầu so sánh. Khi so sánh kiểu dữ liệu, nếu kiểu của Solution trùng với một key trong TypeMap, phải dịch nó sang tên của Student trước khi so.

## 2. Quy tắc So sánh Method
Với mọi method, đây là mảng không quan tâm thứ tự (Unordered Set). Hai method được coi là `Matched` nếu:
- **Kiểu trả về (ReturnType)**: Giống nhau tuyệt đối (Sau khi đã qua TypeMap).
- **Tham số (Params) Count**: Giống nhau. Nếu cả hai >= 2 params, chấp nhận +-1 (sẽ được ghi nhận là sai sau khi matched).
- **Tên (Name)**: Tương tự >= 0.5 (Dùng FuzzyMatcher).
- **Sau khi match**, kiểm tra chi tiết: Scope, Kind, Params chính xác -> báo vào `WrongDetail` nếu sai.
- **Trường hợp Constructor**: Nếu Solution KHÔNG có constructor -> Mặc định cho full điểm phần constructor (Bỏ qua). Nếu CÓ constructor -> So sánh như method bình thường, nhưng KHÔNG quan tâm thứ tự params.
- **Getter/Setter**: Không cần so sánh chi tiết tên/param/scope. Chỉ cần lấy tổng count Getter của Solution so với Student, lệch bao nhiêu báo bấy nhiêu.

## 3. Quy tắc So sánh Attribute
Mảng thuộc tính cũng không quan tâm thứ tự. Hai attribute được coi là `Matched` nếu:
- **Kiểu dữ liệu (Type)**: Khớp tuyệt đối (Sau khi qua TypeMap) — kiểm tra TRƯỚC.
- **Tên (Name)**: `FuzzyScore >= 0.5` HOẶC một chuỗi `Contains` chuỗi còn lại.
- **Sau khi match**, kiểm tra Scope, Kind -> báo vào `WrongDetail` nếu sai.

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
