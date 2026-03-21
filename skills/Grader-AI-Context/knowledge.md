# Logic Tính điểm & Ngưỡng chặn
Điểm môn học Không bao giờ được phép < 0. Bắt buộc dùng 'math.Max(0, score)'. 
> Chú ý Floating Point cùa Go: 10 - (0.1 * 3) = 9.6999999994. 
Bắt buộc dùng hàm làm tròn roundFloat 2 chữ số thập phân trước khi tống vào GradeResult.
