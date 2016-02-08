PWD=$(shell bash -c 'pwd')

coverage: coverage-fix
	go tool cover -html=coverage.out

coverage-fix: test
	sed -i '' 's|'_$(PWD)'|.|g' coverage.out

test:
	go test -coverprofile=coverage.out

run:
	sudo -E go run main.go driver.go

docker-run:
	./hack/docker-run.sh

