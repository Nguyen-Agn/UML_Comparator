# SKILL: drawio-xml-parser

## Purpose (Mục đích)
Phân tích file `.drawio`, `.solution` hoặc `.mmd` (Mermaid), lọc bỏ các thông tin nhiễu (comments, metadata) và trả về dữ liệu `RawModelData` sạch cùng source type tương ứng.

## Use When (Khi nào dùng skill này)
- Bắt đầu pipeline so sánh UML từ nhiều nguồn khác nhau.
- Cần lập trình module con `IFileParser` (Strategy Pattern).
- Cần cung cấp dữ liệu sạch đầu vào và định danh loại nguồn cho `IModelBuilder`.

## Required Inputs
- Đường dẫn file vật lý trên ổ cứng (`filePath` kiểu String).

## Expected Output
- Dữ liệu `domain.RawModelData` đã được lọc (clean).
- Chuỗi `sourceType` (ví dụ: "drawio", "mermaid") để chọn Builder phù hợp.
- Thông báo lỗi `error`.

## Execution Approach (Hướng thực thi)
1. Đọc nội dung file ra bộ nhớ (String/Byte array).
2. Kiểm tra xem file chứa XML nén `<diagram>jZLB...</diagram>` hay XML thường `<mxGraphModel>...`.
3. Nếu có mã hóa nén: Thực thi thuật toán giải mã tuần tự: `Base64 Decode` -> `Decompress (Flate/Zlib)` -> `URL Decode` -> Thu được cấu trúc chuẩn.
4. Nếu file không bị nén: Móc thẳng lấy node XML trả về.

## Quality Criteria
- Code sạch sẽ, chạy tốc độ cao, dùng các package chuẩn của Golang (`encoding/base64`, `compress/flate`, `net/url`).
- Không để rò rỉ bộ nhớ (memory leak) do IO đọc file.

## References
Cần khai thác các tệp sau trong cùng thư mục để thực hiện chuẩn:
- `knowledge.md`: Đặc tả 2 dạng cấu trúc file Draw.io.
- `template.md`: Khung code mẫu bắt buộc để lắp ráp.
- `check.md`: Tiêu chí nghiệm thu nội bộ.
- `recorrect.md`: Nhật ký lỗi sẽ ghi vào khi có exception.
