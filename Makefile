.PHONY: help clean coverage pkgsite reportcard test

help: ## list available targets
	@# Shamelessly stolen from Gomega's Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f coverage.html coverage.out coverage.txt
	rm -f test/dynamicplugintesting/dynfooplug.so

coverage: ## gathers coverage and updates README badge
	@scripts/cov.sh

pkgsite: ## serves Go documentation on port 6060
	@echo "navigate to: http://localhost:6060/github.com/thediveo/go-plugger/v2"
	@scripts/pkgsite.sh

report: ## run goreportcard on this module
	@scripts/goreportcard.sh

test: ## run unit tests
	@go build -tags plugger_dynamic,dynamicplugintesting -buildmode=plugin \
	    -o test/dynamicplugintesting/dynfoo/dynfooplug.so \
    	./test/dynamicplugintesting/dynfoo
	@go test -v -p=1 -count=1 -tags plugger_dynamic,dynamicplugintesting ./...
