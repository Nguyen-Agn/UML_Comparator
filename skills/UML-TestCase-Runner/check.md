# Verification Checklist — UML TestCase Runner

> **Verification Layer** — dùng checklist này sau khi chạy test để xác nhận
> kết quả đúng và không có regression.

---

## Trước khi chạy test

- [ ] Tất cả 9 file trong `UMLs_testcase/incorrect/` đã tồn tại (kiểm tra bằng `list_dir`)
- [ ] `builder/incorrect_uml_test.go` đã được tạo và compile được (`go build ./builder/...`)
- [ ] `domain.ValidateGraph()` có đủ tất cả error codes trong `knowledge.md`

---

## Sau khi chạy `go test ./builder/... -v -run TestIncorrect`

### 1. Pass/Fail Count
- [ ] Không có test nào FAIL (`--- FAIL:` không xuất hiện trong output)
- [ ] Số lượng test PASS = 9 (một per file incorrect)

### 2. Coverage ERRORs
- [ ] `TestIncorrect_EmptyGraph` → log chứa `EMPTY_GRAPH`
- [ ] `TestIncorrect_EmptyNodeName` → log chứa `EMPTY_NODE_NAME`
- [ ] `TestIncorrect_InvalidNodeType` → log chứa `INVALID_NODE_TYPE` hoặc `EMPTY_GRAPH`
- [ ] `TestIncorrect_DanglingEdgeSource` → log chứa `DANGLING_EDGE_SOURCE`
- [ ] `TestIncorrect_SelfLoop` → log chứa `SELF_LOOP_EDGE`

### 3. Coverage WARNs
- [ ] `TestIncorrect_SuspectName` → log chứa `SUSPECT_NODE_NAME` + `FilterErrors` trả 0
- [ ] `TestIncorrect_TrivialName` → log chứa `TRIVIAL_NODE_NAME` + `FilterErrors` trả 0
- [ ] `TestIncorrect_IncompleteAttribute` → log chứa `INCOMPLETE_ATTRIBUTE` + `FilterErrors` trả 0
- [ ] `TestIncorrect_IncompleteMethod` → log chứa `INCOMPLETE_METHOD` + `FilterErrors` trả 0

### 4. Pipeline Behavior
- [ ] ERROR test cases: `len(domain.FilterErrors(errs)) > 0` → confirmed pipeline would stop
- [ ] WARN test cases: `len(domain.FilterErrors(errs)) == 0` → confirmed pipeline would continue

---

## Kiểm Tra Chất Lượng Test

- [ ] Mỗi test chỉ trigger **đúng error code mục tiêu** (không có false-positives từ file mẫu)
- [ ] WARN test files không trigger bất kỳ ERROR nào không mong muốn
- [ ] Log message mỗi test mô tả rõ: file name, actual codes, expected codes

---

## Khi Test FAIL — Quy Trình Debug

1. Đọc log `✘ Expected error code "XYZ" not found. Got: [...]`
2. Mở file `.drawio` tương ứng → kiểm tra cấu trúc XML có thật sự trigger điều kiện không
3. Chạy `go test -v -run TestIncorrect_XYZ` để xem chi tiết
4. Nếu file drawio sai → sửa file
5. Nếu `ValidateGraph()` bị logic lỗi → ghi vào `recorrect.md` của skill liên quan
6. Nếu Builder không tạo đúng node → xem `Skills/Builder-AI-Context/`

---

## Integrity Codes Quick Reference

| Code | Severity | Trigger |
|---|---|---|
| `EMPTY_GRAPH` | ERROR | 0 nodes sau Build |
| `EMPTY_NODE_NAME` | ERROR | node.Name == "" |
| `INVALID_NODE_TYPE` | ERROR | Type ∉ {Class, Interface, Abstract, Actor, Enum} |
| `DANGLING_EDGE_SOURCE` | ERROR | edge.SourceID không có trong nodes |
| `DANGLING_EDGE_TARGET` | ERROR | edge.TargetID không có trong nodes |
| `SELF_LOOP_EDGE` | ERROR | edge.SourceID == edge.TargetID |
| `SUSPECT_NODE_NAME` | WARN | name chứa `<`, `>`, `<<`, `>>`, `&` |
| `TRIVIAL_NODE_NAME` | WARN | len(rune(name)) ≤ 2 |
| `INCOMPLETE_ATTRIBUTE` | WARN | attr không chứa `:` |
| `INCOMPLETE_METHOD` | WARN | method TrimSpace kết thúc bằng `(` |
