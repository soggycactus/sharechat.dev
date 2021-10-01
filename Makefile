.PHONY: clean test

clean:
	rm -rf ./bin Gopkg.lock *.out

unit-test:
	@go test -race -v ./... -tags=unit

unit-coverage:
	@go test -coverprofile=unit_coverage.out ./... -coverpkg=./... -tags=unit

view-coverage: unit-coverage
	@go tool cover -html=unit_coverage.out

lint:
	@golangci-lint run