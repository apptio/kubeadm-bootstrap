clean:
	@rm -rf ./dist

build: clean
	@goreleaser --skip-publish
