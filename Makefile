.PHONY: help dev-keys clean-dev reset-dev

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev-keys: .dev/github-app-key.pem ## Generate local dev secrets (idempotent)

.dev/github-app-key.pem:
	@mkdir -p .dev
	@echo "generating throwaway RSA key at $@ (used by GITHUB_APP_KEY in .envrc)"
	@openssl genrsa -out $@ 2048 2>/dev/null
	@echo "done. run 'direnv reload' to pick up the new key."

clean-dev: ## Wipe all local dev state (.dev/) — destroys postgres, clickhouse, temporal data
	@if pgrep -f process-compose >/dev/null 2>&1; then \
		echo "process-compose is running — stop it first with 'process-compose down'"; \
		exit 1; \
	fi
	rm -rf .dev

reset-dev: clean-dev dev-keys ## Wipe .dev/ and regenerate dev secrets — fresh slate
	@echo "clean slate ready. run 'process-compose up' to start fresh."
