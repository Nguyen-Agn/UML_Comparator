# Cẩm Nang Kiểm Chứng (Verification Checklist)

Mục đích của file này (Verification Layer) là ép Agent phải TỰ DUYỆT lại Code của mình trước khi báo cáo kết quả "Hoàn thành" công việc với người dùng.

## 1. Correctness (Kiểm tra Tính đúng đắn)
- [ ] Code có đọc được file vật lý mà không bị lỗi cấp quyền (Permission/OS Error) không?
- [ ] Chuỗi trả về `domain.RawXMLData` đã lột sạch lớp `<mxfile>` và lấy được trúng phóc nội dung XML `<mxGraphModel>...</mxGraphModel>` chưa?
- [ ] Đạt chuẩn `var _ IFileParser = (*DrawioParser)(nil)` trong `template.md` chưa?

## 2. Completeness (Kiểm tra Tính Đầy đủ)
- [ ] Module đã cover được trường hợp tham số `filePath` rỗng (`""`) chưa?
- [ ] Code có hỗ trợ nhánh if-else để bao bọc CẢ 2 ĐỊNH DẠNG: Compressed (`jZLB...`) và Uncompressed (XML rành rành) chưa? Thiếu 1 trong 2 là lỗi nặng.

## 3. Consequence (Đánh giá Rủi ro Thực chiến)
- [ ] Các thông báo lỗi (Error Message) trả về có CỤ THỂ không?
  - *Sai:* `return err` 
  - *Đúng:* `fmt.Errorf("DrawioParser: failed to base64 decode diagram payload: %w", err)`
- [ ] Khi lập trình IO mở file, có dùng `defer` đóng file lại không? (Nếu dùng `os.ReadFile` thì an toàn, nhưng dùng `os.Open` phải đóng).

---
**Hành động ép buộc sau khi code xong:** 
Agent không được chỉ code suông. Bắt buộc phải tự tạo hoặc yêu cầu User cấp một file `.drawio` thực tế nhét vào folder `testdata/` để viết Unit Test `TestDrawioParser_Parse` trước khi sang module Matcher/Comparator. Chạy Pass thì mới được báo xong việc!
