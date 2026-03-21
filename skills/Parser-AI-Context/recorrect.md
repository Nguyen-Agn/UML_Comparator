# Evolution Log (Nhật ký Cập Nhật Skill & Fix Lỗi)

*Đây là bộ não phụ trợ tiến hóa (Evolution Layer). Nếu Agent (AI) gục ngã vì gặp ca khó ở Module Parser, Agent BẮT BUỘC tự giác cập nhật nguyên nhân vào đây rổi mới đi ngủ.*

## Danh sách Lỗi Gặp Phải & Bài Học:
### [2026-03-20] Run #1 — PASS ✔
- **Kết quả:** 4/4 Unit Test PASSED ngay lần chạy đầu tiên.
- **Các trường hợp đã kiểm tra:** Plain XML parse, Empty path guard, NonExistent file, Edge count validation.
- **Ghi chú:** Cần bổ sung testdata file dạng *Compressed* (Base64+Deflate) để kiểm tra nhánh `decodeBase64Deflate` trong các lần chạy sau.

*(Cập nhật khi có lỗi mới)*.

---
### QUY TẮC KHAI BÁO LOG (Dành cho Agent)
Mỗi sự cố khi fix xong, hãy bê nguyên form sau ghi xuống cuối danh sách:

- **Mã Lỗi / Tín hiệu Error:** (VD: Lỗi `illegal base64 data at input byte` khi đang chạy decode file drawio).
- **Nguyên nhân gốc rễ (Root cause):** (VD: Thẻ `<diagram>` của file do regex móc ra trót bị dính kèm dấu xuồng dòng `\n` hoặc khoảng trắng ở hai đầu).
- **Đòn chốt Fix (Fix strategy):** (VD: Thêm 1 dòng lệnh `payload = strings.TrimSpace(payload)` trước khi truyền thẳng vào hàm `base64.StdEncoding.DecodeString`).
