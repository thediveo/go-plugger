.PHONY: help chores clean coverage pkgsite report test vuln

help: ## list available targets
	@# Shamelessly stolen from Gomega's Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

clean: ## cleans up build and testing artefacts
	rm -f coverage.html coverage.out coverage.txt
	rm -f example/dynplug/dynplug.so

coverage: ## gathers coverage and updates README badge
	@scripts/cov.sh

pkgsite: ## serves Go documentation on port 6060
	@echo "navigate to: http://localhost:6060/github.com/thediveo/go-plugger/v3"
	@scripts/pkgsite.sh

report: ## run goreportcard on this module
	@scripts/goreportcard.sh

test: ## run unit tests
	@go build -race -tags plugger_dynamic -buildmode=plugin \
	    -o example/dynplug/dynplug.so ./example/dynplug
	@go test -v -race -p=1 -count=1 -tags plugger_dynamic ./...

vuln: ## runs govulncheck
	@scripts/vuln.sh

chores: ## updates Go binaries and NPM helper packages if necessary
	@scripts/chores.sh
	