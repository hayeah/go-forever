.PHONY: build
build:
	gox --os="darwin linux" --arch="386 amd64" ./forever