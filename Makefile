SRC_ROOT := $(shell pwd)
PB_ROOT := $(SRC_ROOT)/pb


CACHE_PB_ROOT := $(PB_ROOT)/cachepb
cachepb: $(CACHE_PB_ROOT)/cache.proto
	protoc -I=$(SRC_ROOT) --go_out=$(SRC_ROOT) --go-grpc_out=$(SRC_ROOT) $<

.PHONY:
clean:
	rm -rf $(CACHE_PB_ROOT)/*.pb.go
