# SKILL: rule-based-grader
## Purpose
Tính toán điểm số cuối cùng dựa trên các list lỗi của DiffReport và bảng trừ điểm GradingRules.

## Execution
Gán điểm tuyệt đối ban đầu (VD: 10đ). 
Trích DiffReport, cứ mỗi MissingNodes duyệt trừ đi penalty định biên.
Chặn mốc Max(0, Score) để điểm không bao giờ âm sâu.
