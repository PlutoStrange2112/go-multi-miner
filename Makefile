test:
	go test -v -race ./...

multiminer-test:
	cd multiminer && go test -v -race

coverprofile:
	go test -coverprofile=coverage.out ./...

cover: coverprofile
	go tool cover -html=coverage.out

deps:
	go mod tidy
	go mod download

build:
	go build -o bin/multiminer ./cmd/multiminer

clean:
	rm -rf bin/

install:
	go install ./cmd/multiminer

.PHONY: test multiminer-test coverprofile cover deps build clean install