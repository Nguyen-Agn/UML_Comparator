# Mẫu Thực Thi Bắt Buộc (Execution Template)

1. **Interface Contract:**
package builder
var _ IModelBuilder = (*StandardModelBuilder)(nil)

2. **Quy tắc Single Responsibility:**
Hàm Build() không được phép dài ngoằng xử lý cả Parse XML, cả móc Regex HTML. Bắt buộc tách hàm:
func (b *StandardModelBuilder) parseHTMLToAttributes(html string) []string {
   // Regex logic here
}
