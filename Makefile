PWD=$(shell bash -c 'pwd')
BIN_NAME=local-persist

coverage: coverage-fix
	go tool cover -html=coverage.out

coverage-fix: test
	sed -i '' 's|'_$(PWD)'|.|g' coverage.out

test:
	go test -coverprofile=coverage.out

run:
	sudo -E go run main.go driver.go

docker-run:
	./scripts/docker-run.sh

release: binaries
	./scripts/release.sh

# build for current architecture
binary:
	go build -o bin/$(BIN_NAME) -v

# build all the binaries
binaries: clean-bin binary-linux-amd64 binary-linux-386 binary-linux-arm binary-freebsd-amd64 binary-freebsd-386

clean-bin:
	rm -Rf binary

# build a specific binary
binary-linux-amd64: export GOOS=linux
binary-linux-amd64: export GOARCH=amd64
binary-linux-amd64:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

binary-linux-386: export GOOS=linux
binary-linux-386: export GOARCH=386
binary-linux-386:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

binary-linux-arm: export GOOS=linux
binary-linux-arm: export GOARCH=arm
binary-linux-arm:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

binary-freebsd-amd64: export GOOS=freebsd
binary-freebsd-amd64: export GOARCH=amd64
binary-freebsd-amd64:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

binary-freebsd-386: export GOOS=freebsd
binary-freebsd-386: export GOARCH=386
binary-freebsd-386:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v
