# Verification Checklist — Data Flow Integrity
> Chạy `domain.ValidateGraph(*UMLGraph, label)` trước khi tiếp tục sang Matcher/Comparator.

---

## Trước khi chạy Matcher / Comparator, phải pass ĐỦ các mục sau:

### 1. Node Integrity
- [ ] `len(graph.Nodes) > 0` với CẢ 2 graph (mẫu và sinh viên)
- [ ] Không có `UMLNode.Name == ""` trong bất kỳ node nào
- [ ] Mọi `UMLNode.Type` thuộc `{"Class", "Interface", "Actor"}` — không có giá trị lạ
- [ ] Không có node nào có Name chứa `"\n"` hoặc `"<<"` (= chưa được clean đúng)

### 2. Edge Integrity
- [ ] Mọi `edge.SourceID` tồn tại trong danh sách `node.ID` của graph đó
- [ ] Mọi `edge.TargetID` tồn tại trong danh sách `node.ID` của graph đó
- [ ] Không có `edge.SourceID == edge.TargetID` (self-loop)

### 3. Attribute / Method Quality
- [ ] Không có Attribute nào chỉ là `:` , `)`, `,` hoặc bắt đầu bằng `:` (= fragment từ multi-line method)
- [ ] Không có Method nào chứa unmatched `(` (= method bị cắt giữa chừng)

### 4. Swimlane Structure (Draw.io Specific)
- [ ] Số Node sau Build ≈ số container swimlane trong file (kiểm tra bằng mắt hoặc đếm cell với `parent=rootLayerID`)
- [ ] Mỗi Node phải có ít nhất một trong: `Attributes` hoặc `Methods` có nội dung thực chất

---

**Hành động khi fail:** `ValidateGraph()` trả `[]IntegrityError`. Log prefix `[DATA_INTEGRITY_ERROR]` và **DỪNG pipeline**, không tiếp tục sang Comparator.

**Hành động khi file rỗng / 0 bytes:** Parser trả error → Báo `0 điểm` ngay lập tức.

---

## Integrity Codes

| Code | Ý nghĩa |
|---|---|
| `EMPTY_GRAPH` | Graph có 0 Node |
| `EMPTY_NODE_NAME` | Node có Name rỗng (swimlane bug) |
| `INVALID_NODE_TYPE` | Type không thuộc Class/Interface/Actor |
| `DANGLING_EDGE_SOURCE` | Edge.SourceID không khớp bất kỳ Node nào |
| `DANGLING_EDGE_TARGET` | Edge.TargetID không khớp bất kỳ Node nào |
| `SELF_LOOP_EDGE` | Edge trỏ vào chính nó |
