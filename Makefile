# Makefile
test:
	go test ./src/...

run:

build:
	go build -ldflags="-H windowsgui" -o portable\instructor_suite.exe .\cmd\instructor\main.go
	go build -ldflags="-H windowsgui" -o portable\student_uml.exe .\gui\main.go
	go build -o portable\student_uml_cli.exe .\cmd\visualize\main.go .\cmd\visualize\interactive.go
