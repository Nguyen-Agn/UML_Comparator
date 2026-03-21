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
```xml
<mxfile>
  <diagram id="abc" name="Page-1">
    <mxGraphModel dx="1290" dy="687" grid="1" gridSize="10">
      <root>
        <mxCell id="0" />
        <mxCell id="1" parent="0" />
        <mxCell id="2" value="ClassTaikhoan" style="shape=umlClass" vertex="1" parent="1">
            <mxGeometry x="100" y="100" width="120" height="80" as="geometry" />
        </mxCell>
      </root>
    </mxGraphModel>
  </diagram>
</mxfile>
```
**Cơ chế giải nén:** 
Không cần phải Encode/Decode rườm rà. Code Golang chỉ việc bóc lấy khối text nằm bên trong thẻ `<diagram>...</diagram>` bằng Regex hoặc Simple XML Parser. Mảng block text đó chính là Data đích trả về.
