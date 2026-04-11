# Tài Liệu kiến trúc phần mềm

## I. Kiến trúc
 
### 1. Mô tả chung

   Phần mềm sẽ được chạy dựa trên kiến trúc pipeline.
   Chạy qua các module(giai đoạn khác nhau) từ đầu vào đến kết quả.

   * Parse - (filter content from file .drawio/.solution)
   |
   * Build - (build raw graph)
   |
   * PreMatch - (transfrom raw graph to processable graph)
   |
   * Match - (match nodes -> know who is who)
   |
   * Compare - (compare with members like attributes, methods, relations)
   |
   * Grade - (calculate score)
   | \
   |  * Report - (create report .html)
   |  |
   |  * Visualize - (create visualize .html)
   * Grade_Batch - (create list result .csv)

### 2. Parse

   Đây là module đầu tiên. 
   **Nhiệm vụ**: nhận đưỡng dẫn của tệp *.drawio/.solution* đồng lấy nội dung của tệp
   Tùy vào đuôi file mà có cách parse khác nhau
   - .drawio: lọc các thẻ *html/xml* để lấy nội dung chính của UML (thường là *xmlCell* còn sót lại)
   - .solution: sử dụng giải thuật giải mã cipher để lấy nội dung chính của UML 

```go
type IFileParser interface {
	Parse(filePath string) (domain.RawModelData, error)
}
```

### 3. Build
 
   Module này nhận đầu vào từ Parse.
   **Nhiệm vụ**: Đọc nội dung được truyền vào và chuyến đổi các thành phần UML thành các cấu trúc cây ( nhưng chủ yếu là chuỗi) để chuẩn bị cho bước sau. Ngoài ra, trong module này các định dạng như gạch chân, in đậm, in nghiêng sẽ được chuyển đổi thành các mẫu để chuẩn hóa dữ liệu.

```go
type IModelBuilder interface {
	Build(rawData domain.RawModelData) (*domain.UMLGraph, error)
}
```

### 4. PreMatch

    Module này nhận đầu vào từ Build và đóng vai trò tiền xử lý cho Match.
   **Nhiệm vụ**: Chuyển hóa cấu trúc cây chuỗi thô từ Build thành các cấu trúc dữ liệu tiêu chuẩn. Trong module này, đồng thời sẽ phân tích, đánh dấu và phân loại các thành phần UML (như tên lớp, thuộc tính, phương thức, quan hệ) để chuẩn bị cho bước Match. 
    Ngoài ra, còn chịu trách nhiệm tạo ArcWeight( đây là 1 trọng số đại diện cho kiến trúc, độ phức tạp của UML). `ArcWeight` dùng dưới kiến trúc bitwise, cần có ít khác biệt dưới dạng bitwise thì Class càng có xu hướng giống nhau.
    Tùy vào đây là solution hay student mà có cách xử lý khác nhau
    - student: dùng cách phân tích thông thường như mô tả. Tên thành phần là chuỗi
    - solution: các thành phần UMl sẽ có thể là mảng chuỗi. Đồng thời phân tích __*__ (điểm thành phần tương ứng)

```go
type IPreMatcher interface {
	Process(graph *domain.UMLGraph) (*domain.ProcessedUMLGraph, error)
}

type IUMLSolutionPreMatcher interface {
	ProcessSolution(graph *domain.UMLGraph) (*domain.SolutionProcessedUMLGraph, error)
}
```

### 5. Match

    Module này nhận 1 ProcessedUMLGraph và 1 SolutionProcessedUMLGraph để tiến hành đối chiếu.
   **Nhiệm vụ**: Đối chiếu để biết class nào là tương ứng với nhau trong 2 UML và tạo ra 1 bảng so sánh. Nhờ đó, module tiến chỉ cần xử lý trên những cặp class tương đồng, nhầm tối ưu hiệu suất và độ chính xác.
    Ban đầu, module tiến hành sắp xếp các class dựa theo ArcWeight rồi tiến hành đối chiếu các class gần nhau.
    Lần 1: nếu archSim >= 0.85 thì FinalSim >= 0.8 (và là lớn nhất) thì xem là tương đồng
        * Lần 1 là đối chiếu chủ lực, đối chiếu cả kiến trúc và tên class
    Lần 1.5: Đâu không phải đối chiếu chính, chủ yếu tính ngữ nghĩa của tên class và loại bỏ các cặp tên gần giống nhau về mặc từ nhưng khác nhau hoàn toàn về ngữ nghĩa. (ví dụL "Encrytor" và "Decryptor")
    Lần 2: nếu archSim >= 0.9 thì FinalSim >= 0.4 (và là lớn nhất) thì xem là tương đồng
        * Lần 2 là với, trong trường tên quá khác biệt nhưng chấp nhận về mặt ngữ nghĩa thì đối chiếu, nhưng buộc phải có kiến trúc rất giống nhau
> FinalSim = (archSim * 0.7) + (simScore * 0.3)

```go
type IEntityMatcher interface {
	Match(solution *domain.SolutionProcessedUMLGraph, student *domain.ProcessedUMLGraph) (domain.MappingTable, error)
}
```

### 6. Compare

    Sau khi nhận được bảng so sánh từ Match, module này sẽ tiến hành so sánh chi tiết các thành phần của các class tương đồng.
   **Nhiệm vụ**: So sánh các thành phần của các class tương đồng và tạo ra 1 bảng so sánh chi tiết.
    Trước khi  tiến hành, module sẽ tạo 1 bảng ánh xạ các kiểu dữ liệu từ solution sang student.
    Tùy vào từng thành phần mà có cách so sánh khác nhau:
    - Thuộc tính:
        + Tên: fuzzy > 0.8 với 1 trong những tên trong mảng của solution
        + Kiểu dữ liệu: 100% khớp sau khi qua ánh xạ với 1 trong những kiểu dữ liệu trong mảng của solution
        + Đối với generic type:
            * Loại bên ngoài: so sánh như so sánh tên hoặc bao gồm (contains như List == ArrayList)
            * Loại bên trong: so sánh như so sánh kiểu thường
        + Constructor/ Method:
            * Tên: fuzzy > 0.8 với 1 trong những tên trong mảng của solution
            * Các tham số: không so sánh tên chỉ  so sánh kiểu dữ liệu sau khi qua ánh xạ, số lượng tham số: >= với số lượng tham số của solution +1, nhưng buộc phải có giống toàn bộ các tham số của solution
            * Kiểu trả về: 100% khớp sau khi qua ánh xạ với 1 trong những kiểu trả về trong mảng của solution
    - Quan hệ:
        + So sánh loại 
        + So sánh hướng
    
    - Ngoài ra còn 1 đợt vớt: chủ yếu tìm ra các thành phần sai kiểu dữ liệu

    * Kết quả: 
        - Nếu khớp hoàn toàn -> CorrectDetail
        - Nếu có sai lệch (Sai kiểu dữ liệu, sai phạm vi, sai loại) -> WrongDetail kèm mô tả chi tiết lỗi.
        - Nếu không thể bắt cặp -> ExtraDetail (Thừa) hoặc MissingDetail (Thiếu)
    
```go
type IComparator interface {
	Compare(solution *domain.SolutionProcessedUMLGraph, student *domain.ProcessedUMLGraph, mapping domain.MappingTable) (*domain.DiffReport, error)
}
```

### 7. Grade

    Module này nhận đầu vào từ Compare và cặp SolutionProcessedUMLGraph và ProcessedUMLGraph và GradingRules ( từ PreMatch của solution phân tích và trả về trước đó)
    **Nhiệm vụ**: Tính điểm dựa trên kết quả so sánh.
    Module sẽ duyệt qua từng thành phần có trong SolutionProcessedUMLGraph và tìm kiếm trong DiffReport xem có trong CorrectDetail không? Nếu có thì cộng điểm, nếu không có thì trừ điểm ( từ GradingRules).
    Đồng thời tạo Feedback chi tiết cho từng thành phần.
    
```go
type IGrader interface {
	// Grade computes the final GradeResult.
	Grade(report *domain.DiffReport, sol *domain.SolutionProcessedUMLGraph, stu *domain.ProcessedUMLGraph, rule *GradingRules) (*domain.GradeResult, error)
}
```

### 8. Visualize

    Module này nhận đầu vào từ Compare và tiến hành tạo báo cáo dưới dạng HTML.
    **Nhiệm vụ**: Tạo báo cáo trực quan dựa trên kết quả so sánh.
    Module duyệt qua từng thành phần có trong DiffReport và 2 UMLGraph và thêm(append,insert) vào 1 mẫu html có sẵn.
    Trong bản cho sinh viên, sẽ không hiển thị nội dung của solution và chi tiết lỗi sai cụ thể. 

    Tùy vào từng chế độ. nội dung html có thể được hiển thị trực tiếp qua GUI module hoặc lưu vào tệp mới.
    
```go
type IVisualizer interface {
	ExportHTML(result *domain.GradeResult, outputPath string) error
	ExportStudentHTML(result *domain.GradeResult, outputPath string) error
}
```

### 9. Grade_Batch / Report

    Module này có tính độc lập so với các module khác.
    **Nhiệm vụ**: Tạo báo cáo dựa trên kết quả chấm đồng loại của nhiều sinh viên.
    Sau đó chạy vòng lặp và tạo 1 têp .csv để lưu trữ kết quả.
    
```go
type IReporter interface {
	GenerateReport(batchResult *BatchGradeResult) error
}
```

### 10. GUI

    Module này không liên quan đến quá trình chấm điểm, nó chỉ là giao diện để người dùng tương tác với hệ thống.
    **Nhiệm vụ**: Tạo giao diện người dùng để tương tác với hệ thống.
    
    Module yêu cầu: máy người dùng có Chromium đề hiển thị giao diện nhập xuất.

### 11. Cipher

    Module này có tính độc lập so với các module khác.
    **Nhiệm vụ**: Mã hóa .drawio của solution để tránh việc sinh viên đọc được solution trước khi nộp bài.
    Sau khi mã hóa, tệp sẽ có đuôi .solution

