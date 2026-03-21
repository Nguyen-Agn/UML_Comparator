# Schema: DiffReport

**Package:** `domain`  
**Pipeline Stage:** Output của `IComparator`, Input của `IGrader` và `IVisualizer`

---

## Định nghĩa Go
```go
type DiffReport struct {
    MissingNodes    []string // Tên các Node có trong đáp án nhưng thiếu trong bài sinh viên
    MissingEdges    []string // Mô tả các Quan hệ thiếu (VD: "Dog -> Animal [Inheritance]")
    WrongAttributes []string // Mô tả chi tiết từng thuộc tính/method sai (VD: "Animal.age: missing")
    MatchPercentage float64  // Tỷ lệ % tổng thể match thành công (0.0 → 100.0)
}
```

---

## Quy ước định dạng chuỗi

### `MissingNodes`
Mỗi phần tử là tên Class/Interface/Actor bị thiếu.
```
["Payment", "ILogger"]
```

### `MissingEdges`
Mỗi phần tử theo format: `"SourceName -> TargetName [RelationType]"`
```
["Dog -> Animal [Inheritance]", "Order -> Payment [Association]"]
```

### `WrongAttributes`
Mỗi phần tử theo format: `"ClassName.issueDescription"`
```
[
  "Animal.Attribute 'age : int' missing",
  "Dog.Method 'bark() : void' has wrong visibility (should be public '+', got private '-')",
  "Edge Dog->Animal direction reversed"
]
```

---

## Ví dụ JSON tương đương

```json
{
  "MissingNodes": ["Payment"],
  "MissingEdges": ["Order -> Payment [Association]"],
  "WrongAttributes": [
    "Animal.Attribute 'age : int' missing",
    "Dog.Method has wrong return type"
  ],
  "MatchPercentage": 62.5
}
```
