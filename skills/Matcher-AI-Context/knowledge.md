# Knowledge: Matcher Logic & Similarity Algorithms

## 1. Dữ liệu Đầu vào & Tính Bất Biến (Immutability)
- Đầu vào là `*domain.ProcessedUMLGraph`. Các node bên trong có sẵn `ArchWeight` (uint32).
- Mọi thao tác đều chỉ "đọc".

## 2. Chiến lược ArchWeight-First (Unpack & Tolerance)
Vì `ArchWeight` là một bitfield được nén, việc trừ thẳng hai số `uint32` sẽ không chính xác (lệch ở các bit cao như Class Type sẽ tạo ra delta khổng lồ, làm mất tính cân bằng).
Do đó, khi cầm 1 node `Sol_A` đối chiếu với `Stu_X`, ta Unpack ArchWeight ra struct `ArchTraits` và kiểm tra **Kiến trúc tương tự (Similar)**:
- **Chính xác tuyệt đối**: Loại Class (`ClassType`), Kế thừa (`HasInheritance`), Số interface triển khai (`NumInterfaces`), Số custom type (`NumCustomTypes`). Các trường này bắt buộc phải giống nhau 100%.
- **Chênh lệch <= 15%**: Số Method, Số Attribute, Số Class phụ thuộc, Số Static member. Cho phép hao hụt không quá 15% (Làm tròn lên với `math.Ceil`). Nếu lệch <= 15%, vẫn coi là **giống nhau**.

Tất cả các node chênh lệch nằm trong hạn mức sẽ được hệ thống xem là "Kiến trúc tương tự" và nhóm chung lại với mức ưu tiên cao nhất, cào bằng thứ hạng kiến trúc.

## 3. IFuzzyMatcher Interface
Fuzzy Matching (ví dụ thuật toán Levenshtein) phải được tách biệt thành một submodule.
```go
type IFuzzyMatcher interface {
	// Trả về độ tương đồng (0.0 đến 1.0)
	Compare(s1, s2 string) float64
}
```
Luôn `strings.ToLower` và gỡ dấu cách (trim space) trước khi ném vào submodule này.

## 4. Pipeline Duyệt Node (Node Mapping Strategy)
Thay vì chia làm nhiều vòng đấu (passes) phức tạp, thuật toán gom lại thành 1 vòng lọc thông minh nhất:
1. Với mỗi Node Solution đang xét, tính toán trước `IsArchitectureSimilar` (true/false) và `FuzzyScore` (Độ giống tên do IFuzzyMatcher trả về - tối đa 1.0) đối với TOÀN BỘ Node sinh viên chưa map.
2. Sort danh sách ứng viên (Candidate List) theo 3 bậc (tiêu chí trên cùng quan trọng nhất):
   - **Bậc 1 (Kiến trúc)**: Ưu tiên các node đạt chuẩn `IsArchitectureSimilar == true` lên trên cùng.
   - **Bậc 2 (Văn bản)**: Trong cùng một nhóm kiến trúc (Ví dụ các node đều passing IsArchitectureSimilar), sắp xếp theo độ giống tên (`FuzzyScore`) giảm dần. Điều này đảm bảo node có kiến trúc hợp lệ và tên chính xác nhất luôn là top 1.
   - **Bậc 3 (Delta Fallback)**: Nếu rơi rớt xuống nhóm không passing Architecture, thì sắp xếp bằng theo tổng delta các thành phần cấu trúc.
3. Chốt hạ: Duyệt List đã được Sort, chọn node ĐẦU TIÊN có `FuzzyScore >= threshold`. Nếu quét hết pool mà vẫn không có node nào qua điểm chuẩn Fuzzy, xem như sinh viên làm thiếu node.
