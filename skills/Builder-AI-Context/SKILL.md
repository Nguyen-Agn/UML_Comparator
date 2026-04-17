# SKILL: uml-model-builder
## Purpose
Chuyển đổi chuỗi RawModelData (đã được Parser làm sạch) thành cấu trúc domain.UMLGraph thống nhất dựa trên sourceType (drawio, mermaid).

## Required Inputs
- domain.RawModelData (Chuỗi dữ liệu thô sạch từ IFileParser)
- sourceType (Chuỗi: "drawio" hoặc "mermaid")

## Expected Output
- *domain.UMLGraph (Mô hình OOP hoàn chỉnh)

## Strategy Architecture
Hệ thống sử dụng **AutoBuilder** làm orchestrator (Strategy Pattern) để điều phối việc build dựa trên `sourceType`.

### 1. Draw.io Builder (`builder/drawio`)
Dành cho dữ liệu XML từ Draw.io.
- **IXMLParser**: Parse XML thô và cấu trúc cha-con.
- **ITextSanitizer**: Làm sạch HTML entities và tags.
- **ITypeDetector**: Nhận diện Class/Interface/Enum dựa trên metadata/style.
- **IMemberParser**: Phân tích Attributes và Methods từ child cells.

### 2. Mermaid Builder (`builder/mermaid`)
Dành cho dữ liệu text DSL từ Mermaid.
- **RegexParser**: Sử dụng biểu thức chính quy để trích xuất class blocks, stereotypes và relationships.
- **Stereotype Mapping**:
    - `<<interface>>` -> Interface
    - `<<abstract>>` -> Abstract
- **Relationship Mapping**: Hỗ trợ `<|--`, `..|>`, `*--`, `o--`, `-->`, `..>`.

## Execution Approach
1.  **Orchestration**: Gọi `builder.GetBuilder(sourceType)`.
2.  **Build**: Thực thi phương thức `Build` của concrete strategy tương ứng.
3.  **Validation**: Kết quả trả về là một `UMLGraph` chuẩn hóa, sẵn sàng cho việc so sánh.

## Compile-time Guarantees
Mỗi concrete strategy trong `AutoBuilder` được đảm bảo thoả mãn interface `IModelBuilder` tại thời điểm biên dịch.
