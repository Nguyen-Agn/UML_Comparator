# PROJECT CONTEXT FOR AI AGENT

Đây là tệp tài liệu tóm tắt toàn bộ dự án `UML Comparator`. Bất kỳ AI Agent nào khi tham gia hỗ trợ dự án đều **PHẢI** đọc và tuân thủ các thông tin trong file này trước khi code.

## 1. Tóm tắt dự án (Project Overview)
**Tên dự án:** UML Comparator
**Mục tiêu:** Hệ thống tự động so sánh, chấm điểm và sinh biểu đồ UML (hỗ trợ file `.drawio` và code `Mermaid`). Hệ thống được sử dụng chủ yếu bởi Giảng viên (Instructor) để chấm bài, mã hóa đáp án và tạo đề thi; và bởi Sinh viên (Student) để làm bài và kiểm tra điểm.
**Ngôn ngữ chính:** Golang (Backend), HTML/CSS/VanillaJS (Frontend - thông qua Lorca).
**Thư viện đồ họa:** Mermaid.js (cho phần tạo UML bằng AI).

## 2. Kiến trúc Hệ thống (Architecture)
Hệ thống xử lý bài làm UML thành một 파ipeline theo thứ tự:
1. **Parser:** Đọc XML thô từ `.drawio` (Base64/Inflate).
2. **Builder:** Chuyển đổi dữ liệu XML thành đồ thị dạng string (`UMLGraph`).
3. **Pre-Matcher:** Trích xuất các thực thể UML (Class, Interface, Relation) từ String Graph sang dạng object/struct.
4. **Matcher:** So khớp (Mapping) các node giữa Bài làm (Student) và Đáp án (Solution) dựa trên thuật toán tìm kiếm mờ (Fuzzy) và Semantic AI Similarity.
5. **Comparator:** Đối chiếu chi tiết (thuộc tính, phương thức, quan hệ) sau khi đã map các node.
6. **Grader:** Tính điểm chuyên sâu từ `DiffReport`.
7. **Visualizer:** Kết xuất báo cáo ra định dạng HTML tự chứa (Dawn's Berry theme).
8. **UML Generator:** Gọi API AI (OpenAI/Ollama/Groq) để tự động sinh code Mermaid từ một đoạn văn bản (đề bài). Nằm trong thư mục `src/uml_generator`.

## 3. Các ứng dụng (Apps)
Dự án được phân tách thành nhiều ứng dụng executable (`.exe` trên Windows, binary trên Linux) thông qua thư mục `cmd/`:
- `cmd/instructor/`: Giao diện **Instructor Suite** (chứa Live Compare, Batch Grader, Solution Encrypt, Exam Builder).
- `cmd/umlGenerator/`: Giao diện **UML Generator** giúp tạo mẫu Mermaid từ văn bản AI.
- `cmd/visualize/`: Phiên bản dòng lệnh (CLI) để học sinh tự so sánh.
- `gui/`: Chứa mã nguồn cho giao diện người dùng **Student GUI**.
- `portable/`: Thư mục lưu trữ tất cả các bản build đã sẵn sàng sử dụng.

## 4. Nguyên tắc Thiết kế BẮT BUỘC (Strict Rules)
1. **Interface First:** Mọi tương tác giữa các module đều **PHẢI** thông qua `Interface` được định nghĩa trong package `domain` (nếu liên kết rộng) hoặc tự định nghĩa tại package đó. Tuyệt đối không để class/struct nói chuyện trực tiếp với nhau.
2. **Open/Close Principle (OCP):** Cố gắng mở rộng thay vì chỉnh sửa code hiện có. Không sửa đổi các module khác (như Parser, Builder) nếu đang làm việc tại module UML Generator.
3. **No Cross-Talk:** Các module không được biết về nội bộ của nhau. Luôn coi Interface cung cấp data "đúng như comment".
4. **HTML/JS Embedding:** Giao diện được nhúng hoàn toàn vào file thực thi bằng `go:embed`. Tất cả HTML nằm tại thư mục `gui/view/`. **KHÔNG** sử dụng Framework Frontend (React, Tailwind) trừ khi User yêu cầu rõ ràng. Dùng Vanilla CSS & JS.
5. **Bảo tồn comments:** Không tự ý xóa bỏ các comment hiện có trong code Go.
6. **Xử lý lỗi:** Mọi API gọi ra bên ngoài (như AI endpoint) phải có Error Handling chặt chẽ (vd: bắt lỗi 404, lỗi EOF).

## 5. Hướng dẫn thử nghiệm (Testing)
Tất cả các thành phần module đều có code test nằm trong thư mục `cmd/`. Ví dụ, để chạy thử nghiệm module Compare, có thể dùng `make run_compare`. Các script chạy tự động khác được khai báo trong file `Makefile` gốc.

**Dành cho AI Agent:**
Bất cứ khi nào bắt đầu một tác vụ mới, hãy tham chiếu các nguyên tắc thiết kế ở mục 4 để tránh làm vỡ các module cốt lõi khác. Mọi sửa đổi trên giao diện Lorca (HTML/JS) phải kiểm tra cẩn thận để tránh lỗi parse chuỗi JSON từ Go sang JS.
