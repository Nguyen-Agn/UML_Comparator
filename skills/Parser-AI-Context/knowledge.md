# Kiến Thức Domain Thực Tế 

## 1. Cấu trúc Draw.io
Một file draw.io do hệ thống sinh ra sẽ có đuôi `.drawio` nhưng bên dưới lớp vỏ chính là một định dạng XML của thư viện `mxGraph`.

Có 2 dạng biểu diễn phổ biến mà ứng dụng sẽ gặp:

### Dạng 1: Compressed (Định dạng nén mặc định)
Lý do Draw.io nén vì muốn giảm dung lượng file xuống thấp nhất có thể nếu vẽ hình chằng chịt.
```xml
<mxfile host="app.diagrams.net">
  <diagram id="abc" name="Page-1">
    jZLB... (Một đoạn dài ngoằng được mã hoá Base64)
  </diagram>
</mxfile>
```
**Cơ chế giải nén:** 
Chuỗi `jZLB...` đó đã được xuất xưởng nhờ thuật toán:
`Raw XML string` -> `URL Encode` -> `Deflate (Zlib nén)` -> `Base64 Encode`
=> *Do đó, code Golang nhận vào phải thực hiện ngược lại:*
1. Dùng thư viện `encoding/base64` để Decode Base64 lấy mảng byte tĩnh.
2. Đưa mảng byte qua thư viện `compress/flate` (thuật toán inflate) để bung nén ra.
3. Dùng `net/url` để QueryUnescape (URL Decode).
4. Khúc cuối bắt buộc thu được chuỗi có thẻ `<mxGraphModel>.....`

### Dạng 2: Uncompressed (Plain XML chuẩn)
Draw.io thi thoảng nếu user tắt compress (hoặc xuất Export XML thủ công) thì nó sẽ lưu rõ ràng mọi thẻ:
(Xem cấu trúc mxGraphModel XML chuẩn).

## 2. Cấu trúc Mermaid (.mmd hoặc .mermaid)
Mermaid là một DSL (Domain Specific Language) dạng text thuần.

**Cơ chế xử lý:**
1. Đọc file text theo dòng.
2. **Filtering**: Loại bỏ các dòng trống hoặc các dòng comment bắt đầu bằng `%%`.
3. Trả về nội dung text đã làm sạch và sourceType là `"mermaid"`.

## 3. Strategy Pattern trong Parser
Hệ thống sử dụng `AutoParser` để tự động điều phối file dựa trên extension:
- `.drawio` -> `DrawioParser` (sourceType: "drawio")
- `.solution` -> `SolutionParser` (sourceType: "drawio")
- `.mmd` / `.mermaid` -> `MermaidParser` (sourceType: "mermaid")

Điều này giúp module Builder phía sau không cần quan tâm đến logic đọc file phức tạp, chỉ cần dựa vào `sourceType` để chọn Builder tương ứng.
