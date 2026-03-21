# SKILL: uml-comparator
## Purpose
So sánh 2 UMLGraph đã được Match chẻ sát với nhau. Tìm ra sự khác lạ cục bộ (sai thuộc tính, thiếu class, sai chiều mũi tên kế thừa). Trả ra DiffReport báo lỗi chi tiết.

## Execution Approach
1. **TypeMap Registration**: Dựa vào `MappingTable`, tạo `TypeMap[SolutionNodeName] = StudentNodeName` để phục vụ dịch đổi ngữ nghĩa các kiểu tham số hoặc thuộc tính tự định nghĩa.
2. **Missing Node Check**: Duyệt List Nodes của Mẫu. Nếu không có trong `MappingTable` -> Báo MissingNode.
3. **Class Comparison**: Khi tìm được class SinhVien tương ứng `MappedNode`:
   - So sánh Constructor: Nếu mẫu không có thì auto qua, nếu có thì so trùng kiểu params (nhưng KHÔNG cần đúng thứ tự) và Scope.
   - So sánh Method Set: Getter/Setter chỉ đếm tổng số lượng (báo cảnh báo lệch count nếu có). Method thường thì Tên similarity >= 0.5; Kiểu trả về khớp tuyệt đối; Params không lo tên nhưng khớp tuyệt đối thứ tự và Kiểu (sau TypeMap); Scope khớp tuyệt đối.
   - So sánh Attribute Set: Kiểu dữ liệu khớp tuyệt đối (sau TypeMap); Tên similarity >= 0.5 HOẶC chuỗi này contains chuỗi kia; Scope khớp tuyệt đối.
4. **Edge Comparison**: Duyệt Edges. Tìm Edge tương ứng qua `MappingTable`. Dò Mũi tên ngược (Reverse Arrow) hoặc Missing Edge.
5. Ghi nhận TẤT CẢ sai lệch vào list báo cáo chi tiết (`DetailedErrors`).
