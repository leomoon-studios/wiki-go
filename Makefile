# Use vendored dependencies
GOFLAGS=-mod=vendor

# build directory
BUILD_DIR = build

# get version from git
VERSION = $(shell git describe --tags --always 2>/dev/null || echo "dev")

LDFLAGS = "-X 'wiki-go/internal/version.Version=$(VERSION)' -s -w -extldflags=-static -linkmode 'external'"
BUILDTAGS = -tags netgo,usergo -trimpath -gcflags=all="-l -B -C"

all:
	@echo "Build this program for your architecture:"
	@echo " - make linux_amd64"
	@echo " - make linux_386"
	@echo " - make linux_arm64"
	@echo " - make linux_arm5"
	@echo " - make linux_arm6"
	@echo " - make linux_arm7"
	@echo " - make linux_s390x"
	@echo " - make windows_amd64"
	@echo " - make windows_arm64"
	@echo " - make macos_arm64"

prepare: 
	mkdir -p $(BUILD_DIR)
	find $(BUILD_DIR) -mindepth 1 -and -not -name ".gitkeep" -and -not -path $(BUILD_DIR)/data -and -not -path $(BUILD_DIR)/data/* -and -delete
	@echo "Building version: " $(VERSION)

linux_amd64: prepare
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-amd64 .

linux_386: prepare
	@echo "Building for Linux (386)..."
	GOOS=linux GOARCH=386 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-386 .

linux_arm64: prepare
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-arm64 .

linux_arm5: prepare
	@echo "Building for Linux ARMv5..."
	GOOS=linux GOARCH=arm GOARM=5 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-armv5 .

linux_arm6: prepare
	@echo "Building for Linux ARMv6..."
	GOOS=linux GOARCH=arm GOARM=6 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-armv6 .

linux_arm7: prepare
	@echo "Building for Linux ARMv7..."
	GOOS=linux GOARCH=arm GOARM=7 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-armv7 .

linux_s390x: prepare
	@echo "Building for Linux (s390x)..."
	GOOS=linux GOARCH=s390x go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-linux-s390x .

windows_amd64: prepare
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-windows-amd64.exe .

windows_arm64: prepare
	@echo "Building for Windows (arm64)..."
	GOOS=windows GOARCH=arm64 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-windows-arm64.exe .

macos_arm64: prepare
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build $(BUILDTAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/wiki-go-mac-amd64 .

