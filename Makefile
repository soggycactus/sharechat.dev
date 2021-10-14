.PHONY: clean test

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