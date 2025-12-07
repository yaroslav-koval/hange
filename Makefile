.PHONY:gen-mocks
gen-mocks:
	mockery --config configs/.mockery.yml
