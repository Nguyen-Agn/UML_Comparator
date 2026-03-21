# Hướng Dẫn Vận Hành Hệ Thống Skills Dành Cho Agent

Chào Agent, đây là hệ điều hành "Personal Skill OS" của dự án UML Compare. 
Mục tiêu là giúp hệ thống AI (như bạn) hoạt động chính xác, có ngữ cảnh tốt nhất và tự động kiểm chứng (Verification) trước khi báo cáo kết quả. Nó chuyển đổi tư duy từ "Prompt-based" sang "Skill-based execution".

## Cơ chế hoạt động (The 5-Layer Framework)
Dự án được triển khai theo các cụm Module (VD: Parser, Builder, Matcher). Mỗi module trọng tâm sẽ có một thư mục Skill riêng biệt. 

Khi bạn nhận lệnh để code hoặc phân tích một Module, bạn **BẮT BUỘC** phải đọc các file trong thư mục Skill tương ứng trước khi thao tác.

### Cấu trúc chuẩn của một thư mục Skill bao gồm:
1. **SKILL.md (Intent Layer):** Trạm kiểm soát đầu tiên. Luôn đọc file này đầu tiên để biết Mục tiêu, Đầu vào, Đầu ra kỳ vọng và chiến lược thực thi tổng quát.
2. **knowledge.md (Knowledge Layer):** Chứa tài nguyên, nguyên tắc kinh doanh (Business Rules), mô tả chuẩn Format (VD: Drawio XML Format) để cấp Domain Knowledge cho bạn.
3. **template.md (Execution / Scaffolding Layer):** Không cần tự bịa ra cấu trúc code mới. Dùng các mẫu Pattern chuẩn (VD: SOLID, Interface parsing) được ép khuôn ở đây để giảm không gian logic, ép Agent triển khai đúng định dạng.
4. **check.md (Verification Layer):** Cực kỳ quan trọng. Agent dùng các checklist trong đây để tự đánh giá lại Code/Output mình vừa sinh ra có đạt chuẩn chữ "Done" (Thiết kế đúng) hay không.
5. **recorrect.md (Evolution Layer):** Khi code sinh ra bị bug, chạy test fail, hoặc người dùng chỉ ra lỗi sai. Tự giác ghi log (thói quen, nguyên nhân) vào file này để rút kinh nghiệm về sau.

## Vòng Lặp Hoạt Động Của Agent (S.S.E.V.E):
1. **SCOPE:** Nhận bài toán từ User.
2. **SKILL:** Tìm vào thư mục Skill tương ứng, đọc liền `SKILL.md` và `knowledge.md`.
3. **EXECUTE:** Viết code / Xử lý logic theo sát dàn giáo ở `template.md`.
4. **VERIFY:** Dùng `check.md` để test lại toàn bộ đầu ra. Nếu sai, quay lại vòng 3.
5. **EVOLVE:** Ghi lỗi vấp phải vào `recorrect.md` để phiên bản "lần sau" hiểu biết sâu hơn.
