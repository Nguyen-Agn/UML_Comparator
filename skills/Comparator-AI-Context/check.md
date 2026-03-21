# Verification Checklist (Class Review)
- [ ] Hàm `Compare` có nhận tham số đúng là `ProcessedUMLGraph` thay vì `UMLGraph` nguyên thủy không?
- [ ] Báo cáo `DiffReport` có chia rõ 6 mảng phân loại (3 mảng error, 3 mảng missing) không?
- [ ] Đã xây dựng và sử dụng `TypeMap` (ánh xạ SolutionClass -> StudentClass) khi so sánh Type của Param, Return, và Attribute chưa?
- [ ] Constructor có bật chế độ Auto-Pass nếu Solution không có, và so sánh Params không thứ tự (Unordered Params) nếu Solution có yêu cầu không?
- [ ] Các Method và Attribute có soi xét kỹ Phạm vi truy cập (`Scope`: +, -, #, ~) không?
- [ ] Method thường có so sánh Params theo đúng thứ tự (Strict Order) và khớp type không?
- [ ] Getter/Setter có được tính riêng (chỉ đếm số lượng) thay vì so sánh cứng nhắc không?
- [ ] Bắt được lỗi thiếu Mũi tên và Mũi tên ngược chiều (Reverse Arrow) không?
- [ ] So sánh mảng thuộc tính có vượt qua được bẫy thứ tự (Unordered Arr) không?
