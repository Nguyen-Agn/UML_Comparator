# Schema: GradeResult

**Package:** `domain`  
**Pipeline Stage:** Output cuối cùng của `IGrader`, đưa vào `IVisualizer` để render báo cáo

---

## Định nghĩa Go
```go
type GradeResult struct {
    TotalScore float64  // Điểm cuối (sau khi làm tròn 2 chữ số thập phân, >= 0)
    Feedbacks  []string // Danh sách nhận xét chi tiết cho từng lỗi
}
```

---

## Quy ước `Feedbacks`

Mỗi feedback theo format: `"[LOẠI_LỖI] Mô tả - Điểm trừ: X"`

| Prefix | Ý nghĩa |
|---|---|
| `[MISSING_NODE]` | Thiếu một Class/Interface/Actor |
| `[MISSING_EDGE]` | Thiếu một quan hệ |
| `[WRONG_ATTR]` | Sai thuộc tính hoặc phương thức |
| `[WRONG_DIRECTION]` | Vẽ mũi tên ngược chiều |
| `[TYPO]` | Tên class/attr bị gõ sai nhưng đã được nhận diện |

---

## Ví dụ JSON tương đương

```json
{
  "TotalScore": 6.5,
  "Feedbacks": [
    "[MISSING_NODE] Thiếu class 'Payment' - Điểm trừ: 1.0",
    "[MISSING_EDGE] Thiếu quan hệ 'Order -> Payment [Association]' - Điểm trừ: 0.5",
    "[WRONG_ATTR] 'Animal': thiếu thuộc tính 'age : int' - Điểm trừ: 0.5",
    "[TYPO] 'Acount' nhận diện là 'Account' - Không trừ điểm nhưng ghi chú sai chính tả"
  ]
}
```

## Bất Biến (Invariants)
- `TotalScore >= 0` — không bao giờ âm (được chặn bởi `math.Max(0, score)`).
- `TotalScore` được làm tròn 2 chữ số thập phân (`math.Round(val*100)/100`).
