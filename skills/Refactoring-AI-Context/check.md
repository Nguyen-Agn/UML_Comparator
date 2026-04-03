# CHECK: Verification Layer

Sau khi đã sinh code thực thi Refactor, Agent hãy dùng checklist này để tự đối chiếu xem kết quả code vừa tạo đã thoả mãn các yêu cầu hệ thống ở mức nghiêm ngặt nhất chưa. Tuyệt đối không giao nộp nếu có bất cứ mục nào chưa hoàn thành.

## Checklist Tự Đánh Giá (Self-Correction Matrix):

- [ ] **Interface-First Checklist:** Toàn bộ quá trình refactor đều bắt đầu từ định nghĩa Interface? Các struct có giấu Implementation đi không (chỉ export interface và method constructor `New...`)? Không có struct nào gọi và nói chuyện trực tiếp với Struct của bên ngoài không?
- [ ] **Comments & Contract Checklist:** Từng method bên trong Interface vừa được bạn thiết kế đã có **comment** giải thích một cách rõ ràng (đầu vào, đầu ra, side-effect giả định) chưa?
- [ ] **Open/Closed Principle (OCP) Checklist:** Thiết kế mã nguồn có sẵn sàng mở cho việc thêm mới mà không cần chỉnh sửa lại mã lõi không? Giả sử khi thêm 1 case / 1 function mới, ta có cần phải mò vào file struct cũ sửa `if-else` không?
- [ ] **Dependency Injection Checklist:** Struct có đang khởi tạo lớp ngoài (hardcode dependency instantiate) bên trong nó không? Có truyền Dependency (dạng interface) vào qua tham số constructor không?
- [ ] **Unit Test Checklist:** Refactoring đã kèm Unit Test đầy đủ cho mọi component public? Sử dụng format Table-Driven Test chưa?
- [ ] **Global Constraints Checklist:** Nhìn lại sơ đồ tác động - việc bạn thay đổi module này, có làm phá hỏng/phải chỉnh sửa chéo qua các modules khác (trái User rule) một cách không cần thiết không? 
- [ ] **Compile Validation:** Đã phân tích/thử logic để chắc chắn code không bị Circular Dependency chưa? Đã run go build (hoặc tự đọc rà soát syntax) không cấn lỗi cú pháp nào chưa?
