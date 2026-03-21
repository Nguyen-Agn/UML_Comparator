# Logic So sánh Mũi tên và Mảng (Advanced Comparator)

## 1. TypeMap (Ánh xạ Kiểu Dữ liệu)
Trong UML, kiểu dữ liệu tham số hoặc trả về thường mang tên của một Class khác (vd: `Ship`, `User`). Vì sinh viên có thể đổi tên Class (vd: `Ship` -> `PPShip`), Comparator CẦN xây dựng một `TypeMap` (Từ Solution Name -> Student Name dựa trên `MappingTable`) trước khi bắt đầu so sánh. Khi so sánh kiểu dữ liệu, nếu kiểu của Solution trùng với một key trong TypeMap, phải dịch nó sang tên của Student trước khi so.

## 2. Quy tắc So sánh Method
Với mọi method, đây là mảng không quan tâm thứ tự (Unordered Set). Hai method được coi là `Matched` nếu:
- **Phạm vi truy cập (Scope)**: Khớp tuyệt đối (`+`, `-`, `#`, `~`).
- **Tên (Name)**: Tương tự >= 0.5 (Dùng FuzzyMatcher).
- **Kiểu trả về (ReturnType)**: Giống nhau tuyệt đối (Sau khi đã qua TypeMap).
- **Tham số (Params)**: KHÔNG xét tên tham số (Params Name). KHÁT KHE thứ tự và kiểu của tham số (Type sau TypeMap).
- **Trường hợp Constructor**: Nếu Solution KHÔNG có constructor -> Mặc định cho full điểm phần constructor (Bỏ qua). Nếu CÓ constructor -> So sánh như method bình thường, nhưng KHÔNG quan tâm thứ tự params.
- **Getter/Setter**: Không cần so sánh chi tiết tên/param/scope. Chỉ cẩn lấy tổng count Getter của Solution so với Student, lệch bao nhiêu báo bấy nhiêu.

## 3. Quy tắc So sánh Attribute
Mảng thuộc tính cũng không quan tâm thứ tự. Hai attribute được coi là `Matched` nếu:
- **Phạm vi truy cập (Scope)**: Khớp tuyệt đối (`+`, `-`, `#`, `~`).
- **Kiểu dữ liệu (Type)**: Khớp tuyệt đối (Sau khi qua TypeMap).
- **Tên (Name)**: `FuzzyScore >= 0.5` HOẶC một chuỗi `Contains` chuỗi còn lại (Ví dụ: `age` vs `ageField` -> Hợp lệ do lỗi viết thừa).

## 4. Reverse Arrow (Mũi tên ngược)
Mẫu: 'Class A' --> 'Class B'. 
Sinh viên vẽ: 'Class mapped(B)' --> 'Class mapped(A)'. Đây là lỗi ngược chiều, ghi nhận lỗi `[Reverse Arrow]`.

## 5. Output
Trả về list chi tiết báo cáo sai lệch vào `DetailedErrors`, ghi rõ đích danh Class nào bị lỗi gì (VD: `Class 'Ship': Thiếu parameter thứ 2 của method move()`).
