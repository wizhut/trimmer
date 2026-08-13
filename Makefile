APP_NAME := trimmer
OUTPUT_DIR := bin
DIST_DIR := dist

# Single source of truth for the version is the `version` const in main.go.
VERSION := $(shell sed -n 's/^const version = "\(.*\)"/\1/p' main.go)

.PHONY: help build test vet install all clean release mac linux windows mac-intel mac-arm windows-amd64 windows-arm64 linux-amd64 linux-arm64

.DEFAULT_GOAL := help

help:
	@echo ""
	@echo "  \033[1m$(APP_NAME)\033[0m — Strips leading/trailing whitespace from stdin"
	@echo ""
	@echo "  \033[1;4mUsage\033[0m"
	@echo "    make \033[36m<target>\033[0m"
	@echo ""
	@echo "  \033[1;4mDevelopment\033[0m"
	@echo "    \033[36mbuild\033[0m        Build binary to $(OUTPUT_DIR)/"
	@echo "    \033[36mtest\033[0m         Run tests"
	@echo "    \033[36mvet\033[0m          Run go vet"
	@echo "    \033[36minstall\033[0m      Install to \$$GOPATH/bin"
	@echo "    \033[36mclean\033[0m        Remove build artifacts"
	@echo ""
	@echo "  \033[1;4mRelease\033[0m"
	@echo "    \033[36mall\033[0m          Cross-compile for every platform below"
	@echo "    \033[36mrelease\033[0m      Package cross-platform builds into $(DIST_DIR)/"
	@echo ""
	@echo "  \033[1;4mPlatforms\033[0m"
	@echo "    \033[36mmac\033[0m          darwin: amd64 + arm64"
	@echo "    \033[36mlinux\033[0m        linux: amd64 + arm64"
	@echo "    \033[36mwindows\033[0m      windows: amd64 + arm64"
	@echo ""
	@echo "  \033[1;4mSingle targets\033[0m"
	@echo "    \033[36mmac-intel\033[0m      darwin/amd64"
	@echo "    \033[36mmac-arm\033[0m        darwin/arm64"
	@echo "    \033[36mlinux-amd64\033[0m    linux/amd64"
	@echo "    \033[36mlinux-arm64\033[0m    linux/arm64"
	@echo "    \033[36mwindows-amd64\033[0m  windows/amd64"
	@echo "    \033[36mwindows-arm64\033[0m  windows/arm64"
	@echo ""
	@echo "  Provided by wizhut.tech — https://wizhut.tech"
	@echo ""

build:
	@mkdir -p $(OUTPUT_DIR)
	@echo "  \033[36m[build]\033[0m Compiling $(APP_NAME)..."
	@go build -o $(OUTPUT_DIR)/$(APP_NAME) .
	@echo "  \033[32m[build]\033[0m $(OUTPUT_DIR)/$(APP_NAME)"

vet:
	@echo "  \033[36m[vet]\033[0m Running go vet..."
	@go vet ./...
	@echo "  \033[32m[vet]\033[0m Clean"

install:
	@echo "  \033[36m[install]\033[0m Installing $(APP_NAME)..."
	@go install .
	@echo "  \033[32m[install]\033[0m Done"

test:
	@echo "  \033[36m[test]\033[0m Running tests..."
	@VERBOSE=$$(go test -v ./... 2>&1); RC=$$?; \
	SUMMARY=$$(echo "$$VERBOSE" | grep "^ok\|^FAIL"); \
	echo "$$SUMMARY" | while IFS= read -r line; do \
		case "$$line" in \
			ok*) echo "    \033[32m✅ $$line\033[0m" ;; \
			FAIL*) echo "    \033[31m❌ $$line\033[0m" ;; \
		esac; \
	done; \
	PKGS=$$(echo "$$SUMMARY" | grep -c "^ok" || true); \
	TESTS=$$(echo "$$VERBOSE" | grep -c "^--- PASS" || true); \
	echo ""; \
	if [ $$RC -eq 0 ]; then \
		echo "  \033[32m[test]\033[0m ✅ $$TESTS tests passed across $$PKGS packages"; \
	else \
		FTESTS=$$(echo "$$VERBOSE" | grep -c "^--- FAIL" || true); \
		echo "$$VERBOSE" | grep "^--- FAIL" | sed 's/^/    \033[31m❌ /;s/$$/\033[0m/'; \
		echo ""; \
		echo "  \033[31m[test]\033[0m ❌ $$FTESTS test(s) failed"; \
		exit 1; \
	fi

# ---------------------------------------------------------------------------
# Cross-compilation
#
# One recipe per GOOS/GOARCH pair, driven by the cross macro below:
#   $(call cross,<goos>,<goarch>,<binary suffix>)
# Binaries are named $(APP_NAME)-<goos>-<goarch>[suffix] inside $(OUTPUT_DIR).
# ---------------------------------------------------------------------------

UNIX_PLATFORMS := darwin-amd64 darwin-arm64 linux-amd64 linux-arm64
WIN_PLATFORMS  := windows-amd64 windows-arm64

define cross
	@echo "  \033[36m[build]\033[0m $(1)/$(2)..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(OUTPUT_DIR)/$(APP_NAME)-$(1)-$(2)$(3) .
	@echo "  \033[32m[build]\033[0m $(OUTPUT_DIR)/$(APP_NAME)-$(1)-$(2)$(3)"
endef

all: clean mac linux windows
	@echo ""
	@echo "  \033[32m[all]\033[0m Built $(words $(UNIX_PLATFORMS) $(WIN_PLATFORMS)) binaries in $(OUTPUT_DIR)/"

mac: mac-intel mac-arm
linux: linux-amd64 linux-arm64
windows: windows-amd64 windows-arm64

mac-intel:
	$(call cross,darwin,amd64,)

mac-arm:
	$(call cross,darwin,arm64,)

linux-amd64:
	$(call cross,linux,amd64,)

linux-arm64:
	$(call cross,linux,arm64,)

windows-amd64:
	$(call cross,windows,amd64,.exe)

windows-arm64:
	$(call cross,windows,arm64,.exe)

clean:
	@echo "  \033[36m[clean]\033[0m Removing build artifacts..."
	@rm -rf $(OUTPUT_DIR) $(DIST_DIR)
	@echo "  \033[32m[clean]\033[0m Done"

# Package each cross-build with the README and LICENSE: .tar.gz for unix,
# .zip for windows.
release: all
	@echo ""
	@echo "  \033[36m[release]\033[0m Packaging $(APP_NAME) $(VERSION)..."
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	@for p in $(UNIX_PLATFORMS); do \
		mkdir -p $(DIST_DIR)/$$p; \
		cp $(OUTPUT_DIR)/$(APP_NAME)-$$p $(DIST_DIR)/$$p/$(APP_NAME); \
		cp README.md LICENSE $(DIST_DIR)/$$p/; \
		tar -czf $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$$p.tar.gz -C $(DIST_DIR)/$$p .; \
		rm -rf $(DIST_DIR)/$$p; \
		echo "    \033[32m✅\033[0m $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$$p.tar.gz"; \
	done
	@for p in $(WIN_PLATFORMS); do \
		mkdir -p $(DIST_DIR)/$$p; \
		cp $(OUTPUT_DIR)/$(APP_NAME)-$$p.exe $(DIST_DIR)/$$p/$(APP_NAME).exe; \
		cp README.md LICENSE $(DIST_DIR)/$$p/; \
		zip -j -q $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$$p.zip $(DIST_DIR)/$$p/*; \
		rm -rf $(DIST_DIR)/$$p; \
		echo "    \033[32m✅\033[0m $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$$p.zip"; \
	done
	@echo ""
	@echo "  \033[32m[release]\033[0m Packages ready in $(DIST_DIR)/"
