---
name: uml-testcase-runner
description: Chạy test case trên các file UML draw.io để kiểm tra xem Builder + ValidateGraph có phát hiện đúng lỗi. Dùng để verify rằng pipeline xử lý UML bắt được các loại lỗi sai trong file .drawio.
---

# SKILL: uml-testcase-runner

## Purpose (Mục đích)
Skill này nhận vào một file `.drawio` bị lỗi có chủ đích, chạy qua pipeline
`Builder → ValidateGraph`, sau đó **so sánh kết quả thực tế với expected
errors**. Mục tiêu: đảm bảo hệ thống phát hiện đúng lỗi UML, không bị
false-negative (lỗi có mà không phát hiện) hay false-positive (báo lỗi sai).

Skill này áp dụng tư duy **Verification Layer** từ Strategy: "Làm sao biết
thứ vừa tạo ra là đúng, dùng được, và đáng tin?"

## Use When
- Thêm file `.drawio` mới vào `UMLs_testcase/incorrect/` → cần verify hệ thống bắt đúng lỗi.
- Sửa `ValidateGraph()` hoặc Builder → cần chạy lại toàn bộ test incorrect để đảm bảo không có regression.
- Viết test case mới cho error code chưa được cover.
- Debug: file UML sinh viên nộp không bị báo lỗi nhưng nghi ngờ có vấn đề.

## Required Inputs
- **File path**: đường dẫn tới file `.drawio` cần test
- **Expected error codes**: danh sách `IntegrityError.Code` kỳ vọng (xem `knowledge.md`)
- **Expected severity**: `ERROR` (pipeline stop) hoặc `WARN` (có thể tiếp tục)

## Expected Output
```
TestCase: <filename>
  ✔ Got error code: <EXPECTED_CODE> [<SEVERITY>]
  ✔ Pipeline behavior: <STOP|CONTINUE> (correct)
  --- OR ---
  ✘ MISSING expected code: <EXPECTED_CODE>
  ✘ UNEXPECTED code found: <ACTUAL_CODE>
```

## Execution Approach
1. Đọc file `.drawio` từ `UMLs_testcase/incorrect/<filename>`
2. Chạy `builder.NewStandardModelBuilder().Build(rawXML)`
3. Chạy `domain.ValidateGraph(graph, label)`
4. So sánh actual codes với expected codes từ `testcases.md`
5. Assert: mỗi expected code phải xuất hiện trong actual results
6. Log PASS/FAIL theo từng code

## Quality Criteria
- **Mỗi Error Code phải có ít nhất 1 test case** trong `UMLs_testcase/incorrect/`
- Test case phải **chỉ trigger đúng error code mục tiêu**, không trigger thêm ERROR không mong muốn
- WARN test cases có thể build thành công (pipeline không dừng)
- ERROR test cases phải khiến pipeline stop trước khi vào Matcher

## Edge Cases
- File `.drawio` hoàn toàn rỗng (0 bytes) → Parser trả error trước khi Builder chạy
- File XML hợp lệ nhưng không có thẻ `<mxGraphModel>` → Builder trả error
- Node có cả `EMPTY_NODE_NAME` lẫn `INVALID_NODE_TYPE` cùng lúc → Cả hai phải được báo

## References
- `knowledge.md`: bảng tất cả error codes và cách trigger
- `testcases.md`: bảng test cases cụ thể + expected results
- `check.md`: checklist tự verification
- `../UMLs_testcase/incorrect/`: thư mục chứa các file .drawio lỗi
- `../../domain/validator.go`: implementation của ValidateGraph()
- `../../builder/incorrect_uml_test.go`: Go test file tự động

## Changelog
- v1.0 (2026-03-20): Khởi tạo skill, phủ 9 error/warn codes từ ValidateGraph
