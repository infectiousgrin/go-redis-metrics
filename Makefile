#
# Project Makefile
# Assumes GOPATH is set correctly.
#

TESTFLAG ?= -cover
REDIS_SERVER ?= redis://

help:
	@echo
	@echo "  \033[34mdeps \033[0m - install dependencies"
	@echo "  \033[34mtest \033[0m - run the project tests"
	@echo "  \033[34mclean\033[0m - clean the project"
	@echo

deps:
	@go get -u github.com/tools/godep
	@go get -u github.com/garyburd/redigo/redis

test:
		REDIS_SERVER=$(REDIS_SERVER) \
		go test $(TESTFLAG) ./...

clean:
	@go clean

.PHONY: help deps test clean
