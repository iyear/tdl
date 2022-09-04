.PHONY: build
build:
	goreleaser build --rm-dist --single-target --snapshot
	@echo "go to '.tdl/dist' directory to see the package!"

.PHONY: packaging
packaging:
	goreleaser release --skip-publish --auto-snapshot --rm-dist
	@echo "go to '.tdl/dist' directory to see the packages!"
