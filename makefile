SHELL := /bin/bash
include .env

exceed-headers:
	curl -v $(HOST):$(PORT) -H "X-Big-Header: $$(printf 'A%.0s' {1..9000})"
