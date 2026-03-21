# Tiêu Chuẩn Thực Thi (Execution Template)

Đây là những Pattern và nguyên tắc BẮT BUỘC Agent phải tuân theo khi triển khai code cho Module Parser nhằm đảm bảo tính đồng nhất (SOLID).

## 1. Interface Implementation Pattern
Go không dùng từ khóa `implements`. Để đảm bảo struct implement đúng interface từ lúc compile, bắt buộc chèn dòng `var _ Interface = (*Struct)(nil)` ở cấp toàn cục.

```go
package parser

import (
	"encoding/xml"
	"uml_compare/domain"
)

type DrawioParser struct {
    // Không chứa state (như file handle) để tuân thủ thiết kế Stateless Object
}

// Compile-time interface check
var _ IFileParser = (*DrawioParser)(nil)

// Constructor Pattern chuẩn
func NewDrawioParser() IFileParser {
	return &DrawioParser{}
}

func (p *DrawioParser) Parse(filePath string) (domain.RawXMLData, error) {
	// ... logic xử lý chính
	return "", nil
}
```

## 2. SRP Helper Function Pattern (Quy tắc Trách nhiệm Đơn)
Hàm `Parse` chỉ được đứng ra làm "nhạc trưởng" (hàm điều phối). Các luồng giải mã (Base64) hóc búa cần được tách riêng thành các hàm private (`小寫` method) độc lập.

**Ví dụ Code TỐT:**
```go
func (p *DrawioParser) Parse(filePath string) (domain.RawXMLData, error) {
    content, err := p.readFileContent(filePath)
    if err != nil { ... }
    
    if p.isCompressed(content) {
        return p.decodeBase64Deflate(content) // Trỏ tới hàm con xử lý việc nén
    }
    return p.extractRawXML(content), nil
}
```

**Ví dụ Code XẤU (Nghiêm cấm làm):** 
Viết tuốt luốt logic 30-40 dòng mã hoá Base64, rồi bắt lỗi Zlib Inflate vào tận cùng một block thân bên trong hàm `Parse()`. Khó đọc và đau mắt.

## 3. Quy tắc Error Wrapping (Kẹp lỗi)
Bắt buộc dùng `fmt.Errorf("context message: %w", err)` thay vì chỉ trả về `err` nắn trơn, để biết chính xác lỗi xảy ra ở khâu nào (đọc file, decode, hay unzip).
