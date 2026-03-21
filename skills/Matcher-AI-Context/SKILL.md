# SKILL: entity-matcher

## Purpose
So khớp các Entity (Class/Interface/Enum) giữa `ProcessedUMLGraph` của Đáp án và của Sinh viên để tạo ra `MappingTable`. Nhờ có dữ liệu đã tiền xử lý, Matcher sử dụng cấu trúc `ArchWeight` kết hợp với Fuzzy Text Matching để đưa ra các quyết định nối (map) chính xác nhất, ngay cả khi tên bị sai chính tả hoặc gõ tắt.

## Use When
- Sau khi `Prematcher` đã sinh ra 2 cây `ProcessedUMLGraph` với `ArchWeight` đầy đủ.
- Cần dóng hàng các Node trước khi đưa qua `Comparator` chấm điểm chi tiết.

## Required Inputs
- `solution *domain.ProcessedUMLGraph`: Graph mẫu (read-only).
- `student *domain.ProcessedUMLGraph`: Graph sinh viên (read-only).

## Expected Output
- `domain.MappingTable`: Dictionary Map từ `Solution Node ID -> Student Node ID`.

## Execution Approach
Thuật toán dựa trên 2 trụ cột: **Kiến trúc (ArchWeight)** và **Văn bản (FuzzyMatch)**.
1. **FuzzyMatch Submodule**: Việc so sánh Text phải được decouple ra một Interface riêng (`IFuzzyMatcher`).
2. **Architecture First (Unpack & Tolerance)**: Thay vì trừ thẳng số nguyên, `ArchWeight` được bung ra thành `ArchTraits`. Các class được xem là "Kiến trúc tương tự" nếu giống hệt các đặc tính cấu trúc lõi (Loại class, Kế thừa, Intf, CustomType) và lệch không quá 15% số lượng (Method, Attribute, Rels, Static).
3. **FuzzyScore làm Tie-breaker**: Gom tất cả các node đạt chuẩn "Kiến trúc tương tự" lên top đầu, rồi xếp hạng bên trong bằng độ giống tên `FuzzyScore`. Node có tên khớp nhất nhưng chung kiến trúc sẽ được chọn.

## Quality Criteria
- Module `matcher` tuyệt đối không thay đổi data bên trong input pointers.
- Kết quả trả về không bao giờ có chuyện 1 node Solution trỏ tới 2 node Student (phải map 1-1).
