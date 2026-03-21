# Chiến lược & Kế hoạch Phát triển Hệ thống So sánh UML (Draw.io)

Tài liệu này vạch ra chiến lược, kiến trúc và các bước thực hiện phần mềm chấm điểm và so sánh biểu đồ UML (được vẽ từ draw.io) giữa sinh viên và đáp án mẫu.

## 1. Phân Tích Bài Toán & Phạm Vi (Scope)

**Mục tiêu cốt lõi:**
- Đầu vào: 
  1. Biểu đồ UML đáp án mẫu (.drawio hoặc .xml)
  2. Biểu đồ UML bài làm của sinh viên (.drawio hoặc .xml)
  3. Bộ tiêu chí cấu hình điểm (VD: mỗi class đúng +1đ, đúng mỗi quan hệ +0.5đ, sai tên class -0.5đ...)
- Đầu ra: Bản báo cáo phân tích sự khác biệt (thiếu thực thể, sai tên, sai quan hệ, sai thuộc tính) và tổng điểm.

**Lựa chọn ngôn ngữ:**
- **Đề xuất: Golang (Go)**. Tuy Java có nhiều thư viện xử lý OOP, nhưng Go mang lại sự gọn nhẹ, dễ dàng build ra một file thực thi (binary) duy nhất để gửi cho giáo viên/trợ giảng chạy trên bất kỳ hệ điều hành nào (Windows/Mac/Linux) mà không cần cài thêm môi trường (như JRE của Java). Go cũng xử lý parse XML rất nhanh và hiệu quả.

## 2. Bản Chất Dữ Liệu Của Draw.io

File `.drawio` thực chất là một file XML, chứa thẻ gốc `<mxGraphModel>`.
- Dữ liệu có thể được lưu trữ ở dạng rõ (Plain XML) hoặc dạng nén (Compressed - mã hóa Base64 và nén Deflate).
- Mỗi phần tử trên biểu đồ (Class, Interface, Actor, Mũi tên...) là một thẻ `<mxCell>`.
- **Node (Thực thể):** Có `id`, nội dung (`value` hoặc cấu trúc HTML bên trong), và `style` (chứa các từ khóa như `class`, `swimlane`, `actor` để định hình).
- **Edge (Quan hệ):** Cũng là `<mxCell>` nhưng có thêm trường `source` (ID của node gốc) và `target` (ID của node đích). `style` sẽ định nghĩa đây là kế thừa (ext), liên kết (association) hay phụ thuộc (dependency).

## 3. Chiến Lược Đối Sánh (Matching Strategy)

Vì ID của các node trên bản vẽ của đáp án và của sinh viên sẽ hoàn toàn khác nhau (được draw.io sinh ngẫu nhiên), ta **không thể so sánh theo ID**.

### Bước 3.1: Tiền xử lý (Parsing & Standardization)
- Đọc file `.drawio` từ đáp án và bài sinh viên. Nếu nội dung bị nén (compressed XML), cần giải mã Base64 -> Inflate.
- Lọc ra các `<mxCell>`, phân loại chúng thành 2 mảng: **Nodes** (Thực thể: Class, Interface) và **Edges** (Quan hệ).
- Trích xuất nội dung (Bỏ đi thông tin x, y, width, height để biểu đồ không phụ thuộc vào vị trí hình học).

### Bước 3.2: Thuật toán Map Thực Thể (Node Mapping Algorithm)
- Ta cần tìm ra `Class A` trong bài Mẫu tương ứng với Node mang `ID = xyz` nào trong bài Sinh viên.
- **Tiêu chí Map:**
  - Chuẩn hóa text: xóa khoảng trắng, đưa về chữ thường.
  - Dựa vào Tên Thực thể (Tên Class, Tên Actor). Dùng thuật toán so sánh chuỗi (như Levenshtein Distance) để chấm chước các lỗi gõ sai chữ (typo) của sinh viên (Ví dụ giống 80% trở lên thì coi là map).

### Bước 3.3: So sánh Chi Tiết Thực Thể (Detail Comparison)
- Khi đã kết nối được 2 node tương ứng. Tiếp tục phân tách nội dung bên trong node:
  - Danh sách Thuộc tính (Attributes)
  - Danh sách Phương thức (Methods)
- So sánh các element này để tìm ra sự khác biệt (thừa/thiếu/sai kiểu dữ liệu).

### Bước 3.4: Đối Sánh Quan Hệ (Edge Matching)
- Sử dụng bảng Map ID ở Bước 3.2.
- Ví dụ bài Mẫu có mũi tên từ `Dog` tới `Animal` (Kế thừa).
- Ta tìm trong bài sinh viên: Node sinh viên ứng với `Dog` có mũi tên (style: kế thừa) trỏ tới Node sinh viên ứng với `Animal` không?
- Đánh giá hướng mũi tên, loại quan hệ (multiplicity: 1..*, 0..1).

## 4. Kiến Trúc Phần Mềm

Kiến trúc chính thức chia thành **6 thư mục module** theo đúng nguyên tắc SOLID (SRP + DIP):

| Thư mục | Loại | Mô tả |
|---|---|---|
| `domain/` | Core Models | Định nghĩa tất cả struct dùng chung (`UMLGraph`, `DiffReport`...) |
| `parser/` | Interface + Impl | Đọc & giải nén file `.drawio` → `RawModelData` |
| `builder/` | Interface | Biến `RawModelData` → `UMLGraph` (string-based) |
| `prematcher/` | Interface | Biến `UMLGraph` → `ProcessedUMLGraph` (struct-based) |
| `matcher/` | Interface | So khớp Node 2 bên → `MappingTable` |
| `comparator/` | Interface | Duyệt diff chi tiết → `DiffReport` |
| `grader/` | Interface | Áp rule điểm → `GradeResult` |
| `visualizer/` | Interface | Xuất file báo cáo (`.drawio` màu / HTML) |
| `scheme/` | Documentation | **Lưu mẫu cấu trúc dữ liệu & cấu hình** (xem Mục 4.1) |
| `Skills/` | Agent Context | Hệ thống Skill OS 5-Layer cho AI Agent (xem `Agent_Instruction.md`) |
| `demo/` | Runner | File chạy thủ công để kiểm tra từng module |

### 4.1. Thư mục `scheme/` — Single Source of Truth

Thư mục `scheme/` lưu định nghĩa chính thức của từng kiểu dữ liệu lưu chuyển giữa các module. Bất kỳ thay đổi cấu trúc nào trong `domain/models.go` cũng **phải** đồng bộ vào đây.

| File Schema | Stage (Pipeline) | Mô tả |
|---|---|---|
| `scheme/raw_xml_data.md` | `Parser` → `Builder` | Bất biến & format của chuỗi XML thuần |
| `scheme/uml_graph.md` | `Builder` → `Matcher`/`Comparator` | Cấu trúc `UMLGraph`, `UMLNode`, `UMLEdge` và bảng `RelationType` |
| `scheme/diff_report.md` | `Comparator` → `Grader`/`Visualizer` | Format chuỗi lỗi, quy ước prefix `[MISSING_NODE]`... |
| `scheme/grade_result.md` | `Grader` → Output | Quy tắc làm tròn điểm, bất biến `TotalScore >= 0` |
| `scheme/grading_rules.json` | Config đầu vào `Grader` | Bảng điểm trừ theo từng loại lỗi (do giáo viên điều chỉnh) |

## 5. Kế Hoạch Triển Khai (Roadmap)

### Giai Đoạn 1: POC & Parsing (Tuần 1)
- Tạo 2 file `.drawio` đơn giản: 2 Class liên kết với nhau.
- Viết hàm đọc file `.drawio`, giải nén ra XML và lấy được nội dung vào struct đơn giản.
- In ra console danh sách các Class name và các ID liên kết.

### Giai Đoạn 2: Xây Dựng Model & Mapping (Tuần 2)
- Định nghĩa các Struct dùng chung (UMLStruct, Class, Attribute, Method, Relationship).
- Xây dựng thuật toán Matching các Class dựa theo Text similarity.

### Giai Đoạn 3: Engine So Sánh Cốt Lõi (Tuần 3)
- Hiện thực hoá logic so sánh thành phần bên trong Class.
- Nhận diện đủ và đúng các loại Relationship (Inheritance, Realization, Association, Aggregation, Composition).
- Phát hiện mũi tên ngược chiều.

### Giai Đoạn 4: Đánh Giá & Giao Diện (Tuần 4)
- Thiết lập file `.json` để khai báo luật chấm điểm.
- Định dạng format report output (ví dụ file JSON có thể đọc lại lên UI, hoặc xuất Text thuần).
- (Tùy chọn) Xây dựng Web UI kéo-thả file `.drawio` vào, hiện so sánh trực quan.

## 6. Các Điểm Rủi Ro (Risks & Mitigations)
- **Sinh viên vẽ bằng nhiều đối tượng ghép lại:** Đôi khi sinh viên không dùng UML shape chuẩn của draw.io mà tự vẽ 1 box Text ghép với 1 hình vuông.
  > **Khắc phục:** Nên có 1 file hướng dẫn template chuẩn bắt buộc sinh viên sử dụng công cụ UML Tool có sẵn ở thanh bên trái của Draw.io.
- **Nội dung HTML phức tạp trong thẻ value:** File xml draw.io thường chèn cả thẻ `<div>`, `<b>`, `<i>` vào nội dung.
  > **Khắc phục:** Cần viết một hàm sanitize_html để bóc tách lấy pure text trước khi so sánh.

## 7. Chiến Lược Testing & Visualization

### 7.1. Trực quan hóa kết quả (Visualization)
Có 2 phương pháp chính để sinh ra báo cáo trực quan từ `IVisualizer`:

1. **Native Draw.io (Khuyên dùng):**
   - **Cách hoạt động:** Hệ thống tạo ra một bản copy file `.drawio` bài làm của sinh viên (VD: `student_graded.drawio`).
   - Dựa vào kết quả của `DiffReport`, hệ thống can thiệp thẳng vào thẻ XML để đổi màu `style` các Component:
     - **Xanh lá (Green):** Làm đúng hoàn toàn.
     - **Đỏ (Red):** Vẽ sai, hoặc bị thiếu (hệ thống tự insert thêm block mờ màu đỏ chỉ ra chỗ thiếu).
     - **Vàng (Yellow):** Class đúng nhưng chi tiết bên trong sai (VD: sai kiểu dữ liệu thuộc tính).
   - **Lợi ích:** Trực quan nhất, sinh viên nhận lại bài mở lên phân tích ngay trên hình mình vẽ. Phù hợp với cách sửa bài tự nhiên.
2. **HTML / Mermaid Report:**
   - Render ra 1 file HTML tĩnh, sử dụng thư viện Mermaid.js để vẽ lại các Entity bị sai, kết hợp dạng bảng Table Side-by-Side (Bài Mẫu vs Bài Sinh viên) để sinh viên dễ đọc lỗi dạng Text.

### 7.2. Phương pháp Testing Từng Module (Unit Testing & Mocks)
Vì áp dụng chuẩn SOLID (phân chia bằng Interface), việc test ở Golang sẽ vô cùng độc lập (Isolation) bằng kỹ thuật Mock Object kết hợp **Table-Driven Tests** cực kỳ đỉnh của Go.

- **Test `IFileParser` (Input / Output):**
  - Không test nội dung đồ thị. Chỉ đưa vào thư mục `testdata/` các mẫu file thực tế (`.drawio` cũ/mới, file giải nén, file trống).
  - Viết test kiểm tra xem chương trình có bung ép file Base64 ra chữ XML thô được không, và không bị trượt/crash lỗi nhảm.
- **Test `IModelBuilder`:**
  - Inject 1 chuỗi XML thô ngắn gọn tự viết tay. Test xem vòng lặp có đẩy nó thành Struct `UMLGraph` chuẩn, và regex hàm có lôi ra chữ `int` từ chuỗi `+ age : int` hay không.
- **Test `IEntityMatcher` & `IComparator` (Lõi logic thuật toán):**
  - **Không cần đọc file (No IO).** Do cấu trúc nhận vào là `UMLGraph`, ta tự khởi tạo data giả ngay trong code unit test.
  - *Ví dụ:* Code chạy giả lập `UMLGraph` mẫu chứa Class `Animal`, và `UMLGraph` sinh viên chứa Class `Animals` (Sinh viên gõ sai chữ `s`). 
  - Đẩy vào hàm chạy test xem thuật toán bù Typo có "nhận diện" được hai node này là một và bắt lỗi syntax hay không. Việc này chạy cực mạnh và nhanh.
- **Test `IGrader`:**
  - Nhập vào 1 `DiffReport` giả vờ có 1 lỗi sai Tên, 1 điểm cộng Mũi tên đúng. Truyền vào Config JSON -> Test xem điểm máy chấm ra có khớp phép cộng trừ toán học hay không.
