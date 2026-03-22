# SKILL: uml-comparator
## Purpose
So sánh 2 UMLGraph đã được Match chẻ sát với nhau. Tìm ra sự khác lạ cục bộ (sai thuộc tính, thiếu class, sai chiều mũi tên kế thừa). Trả ra `*DiffReport` chứa các con trỏ trực tiếp đến đối tượng bị lỗi.

## Execution Approach
1. **TypeMap Registration**: Dựa vào `MappingTable`, tạo `TypeMap[SolutionNodeName] = StudentNodeName` để phục vụ dịch đổi ngữ nghĩa các kiểu tham số hoặc thuộc tính tự định nghĩa.
2. **Missing Node Check**: Duyệt List Nodes của Mẫu. Nếu không có trong `MappingTable` -> Báo `MissingDetail.Class` (Stu = nil).
3. **Class Comparison**: Khi tìm được class SinhVien tương ứng `MappedNode`:
   - So sánh Constructor: Nếu mẫu không có thì auto qua, nếu có thì so trùng kiểu params (nhưng KHÔNG cần đúng thứ tự) và Scope.
   - So sánh Attribute: Match bằng **Type** (sau TypeMap) rồi **Name** (fuzzy), sau đó kiểm tra Scope và Kind -> `WrongDetail/CorrectDetail`.
   - So sánh Method Set: Match bằng **ReturnType + ParamCount (+-1 nếu cả hai >=2)** rồi **Name** (fuzzy), sau đó kiểm tra Scope, Kind, Params chính xác -> `WrongDetail/CorrectDetail`. Getter/Setter chỉ đếm số lượng.
4. **Edge Comparison**: Duyệt Edges. Tìm Edge tương ứng qua `MappingTable`. Dò: Khớp hoàn toàn -> `CorrectDetail`; Sai loại (wrong type) -> `WrongDetail`; Mũi tên ngược (Reverse Arrow) -> `WrongDetail`; Không tìm thấy -> `MissingDetail`. Student có thừa -> `ExtraDetail`.
5. **Output**: Trả về `*domain.DiffReport` gồm các danh sách: `NodeDiff`, `AttributeDiff`, `MethodDiff`, `EdgeDiff`. 
   - Mỗi struct Diff chứa 3 trường chính: `Sol` (con trỏ đến đối tượng mẫu), `Stu` (con trỏ đến đối tượng sinh viên), và `Description`.
   - Giúp Grader có toàn bộ dữ liệu (ArchWeight, Type, Params...) để chấm điểm mà không cần parse string hay lookup lại.
