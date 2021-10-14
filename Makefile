.PHONY: clean test

build: 
	go build -ldflags="-s -w" -o bin/sharechat cmd/sharechat/main.go

clean:
	rm -rf ./bin Gopkg.lock *.out

int-test:
	@go test -race -v ./sharechat/... -tags=int

unit-test:
	@go test -race -v ./sharechat/... -tags=unit

e2e-test:
	@go test -v ./test/e2e/... -tags=e2e

unit-coverage:
	@go test -coverprofile=unit_coverage.out ./sharechat/... -coverpkg=./sharechat/... -tags=unit

view-coverage: unit-coverage
	@go tool cover -html=unit_coverage.out

lint:
	@golangci-lint run

migration:
	goose -dir=migrations create $(file) $(dialect)

goose-up:
	goose -dir=migrations postgres "user=user dbname=public password=password host=localhost sslmode=disable" up

goose-down:
	goose -dir=migrations postgres "user=user dbname=public password=password host=localhost sslmode=disable" down