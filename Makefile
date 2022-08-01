PWD=$(shell bash -c 'pwd')
BIN_NAME=local-persist

coverage:
	GO_ENV=test go test -v -coverprofile=coverage.out ./... && sed -i '' 's|'_$(PWD)'|.|g' coverage.out && go tool cover -html=coverage.out

test: export GO15VENDOREXPERIMENT=1
test: export GO_ENV=test
test:
	go test -v .

run:
	sudo -E go run main.go driver.go

docker-run:
	./scripts/docker-run.sh

docker-build:
docker-build: clean-bin
	./scripts/docker-build.sh

release:
release: docker-build
	./scripts/release.sh

# build for current architecture
binary: export GO15VENDOREXPERIMENT=1
binary:
	go build -o bin/$(BIN_NAME) -v

# build all the binaries
binaries: clean-bin binary-linux-amd64 binary-linux-arm64 # binary-linux-386 binary-linux-arm binary-freebsd-amd64 binary-freebsd-386

clean-bin:
	rm -Rf bin/*

# build a specific binary
binary-linux-amd64: export GOOS=linux
binary-linux-amd64: export GOARCH=amd64
binary-linux-amd64: export GO15VENDOREXPERIMENT=1
binary-linux-amd64:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v


binary-linux-arm64: export GOOS=linux
binary-linux-arm64: export GOARCH=arm64
binary-linux-arm64: export GO15VENDOREXPERIMENT=1
binary-linux-arm64:
	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

# docker doesn't currently support x32 architecture
# binary-linux-386: export GOOS=linux
# binary-linux-386: export GOARCH=386
# binary-linux-386: export GO15VENDOREXPERIMENT=1
# binary-linux-386:
# 	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

# docker doesn't currently support arm architecture
# binary-linux-arm: export GOOS=linux
# binary-linux-arm: export GOARCH=arm
# binary-linux-arm: export GO15VENDOREXPERIMENT=1
# binary-linux-arm:
# 	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

# cowardly unwilling to support other architectures for now
# binary-freebsd-amd64: export GOOS=freebsd
# binary-freebsd-amd64: export GOARCH=amd64
# binary-freebsd-amd64: export GO15VENDOREXPERIMENT=1
# binary-freebsd-amd64:
# 	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v

# docker doesn't currently support x32 architecture
# binary-freebsd-386: export GOOS=freebsd
# binary-freebsd-386: export GOARCH=386
# binary-freebsd-386: export GO15VENDOREXPERIMENT=1
# binary-freebsd-386:
# 	go build -o bin/$(GOOS)/$(GOARCH)/$(BIN_NAME) -v
