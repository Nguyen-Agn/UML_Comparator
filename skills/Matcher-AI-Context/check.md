# Verification: Matcher Quality Checklist

Dùng checklist này để kiểm định quá trình code Matcher:

## 1. Structural Correctness
- [ ] Tham số hàm `Match()` đã đổi đúng thành `*domain.ProcessedUMLGraph`?
- [ ] Module không thay đổi các node trong 2 biến con trỏ truyền vào (Immutability)?
- [ ] IFuzzyMatcher được pass qua qua constructor Injection `NewStandardEntityMatcher()`?

## 2. ArchWeight-First Logic Validation
- [ ] Module dùng `ArchWeight` để sort the candidates (delta càng nhỏ thì index càng nhỏ)?
- [ ] Nếu 2 node cùng `ArchWeight` có được xử lý tie-breaker hợp lý không?

## 3. Map Output Restrictions
- [ ] MappingTable chỉ mapping những cặp vượt qua ngưỡng `Threshold > 0.8`?
- [ ] Bắt buộc đảm bảo `1:1 mapping mapping`. Nếu Node sinh viên `Stu_X` đã map cho `Sol_A`, thì tuyệt đối khôg được tham gia xét map cho `Sol_B` nữa.
