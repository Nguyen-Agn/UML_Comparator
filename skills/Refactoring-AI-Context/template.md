# TEMPLATE: Dàn Giáo Tham Khảo Cho Refactoring Module

Dưới đây là một pattern chuẩn nên sử dụng làm template để code lại một module bất kỳ ở Golang trong dự án nhằm tuân thủ quy tắc Interface-First, SOLID và Unit Test chuẩn.

## 1. Mẫu Thiết kế Interface (Interface-first Design)

```go
// Tên file: [module_name]_interface.go (hoặc gộp vào file gốc domain)
package mymodule

// MyModuleService định nghĩa các nghiệp vụ cốt lõi mà module cung cấp.
type MyModuleService interface {
    // ProcessData thực hiện xử lý dữ liệu thô []byte và trả về OutputModel.
    // Nếu data rỗng hoặc sai format, hàm trả về error định nghĩa sẵn.
    ProcessData(data []byte) (*OutputModel, error)
}
```

## 2. Mẫu Triển khai Implementation (SRP & Dependency Injection)

```go
// Tên file: [module_name]_impl.go
package mymodule

// concreteService triển khai interface MyModuleService.
// Struct này được giữ ở trạng thái unexported (private).
type concreteService struct {
    // Inject dependency qua Interface, tuyệt đối không dùng concrete struct.
    validator ValidatorInterface 
}

// NewConcreteService là constructor trả về interface MyModuleService.
func NewConcreteService(v ValidatorInterface) MyModuleService {
    return &concreteService{
        validator: v,
    }
}

// ProcessData là implementation thực tế.
func (s *concreteService) ProcessData(data []byte) (*OutputModel, error) {
    if err := s.validator.Validate(data); err != nil {
        return nil, err
    }
    // Thực thi business logic tuân thủ Single Responsibility...
    return &OutputModel{}, nil
}
```

## 3. Mẫu Khai Báo Unit Test (Table-Driven with Mock)

```go
// Tên file: [module_name]_test.go
package mymodule_test

import (
    "testing"
    "your_project_path/mymodule"
)

// fakeValidator triển khai ValidatorInterface dùng riêng cho test
type fakeValidator struct {
    errToReturn error
}

func (f *fakeValidator) Validate(data []byte) error {
    return f.errToReturn
}

func TestProcessData(t *testing.T) {
    tests := []struct {
        name        string
        mockErr     error
        inputData   []byte
        wantErr     bool
    }{
        {
            name: "Happy Path - Data Hợp Lệ",
            mockErr: nil,
            inputData: []byte("valid"),
            wantErr: false,
        },
        {
            name: "Fail Path - Validate Lỗi",
            mockErr: mymodule.ErrInvalidFormat,
            inputData: []byte("invalid"),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup Dependency
            mockVal := &fakeValidator{errToReturn: tt.mockErr}
            
            // Initiate Service
            svc := mymodule.NewConcreteService(mockVal)
            
            // Execute
            _, err := svc.ProcessData(tt.inputData)
            
            // Verify
            if (err != nil) != tt.wantErr {
                t.Errorf("expected error %v, got %v", tt.wantErr, err)
            }
        })
    }
}
```
