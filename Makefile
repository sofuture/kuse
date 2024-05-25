.PHONY: default
default: build

.PHONY: build
build: build/kuse

.PHONY: clean
clean:
	rm -rf build

build/kuse: cmd/kuse/main.go
	go build -o build/kuse cmd/kuse/main.go