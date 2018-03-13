BEAT_NAME=pprofbeat
ES_BEATS?=$(GOPATH)/src/github.com/elastic/beats
-include $(ES_BEATS)/libbeat/scripts/Makefile

.PHONY: all
all: update

.PHONY: collect
collect:
