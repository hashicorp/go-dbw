


tools:
	go generate -tags tools tools/tools.go


fmt:
	gofumpt -w $$(find . -name '*.go' | grep -v pb.go)