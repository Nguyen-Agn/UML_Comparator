---
name: "SOLID & Unit Test Refactoring"
description: "Skill để tái cấu trúc (refactor) một module bất kỳ theo chuẩn SOLID, áp dụng Interface-first và bổ sung Unit Test hoàn chỉnh."
---

# SKILL: Refactoring Module theo chuẩn SOLID & Unit Test

## 1. Mô Tả (Intent)
Skill này cung cấp định hướng và quy trình bắt buộc để Agent tái cấu trúc lại một module hoặc tính năng bất kỳ có sẵn trong dự án. Trọng tâm của việc refactoring là đảm bảo mã nguồn tuân thủ nguyên tắc SOLID, đặc biệt là Open/Closed Principle, sử dụng Interface-first design, và đảm bảo mọi logic đều được bao phủ bởi Unit Test.
Mục tiêu là làm cho mã nguồn linh hoạt hơn, dễ mở rộng, dễ maintain và tránh sửa đổi chéo giữa các module không liên quan.

## 2. Đầu Vào (Inputs)
- Mã nguồn hiện tại của module cần refactor.
- Các yêu cầu mới (nếu có) đòi hỏi việc refactor.
- Mô tả các giới hạn, sự ràng buộc của hệ thống đối với module này.

## 3. Đầu Ra Kỳ Vọng (Expected Outputs)
- Mã nguồn mới đã được tách ghép theo các nguyên tắc SOLID.
- Interface rõ ràng thống nhất cho các thành phần, với comment giải thích rõ ràng chức năng.
- Các file `*_test.go` đầy đủ, bao phủ (coverage) các ca kiểm thử chính (happy path, edge cases).
- Mã nguồn sau khi refactor không làm phá vỡ logic cũ hoặc module khác.

## 4. Chiến Lược Thực Thi Tổng Quát
1. **Phân tích (Analyze):** Nhận dạng module hiện tại, đánh giá vi phạm nguyên lý SOLID (ví dụ: file có quá nhiều trách nhiệm, chứa logic hard-code, dính chặt với concrete class).
2. **Thiết kế Interface (Interface-first):** Định nghĩa Interface mô tả hành vi, với comment cụ thể cho từng method. Bắt buộc để cho các module giao tiếp thông qua Interface chứ không giao tiếp trực tiếp qua Implementation (Class).
3. **Phân rã (Decouple):** Tách Implementations ra khỏi nhau để thỏa mãn Single Responsibility (SRP) và Open/Closed (OCP).
4. **Tích hợp Unit Test:** Viết các Unit Test dễ dàng nhờ vào các Interface vừa được thiết kế; áp dụng Table-Driven Test.
5. **Kiểm chứng (Verify):** Xác nhận code chạy đúng, Pass test và tự review qua `check.md`.

Chuyển sang bước tiếp theo: Đọc `knowledge.md` để nắm bắt luật chơi và rule bắt buộc khi thực hiện.
