.PHONY:gen-mocks
gen-mocks:
	rm -frd
	mockery --config configs/.mockery.yml

.PHONY:gen-cli-docs
gen-cli-docs:
	go run docs/commands/main.go -docs-dir docs/commands

.PHONY:gen-mocks
build-and-check:
	go install .
	hange version

.PHONY:coverage
coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# human readable tests coverage by packages
.PHONY:coverage-filtered
coverage-filtered:
	@mod=$$(go list -m); \
	pkgs=$$(go list ./... | grep -v '/mocks' | grep -v '/configs'); \
	go test $$pkgs -coverprofile=coverage.out; \
	go tool cover -func=coverage.out

# shows pure percentage of tests coverage
.PHONY:coverage-percent
coverage-percent:
	@pkgs=$$(go list ./... | grep -v '/mocks' | grep -v '/configs' | grep -v '/hange/pkg/factory'); \
	go test $$pkgs -coverprofile=coverage.out >/dev/null
	@go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$$3); printf "%d\n", ($$3 + 0.5)}'
