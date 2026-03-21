# Test Cases — UML Incorrect Files

> **Execution Layer** — bảng test cases chi tiết cho từng file trong
> `UMLs_testcase/incorrect/`. Mỗi hàng = 1 test case với expected results
> rõ ràng, bao gồm: expected error codes, severity, và behavior pipeline.

---

## Quy Ước

| Ký hiệu | Ý nghĩa |
|---|---|
| ✅ MUST have | Test phải assert code này xuất hiện → FAIL nếu không có |
| ⚠️ MUST NOT have unexpected ERROR | WARN ok, nhưng không được có ERROR ngoài danh sách |
| 🛑 Pipeline STOP | `FilterErrors()` trả > 0 item → không được tiếp tục Matcher |
| 🟢 Pipeline CONTINUE | `FilterErrors()` trả 0 item → có thể tiếp tục Matcher |

---

## Bảng Test Cases

### TC-01: err_empty_graph.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/err_empty_graph.drawio` |
| **Mô tả** | File XML hợp lệ nhưng `<root>` không chứa class/node nào |
| **Expected Codes** | ✅ `EMPTY_GRAPH` (ERROR) |
| **Pipeline** | 🛑 STOP — không được tiếp tục |
| **Go Test** | `TestIncorrect_EmptyGraph` |
| **Ghi chú** | ValidateGraph trả about luôn sau EMPTY_GRAPH, không check thêm |

---

### TC-02: err_empty_node_name.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/err_empty_node_name.drawio` |
| **Mô tả** | Swimlane container có `value=""` — tên class rỗng hoàn toàn |
| **Expected Codes** | ✅ `EMPTY_NODE_NAME` (ERROR) |
| **Pipeline** | 🛑 STOP |
| **Go Test** | `TestIncorrect_EmptyNodeName` |
| **Ghi chú** | Node vẫn được Builder tạo ra (count > 0) nhưng Name = "" |

---

### TC-03: err_invalid_node_type.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/err_invalid_node_type.drawio` |
| **Mô tả** | Node dùng `style="rounded=1"` thay vì `swimlane` hoặc `umlClass` |
| **Expected Codes** | ✅ `INVALID_NODE_TYPE` (ERROR) |
| **Pipeline** | 🛑 STOP |
| **Go Test** | `TestIncorrect_InvalidNodeType` |
| **Ghi chú** | Builder có thể không nhận ra node → Type="" hoặc node không được tạo (EMPTY_GRAPH). Test assert ít nhất 1 error trong {INVALID_NODE_TYPE, EMPTY_GRAPH} |

---

### TC-04: err_dangling_edge.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/err_dangling_edge.drawio` |
| **Mô tả** | Edge có `source="999"` — ID không tồn tại trong graph |
| **Expected Codes** | ✅ `DANGLING_EDGE_SOURCE` (ERROR) |
| **Pipeline** | 🛑 STOP |
| **Go Test** | `TestIncorrect_DanglingEdgeSource` |
| **Ghi chú** | Node "Animal" vẫn valid. Chỉ edge mới có lỗi |

---

### TC-05: err_self_loop.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/err_self_loop.drawio` |
| **Mô tả** | Edge với `source="2"` và `target="2"` — trỏ vào chính nó |
| **Expected Codes** | ✅ `SELF_LOOP_EDGE` (ERROR) |
| **Pipeline** | 🛑 STOP |
| **Go Test** | `TestIncorrect_SelfLoop` |
| **Ghi chú** | Node "SelfRef" valid. Edge mới là vấn đề |

---

### TC-06: warn_suspect_name.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/warn_suspect_name.drawio` |
| **Mô tả** | Node value = `<<interface>> IShow` — stereotype chưa được clean |
| **Expected Codes** | ✅ `SUSPECT_NODE_NAME` (WARN) |
| **Pipeline** | 🟢 CONTINUE (WARN không stop pipeline) |
| **Go Test** | `TestIncorrect_SuspectName` |
| **Ghi chú** | FilterErrors() phải trả 0. FilterWarns() phải trả ≥ 1 |

---

### TC-07: warn_trivial_name.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/warn_trivial_name.drawio` |
| **Mô tả** | Node có `value="A"` — tên 1 ký tự, rõ ràng là placeholder |
| **Expected Codes** | ✅ `TRIVIAL_NODE_NAME` (WARN) |
| **Pipeline** | 🟢 CONTINUE |
| **Go Test** | `TestIncorrect_TrivialName` |
| **Ghi chú** | Tên "A" có 1 rune ≤ 2, không rỗng → TRIVIAL trigger |

---

### TC-08: warn_incomplete_attr.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/warn_incomplete_attr.drawio` |
| **Mô tả** | Class "Account" có attribute `"- id"` — không có `:` (thiếu type) |
| **Expected Codes** | ✅ `INCOMPLETE_ATTRIBUTE` (WARN) |
| **Pipeline** | 🟢 CONTINUE |
| **Go Test** | `TestIncorrect_IncompleteAttribute` |
| **Ghi chú** | Attribute `"- name: String"` cùng class vẫn đúng — chỉ `"- id"` bị flag |

---

### TC-09: warn_incomplete_method.drawio
| Field | Value |
|---|---|
| **File** | `UMLs_testcase/incorrect/warn_incomplete_method.drawio` |
| **Mô tả** | Class "Database" có method `"+ connect("` — chưa đóng ngoặc |
| **Expected Codes** | ✅ `INCOMPLETE_METHOD` (WARN) |
| **Pipeline** | 🟢 CONTINUE |
| **Go Test** | `TestIncorrect_IncompleteMethod` |
| **Ghi chú** | Method `"+ disconnect(): void"` cùng class vẫn đúng — chỉ `"+ connect("` bị flag |

---

## Tổng Kết Coverage

| Severity | Error Codes Covered | Files |
|---|---|---|
| ERROR | EMPTY_GRAPH, EMPTY_NODE_NAME, INVALID_NODE_TYPE, DANGLING_EDGE_SOURCE, SELF_LOOP_EDGE | 5 files |
| WARN | SUSPECT_NODE_NAME, TRIVIAL_NODE_NAME, INCOMPLETE_ATTRIBUTE, INCOMPLETE_METHOD | 4 files |
| **Total** | **9 codes** | **9 files** |

> **Còn thiếu:** `DANGLING_EDGE_TARGET` — có thể thêm TC-04b nếu cần cover riêng.
