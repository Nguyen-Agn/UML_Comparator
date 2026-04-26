# Makefile
test:
	go test ./src/...

run:

build:
	go build -ldflags="-H windowsgui" -o portable/instructor_suite.exe ./cmd/instructor/main.go
	go build -ldflags="-H windowsgui" -o portable/student_uml.exe ./gui/main.go
	go build -o portable/student_uml_cli.exe ./cmd/visualize/main.go ./cmd/visualize/interactive.go

build_linux:
	GOOS=linux go build -o portable/instructor_suite_linux ./cmd/instructor/main.go
	GOOS=linux go build -o portable/student_uml_linux ./gui/main.go

build_all:
	go build -ldflags="-H windowsgui" -o portable/instructor_suite.exe ./cmd/instructor/main.go
	go build -ldflags="-H windowsgui" -o portable/student_uml.exe ./gui/main.go
	go build -o portable/student_uml_cli.exe ./cmd/visualize/main.go ./cmd/visualize/interactive.go
	GOOS=linux go build -o portable/instructor_suite_linux ./cmd/instructor/main.go
	GOOS=linux go build -o portable/student_uml_linux ./gui/main.go

no_use:
	GOOS=linux go build -o portable/student_uml_cli_linux ./cmd/visualize/main.go ./cmd/visualize/interactive.go
