# Improved Matching Strategy: 3-Stage Pipeline (Fuzzy -> Filter -> Rank)

## 🎯 Goal
Khắc phục tình trạng false positive (nhận diện sai) đối với các node có tên mang ý nghĩa trái ngược nhưng lại có độ tương đồng chuỗi cao:
- ❌ Hủy bỏ: `EncryptService` ↔ `DecryptService`
- ✅ Giữ lại: Khả năng chịu đựng lỗi đánh máy nhỏ (VD: `Acount` vs `Account`) và sai khác chuẩn đặt tên (VD: `UserService` vs `user_service`).

---

## 🧠 Core Concept
Thay vì chỉ dựa thuần túy vào độ tương đồng chuỗi (`fuzzyMatcher.Compare()`) để đưa ra quyết định hay chuyển hẳn sang "100% exact match", ta sẽ chia tách luồng matching thành 3 giai đoạn độc lập. Cách tiếp cận này giúp giữ lại được sự linh hoạt (typo-tolerance) nhưng đồng thời bổ sung sự chặt chẽ về mặt ngữ nghĩa (semantic correctness).

### The Pipeline
1. **Candidate Generation (Fuzzy)**: Sử dụng các thuật toán như Levenshtein để gom các candidate tiềm năng.
2. **Identity Filter (Strict Rules)**: Token hóa các identifier để áp dụng các bộ lọc ngữ nghĩa (Từ trái nghĩa, độ phủ token).
3. **Ranking (Scoring)**: Chỉ áp dụng công thức tính điểm và sắp xếp cho các candidate **đã vượt qua** phễu Identity Filter. Nhờ vậy, điểm kiến trúc (Arch Weight) không thể "cứu" 1 candidate sai ngữ nghĩa.

---

## 🛠️ Detailed Implementation Plan

### Stage 1: Candidate Generation (Current `IFuzzyMatcher`)
- Vẫn dùng `LevenshteinMatcher` để tính `simScore` ban đầu.
- Ở giai đoạn này, các cặp như `EncryptService` và `DecryptService` vẫn sẽ ra điểm cao (ví dụ `0.85`) và được thêm vào candidate pool.

### Stage 2: Identity Filter (New Component)
Đây là "người gác cổng" bắt buộc phải đi qua trước khi Candidate được đưa vào list sắp xếp.

#### Process:
1. **Tokenization (Tách từ)**: Phân rã theo chuẩn PascalCase, camelCase, snake_case...
   - `EncryptService` → `["encrypt", "service"]`
   - `DecryptService` → `["decrypt", "service"]`
2. **Validation Rules (Lọc)**: 
   - **Antonym Reject:** Kiểm tra qua một từ điển hoặc engine xử lý từ trái nghĩa. Nếu phát hiện `encrypt` ↔ `decrypt`, candidate lập tức bị **LOẠI BỎ**.
   - **Token Overlap:** Cần đảm bảo core concept giữa 2 tên có sự trùng lặp (ví dụ class `Manager` và class `Service` dù có fuzzy score ổn nhưng không có token nào chung thì cần xem xét loại bỏ).

#### Detecting Antonyms
Xây dựng một module nhỏ chuyên phát hiện các từ đối xưng trong lập trình:
1. **Dictionary-based**: Các cặp cố định.
   ```go
   var antonymPairs = map[string]string{
       "encode": "decode",
       "login": "logout",
       "open": "close",
       "lock": "unlock",
       // ...
   }
   ```
2. **Prefix-based Morphology**: Dựa trên các tiền tố trái nghĩa bổ nghĩa cho cùng một root word.
   - Root: `"crypt"`
   - Prefixes: `"en-"` và `"de-"` 
   - => Cùng root nhưng trái prefix => Trái nghĩa.

### Stage 3: Ranking and Assignment
- Chỉ những candidate "sạch" (đã qua Identity Filter) mới được đưa vào sort bằng 3-tier algorithm ở `standard_entity_matcher.go` (Ưu tiên theo: ArchWeight -> FuzzyScore -> ArchDelta).
- Do candidate rác đã bị lọc hết, việc gán điểm (finalScore) với ngưỡng `minSimScore` ở các Pass được an toàn mà không sợ sai lệch ý nghĩa.

---

## 🚀 Feasibility Analysis in Current Architecture
Phương pháp này **CỰC KỲ KHẢ THI** và rất "vừa vặn" với kiến trúc hiện tại của dự án. 

**Tại sao phương pháp này lại hoàn hảo:**
- **Không làm vỡ core logic:** Luồng Pass 1 & Pass 2 cũng như cách sắp xếp Candidate tại file `matcher/standard_entity_matcher.go` hoàn toàn được giữ nguyên. 
- **Isolated Change (Thay đổi mang tính module):** Ta chỉ cần inject thêm một bước check ngay khi tạo Candidate (`matcher/standard_entity_matcher.go` - dòng ~60).
  ```go
  // ... Logic hiện tại
  simScore := m.fuzzyMatcher.Compare(solNode.Name, stuNode.Name)
  
  // NEW: Pass through Identity Filter
  if !m.identityValidator.IsValid(solNode.Name, stuNode.Name) {
      continue // Skip add candidate
  }
  
  // ... Push to candidates list
  ```
- **Bảo toàn Test Cases:** Các test case có trong `standard_entity_matcher_test.go` khẳng định khả năng chịu typo như `Acount` sẽ vẫn pass xanh rờn. Ta chỉ cần bổ sung thêm mock test để catch các case Antonyms.
