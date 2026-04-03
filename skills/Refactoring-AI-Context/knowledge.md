# KNOWLEDGE: Nguyên tắc kinh doanh và Luật Refactoring

## 1. Nguyên Tắc Cốt Lõi (Business Rules & Core Rules)
Để áp dụng kỹ năng Refactoring theo SOLID, Agent bắt buộc phải tuân theo các User Rules gốc sau:

- **Interface-first:** Luôn luôn tư duy và thiết kế Interface trước tiên (Interface-first approach). Các class (Implementation) không bao giờ giao tiếp với nhau trực tiếp (concrete-to-concrete injection), mà bắt buộc phải thông qua interfaces.
- **Clear Comments:** Interface **bắt buộc** phải được comment rõ ràng (như godoc) cho TỪNG method (giải thích chi tiết đầu vào, đầu ra, và các hành xử dự kiến).
- **Trusted Contracts:** Giả định rằng Interface sẽ cung cấp hoặc trả về dữ liệu ĐÚNG CHÍNH XÁC như những gì comment đã ghi đè. Cứ dựa vào comment Interface mà code/test, không cần đi soi chéo vào bên trong implement của module khác.
- **Open/Closed Principle (OCP):** Cấu trúc lại module sao cho hệ thống luôn "Open for extension, but closed for modification". Khi có behavior/requirement mới, chúng ta chỉ việc tạo class/struct mới triển khai interface, chứ KHÔNG phải sửa vào core struct hiện tại. Tránh các chuỗi `if-else` hay `switch-case` khổng lồ, ưu tiên dùng Strategy pattern, Registry map...
- **Isolation:** Cố gắng tối đa KHÔNG sửa đổi các module khác bên ngoài (những module không nằm trong phạm vi cần refactor) trừ khi bắt buộc phải sửa để khớp với cấu trúc Interface mới của module hiện tại.

## 2. Quy ước Unit Test
- Mọi module sau khi refactor phải đi kèm với unit test (`*_test.go`).
- Sử dụng Table-Driven tests đối với Golang để cover được hàng loạt các test scenarios (happy case, fail case) một cách gọn gàng, rõ ràng.
- Gắn chặt với Interface: Khi viết unit test cho các thành phần có phụ thuộc module bên ngoài, hãy sử dụng Mock (hoặc struct Test/Fake tự tạo triển khai Interface đó) thay vì gọi trực tiếp vào concrete implementation thật.

Tiếp theo: Chuyển sang `template.md` để lấy ví dụ Dàn giáo thiết kế.
