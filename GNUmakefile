default: testacc

# Run acceptance tests
.PHONY: testacc docs
testacc:
	TF_ACC=1 go test ./internal/provider -v $(TESTARGS) -timeout 120m

govet:
	go vet ./...

build: govet
	go build .

replace:
	go mod edit -replace github.com/tabular-io/tabular-sdk-go=${shell dirname ${shell pwd}}/tabular-sdk-go

docs:
	tfplugindocs generate

clean:
	sed -i '' -e '/replace/d' go.mod

