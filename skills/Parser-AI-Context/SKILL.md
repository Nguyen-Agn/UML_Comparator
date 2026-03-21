# SKILL: drawio-xml-parser

## Purpose (Mục đích)
Phân tích file `.drawio` hoặc xml của sinh viên và đáp án, biến chúng từ dạng nén/mã hoá (hoặc raw) về dạng chuỗi `RawXMLData` (thuần text XML) một cách chính xác.

## Use When (Khi nào dùng skill này)
- Vừa mới bắt đầu pipeline so sánh UML.
- Cần lập trình module con `IFileParser` bằng Golang.
- Cần cung cấp dữ liệu sạch đầu vào (XML thô) cho cụm `IModelBuilder`.

## Required Inputs
- Đường dẫn file vật lý trên ổ cứng (`filePath` kiểu String).

## Expected Output
- Dữ liệu `domain.RawXMLData` chứa cấu trúc XML chuẩn của thẻ `<mxGraphModel>`.
- Thông báo lỗi `error` (nếu hỏng file, sai định dạng).

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
