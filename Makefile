build:
	go build -i -race -v -o ${GOPATH}/bin/gojob ./cmd/

exec:
	gojob

run:
	go run ./cmd/ -target ${opt}

run1: opt=1
run1: run

run2: opt=2
run2: run

run3: opt=3
run3: run

exec_indeed:
	gojob -target 1

exec_stackoverflow:
	gojob -target 2

exec_linkedin:
	gojob -target 3

exec_blockchain:
	gojob -key blockchain

#run: build exec
#	#go run -race ./cmd/main.go

#run1: build exec1
#	#go run -race ./cmd/main.go -target 1

#run2: build exec2
#	#go run -race ./cmd/main.go -target 2

#run3: build exec3
#	#go run -race ./cmd/main.go -target 3
