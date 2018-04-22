
GOBIN = $(shell pwd)/build/bin
GO ?= latest

dacapp:
	build/env.sh go run build/ci.go install ./app
	@echo "Done building."
	@echo "Run \"$(GOBIN)/dacapp\" to launch dacChain."