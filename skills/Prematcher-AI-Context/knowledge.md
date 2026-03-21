# Knowledge: Prematcher Business Logic & Rules

## 1. Attribute Parsing Rules
Chuỗi thuộc tính thường có định dạng: `[Scope] Name : Type [= DefaultValue]`
- **Scope**: (+: public, -: private, #: protected). Nếu không có, mặc định là public (+).
- **Name**: Chuỗi trước dấu `:`.
- **Type**: Chuỗi sau dấu `:`.

## 2. Method Parsing Rules
Chuỗi phương thức: `[Scope] Name(params) : ReturnType`
- **Scope**: Tương tự Attribute.
- **Inputs**: Cần bóc tách danh sách params bên trong `()`. Mỗi param có dạng `Name : Type`.
- **Output**: Là `ReturnType` sau dấu `:`.
- **Type**: Phân loại theo business logic:
  - `constructor`: Method Name trùng với Class Name, hoặc tên là `constructor`, `init`.
  - `getter`: Method Name bắt đầu bằng `get` (không phân biệt hoa/thường) VÀ có từ 0 đến 1 tham số.
  - `setter`: Method Name bắt đầu bằng `set` (không phân biệt hoa/thường) VÀ có lớn hơn 0 tham số.
  - `custom`: Tất cả các trường hợp còn lại.

## 3. ArchWeight Calculation (Bitwise)
ArchWeight là một số `uint32` dùng để mô tả nhanh đặc điểm của một Class. 
- Bit 29-31: Loại Class (3 bit - 0: Unknown, 1: Class, 2: Interface, 3: Abstract, 4: Enum) [Quan trọng nhất]
- Bit 28: Có thừa kế không? (1 bit - 1: Có, 0: Không)
- Bit 24-27: Số lượng Interface thực thi (4 bit - Max 15)
- Bit 18-23: Số lượng Method (6 bit - Max 63). **Lưu ý**: Không tính đếm các method có Type là `getter` hoặc `setter`.
- Bit 13-17: Số lượng Attribute (5 bit - Max 31)
- Bit 9-12: Số lượng Class liên quan/phụ thuộc (4 bit - Max 15)
- Bit 6-8: Số lượng Custom Type <T> (3 bit - Max 7)
- Bit 2-5: Số lượng Static members (4 bit - Max 15)
- Bit 0-1: Dự phòng (Not used)

## 4. Relationship Mapping
- **Inheritance (Extends)**: Nếu một Edge có style `endArrow=block` (mũi tên rỗng), node Source `Inherits` từ Target.
- **Realization (Implements)**: Nếu Edge có style `dashed=1` và `endArrow=block`, node Source `Implements` Target.

## 5. Text Normalization
- Mọi chuỗi text cần xóa bỏ các thẻ HTML dư thừa (nếu module trước chưa xử lý hết).
- Trim khoảng trắng ở hai đầu.
