# RECORRECT: Evolution Layer

(File này dùng để Agent tự giác ghi log bài học hệ thống (Lesson learned) nếu quá trình thực thi bị đánh sập trong các bài test hoặc bị chỉ trích từ phía User do cấu trúc bị sai/vi phạm luồng gốc).

## Error Logs & Mitigations:

*Tự động ghi nối vào danh sách này nếu quá trình refactoring gây lỗi:*

### [Format ghi]
- **Date / Context**: VD: Yêu cầu refactor Module A.
- **Root cause / Lỗi**: VD: Đã gộp 2 xử lý nghiệp vụ của Parser vào chung 1 file và dùng Struct thay vì Interface cho đối tượng Checker, vi phạm OCP và không thể Inject Mock ở Unit Test.
- **Action/Rule rút ra**: VD: Bắt buộc nhẩm lại nguyên tắc: "Dependencies truyền vào hàm bắt buộc phải là Interface type. Nếu bên ngoài chưa có Interface, TỰ VIẾT THÊM Interface cho nó ở phía Consumer."
