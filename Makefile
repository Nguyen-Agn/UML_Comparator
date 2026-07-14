# Makefile
test:
	go test ./src/...

run:
	@echo "Sử dụng các lệnh sau để chạy thử nghiệm các module:"
	@echo "  make run_parse"
	@echo "  make run_build"
	@echo "  make run_prematch"
	@echo "  make run_match"
	@echo "  make run_compare"
	@echo "  make run_cipher"
	@echo "  make run_similarity"

run_parse:
	go run ./cmd/parse/main.go

run_build:
	go run ./cmd/build/main.go

run_prematch:
	go run ./cmd/prematch/main.go

run_match:
	go run ./cmd/match/main.go

run_compare:
	go run ./cmd/compare/main.go

run_cipher:
	go run ./cmd/cipher/main.go

run_similarity:
	go run ./cmd/similarity/main.go

open_suite:
	portable/instructor_suite.exe

open_student_gui:
	portable/student_uml.exe

open_student_cli:
	portable/student_uml_cli.exe

open_generator:
	portable/uml_generator.exe

open_editor:
	portable/mermaid_editor.exe

build:
	go build -ldflags="-H windowsgui" -o portable/instructor_suite.exe ./cmd/instructor/main.go
	go build -ldflags="-H windowsgui" -o portable/student_uml.exe ./gui/main.go
	go build -o portable/student_uml_cli.exe ./cmd/visualize/main.go ./cmd/visualize/interactive.go
	go build -ldflags="-H windowsgui" -o portable/mermaid_editor.exe ./cmd/mermaidEditor/...

build_linux:
	GOOS=linux go build -o portable/instructor_suite_linux ./cmd/instructor/main.go
	GOOS=linux go build -o portable/student_uml_linux ./gui/main.go
	GOOS=linux go build -o portable/mermaid_editor ./cmd/mermaidEditor/... 

build_all:
	go build -ldflags="-H windowsgui" -o portable/instructor_suite.exe ./cmd/instructor/main.go
	go build -ldflags="-H windowsgui" -o portable/student_uml.exe ./gui/main.go
	go build -o portable/student_uml_cli.exe ./cmd/visualize/main.go ./cmd/visualize/interactive.go
	go build -ldflags="-H windowsgui" -o portable/uml_generator.exe ./cmd/umlGenerator/...
	GOOS=linux go build -o portable/instructor_suite_linux ./cmd/instructor/main.go
	GOOS=linux go build -o portable/student_uml_linux ./gui/main.go
	GOOS=linux go build -o portable/uml_generator_linux ./cmd/umlGenerator/...

no_use:
	GOOS=linux go build -o portable/student_uml_cli_linux ./cmd/visualize/main.go ./cmd/visualize/interactive.go
