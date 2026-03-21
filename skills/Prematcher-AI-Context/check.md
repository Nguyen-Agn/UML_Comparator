# Verification: Prematcher Quality Checklist

Dùng checklist này để tự kiểm tra trước khi hoàn tất task:

## 1. Structural Correctness
- [ ] Interface `IPreMatcher` được implement đầy đủ không?
- [ ] Biến chứng thực `var _ IPreMatcher = (*Struct)(nil)` có hiện diện không?
- [ ] `ProcessedUMLGraph` có đủ số lượng Node và Edge như bản gốc không?

## 2. Parsing Accuracy
- [ ] Các thuộc tính có scope (+, -, #) được nhận diện đúng không?
- [ ] Tên và Kiểu của thuộc tính/phương thức có bị dính dấu `:` không?
- [ ] Các tham số của phương thức (`Inputs`) có được parse đầy đủ không?

## 3. Architecture Logic
- [ ] `ArchWeight` có được tính toán dựa trên các hằng số bitmask không?
- [ ] Trường `Inherits` có trỏ đúng ID của lớp cha không?
- [ ] Trường `Implements` có chứa đủ danh sách ID các interface không?

## 4. Code Quality
- [ ] Code có dùng `domain` models thay vì tự định nghĩa struct trùng lặp không?
- [ ] Có xử lý lỗi `error` trả về từ các hàm helper không?
- [ ] Có comment giải thích các regex phức tạp không?
