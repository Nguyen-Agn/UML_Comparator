# Evolution Log
1. **Unordered Match (Cạm bẫy vòng lặp)**: Lỗi phổ biến Agent hay mắc là dùng vòng lặp `for i := range arr1` để so sánh với `arr2[i]`. Điều này làm thuật toán comparator sụp đổ vì UML không quan trọng thứ tự khai báo hàm/thuộc tính. Phải so sánh mảng bằng logic 2 vòng lặp lồng nhau hoặc Map Counting.
2. **TypeMap Mapping**: Lỗi bỏ sót việc dịch type. Kiểu dữ liệu tham số hoặc trả về có thể là tên của một class khác bị sinh viên đổi tên. Lúc kiểm tra Type phải dò tìm qua từ điển `TypeMap` (được sinh ra từ `MappingTable`).
3. **Mũi tên ngược (Reverse Arrow)**: Khi dò tìm Edge, phải thử thêm trường hợp đảo Source/Target để vạch mặt sinh viên chỉ điểm sai hướng tham chiếu kế thừa/tập hợp.
4. **Constructor**: Constructor tham số không có thứ tự nhất định. Bỏ qua nếu Mẫu không có.
5. **Getters/Setters**: Không so chi tiết Scope/Type, chỉ đếm số lượng tổng bằng prefix `get` và `set`.
