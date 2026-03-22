# SKILL: uml-pre-matcher

## Purpose
Biến đổi mô hình `domain.UMLGraph` (dạng chuỗi thô) thành `domain.ProcessedUMLGraph` (dạng cấu trúc chi tiết). 
Đây là bước đệm quan trọng để module Matcher có thể so khớp chính xác dựa trên thuộc tính, phương thức và trọng số kiến trúc (ArchWeight).

## Use When
- Khi Builder đã tạo xong `UMLGraph`.
- Khi cần phân tích sâu cấu trúc của từng Class/Interface (tách biệt Name, Scope, Type của Attribute/Method).
- Khi cần tính toán trọng số thiết kế (ArchWeight) để hỗ trợ so khớp nâng cao.

## Required Inputs
- `*domain.UMLGraph`: Đồ thị UML thô thu được từ module Builder.

## Expected Output
- `*domain.ProcessedUMLGraph`: Đồ thị đã qua tiền xử lý, các chuỗi text được bóc tách thành struct. Mỗi node có `ArchWeight` (bitmask kiến trúc) và `Shortcut` (bitmask hướng dẫn: Bit 0=getters, Bit 1=setters).

## Execution Approach
1. **Node Processing**: Duyệt qua danh sách Nodes trong `UMLGraph`.
2. **Member Parsing**: 
   - Parse `Attributes` trước để lấy danh sách thuộc tính cơ sở.
   - Parse `Methods` sau. Với các method `get/set`, thực hiện so khớp fuzzy (>= 80%, không phân biệt hoa thường) với danh sách thuộc tính của node đó để định nghĩa Type là `getter` hoặc `setter`.
   - Đảm bảo mỗi thuộc tính chỉ gắn với tối đa một bộ get/set (one-to-one mapping).
3. **Weight Calculation**: Tính toán `ArchWeight` dựa trên các đặc điểm nhận diện (VD: Singleton, Interface implementation, etc.).
4. **Relationship Resolution**: Thiết lập các trường `Inherits` và `Implements` dựa trên danh sách Edges.

## Quality Criteria
- Phải bóc tách đúng Scope (+, -, #) của thuộc tính và phương thức.
- Tên thuộc tính/phương thức phải được làm sạch (trim space, remove HTML if any).
- Trọng số `ArchWeight` phải được tính toán nhất quán theo bitmask (nếu có).

## Verification
- Kiểm tra xem số lượng Node/Edge có được giữ nguyên không.
- Kiểm tra xem các thuộc tính phức tạp (VD: `- name : String = "Default"`) có được parse đúng Name và Type không.

## Changelog
- v1.0: Khởi tạo skill cho Prematcher module.
