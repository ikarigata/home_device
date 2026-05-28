GOOS  ?= linux
GOARCH ?= arm64
DIST  := dist

.PHONY: build trigger image clean

build: trigger image

trigger:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -tags lambda.norpc -o $(DIST)/trigger/bootstrap ./cmd/trigger
	cd $(DIST)/trigger && zip -q -j ../trigger.zip bootstrap

image:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -tags lambda.norpc -o $(DIST)/image/bootstrap ./cmd/image
	cd $(DIST)/image && zip -q -j ../image.zip bootstrap

clean:
	rm -rf $(DIST)
