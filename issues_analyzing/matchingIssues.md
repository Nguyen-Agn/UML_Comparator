# Matching Issues Analysis

## Problem

Trong phần matching tên class, hệ thống hiện cho phép fuzzy match dựa trên độ giống ký tự. Điều này hợp lý với typo nhỏ như `Acount` so với `Account`, nhưng dẫn đến false positive với các tên khác nghĩa nhưng gần nhau về mặt chuỗi, ví dụ:

- `encryption` vs `decryption`
- `encrypt` vs `decrypt`

Các cặp này có thể bị match dù ý nghĩa hoàn toàn khác.

## Files Involved

### 1. [`matcher/fuzzy_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/fuzzy_matcher.go)

Đây là nơi tính điểm giống nhau giữa hai tên bằng Levenshtein thuần.

Điểm đáng chú ý:

- Hàm `Compare()` ở [matcher/fuzzy_matcher.go:22](/d:/vscode-workspace/UML_Comparator/matcher/fuzzy_matcher.go#L22)
- Logic chỉ:
  - lowercase
  - trim space
  - tính edit distance
  - chuẩn hóa về khoảng `[0,1]`

Vấn đề:

- Không có phân tích semantic.
- Không có tokenization identifier.
- Không có rule chặn các cặp đối nghĩa như `encrypt/decrypt`.

Kết quả là nếu hai tên giống nhiều ký tự, điểm vẫn cao.

### 2. [`matcher/standard_entity_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go)

Đây là file trực tiếp gây ra việc false positive được chấp nhận thành match.

Các điểm chính:

- Tại [matcher/standard_entity_matcher.go:59](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L59), `simScore` được lấy trực tiếp từ `fuzzyMatcher.Compare(solNode.Name, stuNode.Name)`.
- Tại [matcher/standard_entity_matcher.go:79](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L79), candidate được sort ưu tiên theo architecture similarity trước.
- Tại [matcher/standard_entity_matcher.go:109](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L109), chỉ cần `candidate.simScore >= minSimScore` là candidate được bind.
- Tại [matcher/standard_entity_matcher.go:131](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L131), pass 1 dùng threshold được truyền vào constructor.
- Tại [matcher/standard_entity_matcher.go:134](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L134), pass 2 hard-code `minSimScore = 0.4`.

Kết luận:

- Đây là file đang thực sự cho phép `encryption` match với `decryption`.
- `fuzzy_matcher.go` tạo ra điểm cao.
- `standard_entity_matcher.go` chấp nhận điểm đó và bind candidate.

## Why Changing Threshold To 1.0 Is Not Enough

Hiện tại nếu chỉ đổi constructor từ `0.8` lên `1.0` thì vẫn chưa đủ.

Lý do:

- Pass 1 đúng là sẽ nghiêm hơn, vì [matcher/standard_entity_matcher.go:132](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L132) dùng `m.similarityThresh`.
- Nhưng pass 2 ở [matcher/standard_entity_matcher.go:135](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L135) vẫn luôn chạy với ngưỡng `0.4`.
- Nghĩa là class không match ở pass 1 vẫn có thể được match ở pass 2 nếu architecture đủ giống và tên đạt mức fuzzy tối thiểu.

Vì vậy:

- Đổi threshold trong constructor lên `1.0` không đủ để đạt yêu cầu "chỉ chấp nhận 100% matched".

## Additional Note About Final Similarity

Tại [matcher/standard_entity_matcher.go:110](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L110), điểm cuối cùng được tính như sau:

`finalSim = 0.7 * archSim + 0.3 * nameSim`

Điều này có nghĩa:

- Ngay cả khi tên exact, `finalSim` vẫn chưa chắc bằng `1.0` nếu architecture không perfect.
- Do đó cần phân biệt rõ hai khái niệm:
  - `name exact match`
  - `overall similarity == 1.0`

Hai khái niệm này hiện không phải là một.

## Why Node ID Does Not Solve This

Một câu hỏi dễ phát sinh là: trong `builder`, mỗi node đều đã có `ID`, vậy tại sao matcher không dùng luôn `ID` để đối chiếu?

Lý do là `ID` trong graph hiện tại chỉ là ID nội bộ của Draw.io trong từng file, không phải semantic ID ổn định giữa file solution và file student.

Vai trò thực tế của `ID` hiện tại là:

- dùng để liên kết node với edge trong cùng một graph
- dùng để truy ngược owner của member hoặc endpoint của relation
- dùng làm key nội bộ khi ghi kết quả mapping

Nhưng `ID` này không thể dùng để match trực tiếp giữa hai file, vì:

- cùng một class có thể mang ID khác nhau hoàn toàn giữa solution và student
- chỉ cần vẽ lại, copy/paste, hoặc tạo diagram mới là Draw.io có thể sinh ID mới
- không có bảo đảm rằng `User` trong solution và `User` trong student sẽ giữ cùng một ID

Nói cách khác:

- `ID` giúp giữ integrity trong từng graph
- nhưng không giúp nhận diện identity của class xuyên qua hai file khác nhau

Vì vậy matcher hiện tại buộc phải dựa vào:

- name
- type
- architecture

Trong [`matcher/standard_entity_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go), `ID` chỉ được dùng sau khi đã match xong để lưu kết quả `solution node ID -> student node ID`, chứ không phải tiêu chí để quyết định match.

## Where Threshold Is Passed In

Threshold hiện được khởi tạo là `0.8` tại:

- [cmd/match/main.go:39](/d:/vscode-workspace/UML_Comparator/cmd/match/main.go#L39)
- [cmd/compare/main.go:92](/d:/vscode-workspace/UML_Comparator/cmd/compare/main.go#L92)
- [cmd/visualize/main.go:129](/d:/vscode-workspace/UML_Comparator/cmd/visualize/main.go#L129)
- [cmd/grade_batch/main.go:69](/d:/vscode-workspace/UML_Comparator/cmd/grade_batch/main.go#L69)

Nhưng các điểm này chỉ ảnh hưởng pass 1, không ảnh hưởng pass 2 hard-coded.

## Existing Test That Confirms Current Behavior

File [`matcher/standard_entity_matcher_test.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher_test.go) đã phản ánh rất rõ chủ đích hiện tại.

Đặc biệt:

- [matcher/standard_entity_matcher_test.go:145](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher_test.go#L145) `TestTwoPassMatching`

Test này cố tình mong đợi một class có tên rất khác vẫn được match ở pass 2 nếu architecture giống:

- `CruiseShip` vs `PPShip`

Điều đó cho thấy behavior hiện tại không phải bug ngẫu nhiên ở runtime, mà là đúng theo thiết kế matcher hiện tại.

## Conclusion

Nếu xét đúng câu hỏi "nó sai ở file nào", thì file chịu trách nhiệm chính là:

- [`matcher/standard_entity_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go)

Nguồn gốc của việc tên trái nghĩa vẫn có điểm cao nằm ở:

- [`matcher/fuzzy_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/fuzzy_matcher.go)

Nếu áp dụng yêu cầu hiện tại là "100% matched thì mới chấp nhận", thì chỉ đổi threshold từ `0.8` lên `1.0` là không đủ, vì:

- pass 2 vẫn cho match từ `0.4`
- và `finalSim` là điểm blend giữa architecture và name, không đồng nghĩa với exact name match

## Solution

Giải pháp phù hợp nhất với yêu cầu hiện tại là đổi rule matching cho class name từ fuzzy sang exact normalized match.

### Target Rule

Chỉ chấp nhận match class khi:

- node type giống nhau
- class name giống nhau 100% sau khi normalize

Architecture vẫn có thể giữ lại, nhưng chỉ nên dùng để:

- ưu tiên candidate nếu có nhiều node cùng tên
- hỗ trợ xếp hạng

Architecture không nên được phép "cứu" một candidate có tên khác.

### Recommended Matching Policy

Đối với class name:

- không dùng fuzzy threshold `0.8`
- không dùng pass 2 với ngưỡng `0.4`
- chỉ dùng exact normalized match

Đối với attribute/method:

- có thể tiếp tục giữ fuzzy matching nếu vẫn muốn hỗ trợ typo nhỏ

Lý do:

- class name là định danh chính của entity
- attribute và method thường có thể chấp nhận typo tolerance hơn

### Suggested Normalization

Để tránh quá cứng với khác biệt format đặt tên, nên normalize class name trước khi so sánh:

- lowercase
- trim space
- bỏ `_`
- bỏ `-`
- bỏ khoảng trắng

Ví dụ:

- `UserService` và `user_service` có thể coi là giống nhau
- `Encryption` và `Decryption` vẫn không giống nhau

### Required Matcher Changes

Nếu triển khai theo yêu cầu "100% matched thì mới chấp nhận", thì về logic matcher cần đổi như sau:

1. Trong [`matcher/standard_entity_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go)

- không bind candidate chỉ vì `candidate.simScore >= minSimScore`
- thay điều kiện accept bằng exact normalized name match

2. Loại bỏ hoặc vô hiệu hóa pass 2 cho class matching

- hiện tại pass 2 ở [matcher/standard_entity_matcher.go:134](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go#L134) vẫn cho match với ngưỡng `0.4`
- đây là phần trực tiếp làm hỏng yêu cầu "100%"

3. Không dùng `finalSim` làm tiêu chuẩn chấp nhận match tên class

- `finalSim` hiện là blend giữa architecture và name
- giá trị này phù hợp để report mức độ giống nhau
- nhưng không phù hợp để quyết định class identity nếu yêu cầu là exact name

### Practical Outcome

Sau khi áp dụng rule này:

- `Account` vs `Acount` sẽ không còn được match
- `Encryption` vs `Decryption` sẽ không còn được match
- `UserService` vs `user_service` vẫn có thể được match nếu normalization được áp dụng

### Trade-off

Ưu điểm:

- loại bỏ false positive nguy hiểm ở class identity
- behavior rõ ràng, dễ giải thích
- đúng hơn với logic chấm bài UML theo tên lớp chuẩn

Nhược điểm:

- typo thật ở class name sẽ không còn được bỏ qua
- số lượng unmatched class có thể tăng lên

Nếu mục tiêu hiện tại là độ chính xác nghiêm ngặt cho class identity, thì đây là trade-off hợp lý.

## Implementation Outlook

Nếu giả định hệ thống sẽ được chỉnh theo hướng "class chỉ được match khi exact normalized name", thì thay đổi này là khả thi và có phạm vi ảnh hưởng tương đối kiểm soát được.

Phần trọng tâm của chỉnh sửa sẽ nằm ở matcher, không nằm ở parser, builder hay cấu trúc domain cốt lõi. Nói cách khác, đây là thay đổi ở tầng decision logic chứ không phải thay đổi kiến trúc pipeline.

Phạm vi ảnh hưởng chính sẽ là:

- [`matcher/standard_entity_matcher.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher.go) vì đây là nơi quyết định candidate nào được accept
- [`matcher/standard_entity_matcher_test.go`](/d:/vscode-workspace/UML_Comparator/matcher/standard_entity_matcher_test.go) vì các test hiện tại đang xác nhận fuzzy/two-pass behavior cho class matching
- các entrypoint trong `cmd/*` đang khởi tạo matcher với threshold `0.8`, do ý nghĩa của threshold này có thể sẽ cần được làm rõ hoặc dọn lại
- phần hiển thị trong [`cmd/compare/main.go`](/d:/vscode-workspace/UML_Comparator/cmd/compare/main.go) nếu hệ thống vẫn dùng `Similarity == 1.0` để diễn giải là "perfect match"

Các phần ít hoặc gần như không bị ảnh hưởng:

- `builder/*` vì node `ID` vẫn chỉ là internal graph identifier
- `prematcher/*` vì logic tiền xử lý không quyết định class identity giữa hai graph
- `comparator/*` về mặt cơ chế, dù kết quả diff có thể thay đổi do mapping đầu vào nghiêm hơn

Kết luận thực tế là hướng sửa này hoàn toàn làm được, nhưng phải hiểu rằng đây không phải chỉ là việc tăng threshold từ `0.8` lên `1.0`. Để đạt đúng mục tiêu "100% matched thì mới chấp nhận", matcher phải đổi rule accept candidate, đồng thời loại bỏ hoặc vô hiệu hóa cơ chế pass 2 đang cho phép fuzzy rescue đối với class name.
