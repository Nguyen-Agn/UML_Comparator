# SKILL: rule-based-grader
## Purpose
Tính toán điểm số cuối cùng dựa trên các list lỗi của DiffReport. 
Tính toán bằng lấy điểm chuẩn trừ đi điểm lỗi (miss/wrong).
+ Các điểm lỗi thông qua {
    edge = 1, 
    node = 1,//(class, interface, enum,...)
    attribute = 1,
    method = 1,
}
Điểm tối đa mặc định là 20, truyến vào qua Dependency Injection



## Execution
Gán điểm tuyệt đối ban đầu (VD: 10đ). 
Trích DiffReport, cứ mỗi MissingNodes duyệt trừ đi penalty định biên.
Chặn mốc Max(0, Score) để điểm không bao giờ âm sâu.
