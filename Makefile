build:
	go build -i -v -o ${GOPATH}/bin/gojob ./cmd/

exec:
	gojob

exec1:
	gojob -target 1

exec2:
	gojob -target 2

exec3:
	gojob -target 3

run: build exec
	#go run ./cmd/main.go

run1: build exec1
	#go run ./cmd/main.go -target 1

run2: build exec2
	#go run ./cmd/main.go -target 2

run3: build exec3
	#go run ./cmd/main.go -target 3
