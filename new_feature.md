Tôi muốn tạo 1 thêm tính năng *Tạo mẫu UML tự động*.
Ý Tưởng chính là Paste đề bài vào, và AI/Agent sẽ tạo ra uml mẫu theo chuẩn.
Ngoài ra, còn hiển thị và cho phép chỉnh sửa theo chuẩn và người dùng( giảng viên) nhập điểm của từng thành phần tương ứng. Nếu ổn, thì cho phép lưu lại dưới dạng file Solution chuẩn (mermaid hoặc drawio).

Giao diện:
1. Nút Option mở để nhập API. (lưu lại để dùng về sau)
2. Một ô text để paste đề bài.
3. Nút *Tạo mẫu UML tự động*.
4. Khu vực hiển thị và chỉnh sửa uml và điểm số.
5. Nút *Lưu file*.
6. Chọn History. (lưu lại bài đã tạo)
// 1->6 laf thứ tự ưu tiên phát triển trước ( 1 trước 6);

Mô tả:
- Khi người dùng click nút *Tạo mẫu UML tự động* Agent sẽ phân tích đề bài và tạo ra uml mẫu (mermaid) và dùng tận dụng module Parser và Builder đã có để có thể tạo ra các cấu trúc tương ứng hỗ trợ render.
- Sau khi tạo, người dùng có thể chỉnh sửa uml mẫu và điểm số tương ứng của từng thành phần (Class, Relationship, Attribute, Method).
- Nếu người dùng ấn nút *Lưu file*, AI sẽ lưu lại uml mẫu và điểm dưới dạng file Solution chuẩn (mermaid hoặc drawio). (ưu tiên phát triển mermaid trước trong giai đoạn đầu).
- Người dùng có thể mở History để xem lại các bài đã tạo. Và có thể load lại bài đã tạo. (không ưu tiên phát triển history trong giai đoạn đầu).

Kiến trúc dự kiến:
- 1 Module mới tên là "UML Generator"
- Module này sẽ chứa logic để phân tích đề bài và tạo ra uml mẫu.
- Module này sẽ chứa logic để lưu lại uml mẫu và điểm dưới dạng file Solution chuẩn (mermaid hoặc drawio).

Yêu cầu:
- Mọi thứ sẽ nằm trong thư mục module tương ứng với kiến trúc đã có (ở trên).
- 1 cmd/umlGenerator/main.go để chạy thử nghiệm. (open GUI)
- Sử dụng giao diện (style như các module đã có). [[gui\view\instructor_view.html]] và [[gui\view\main_window.html]]
