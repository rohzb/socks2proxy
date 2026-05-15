# socks2proxy build and development tasks.
BINARY := socks2proxy
CMD_PATH := ./cmd/socks2proxy
BUILD_DIR := .
DIST_DIR := dist
PKG_DIR := release
# VERSION is normalized to a semver-like string:
# - exact tag vX.Y.Z           -> X.Y.Z
# - exact tag with dirty tree   -> X.Y.Z-dirty
# - commits after tag          -> X.Y.Z-dev.N+gSHA[.dirty]
# - no matching semver tag     -> 0.0.0-dev+gSHA[.dirty]
VERSION ?= $(shell sh -c '\
d=$$(git describe --tags --dirty --always --match "v[0-9]*.[0-9]*.[0-9]*" 2>/dev/null || true); \
c=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown); \
if printf "%s" "$$d" | grep -Eq "^v[0-9]+\\.[0-9]+\\.[0-9]+$$"; then \
  printf "%s" "$${d#v}"; \
elif printf "%s" "$$d" | grep -Eq "^v[0-9]+\\.[0-9]+\\.[0-9]+-dirty$$"; then \
  printf "%s" "$${d#v}"; \
elif printf "%s" "$$d" | grep -Eq "^v[0-9]+\\.[0-9]+\\.[0-9]+-[0-9]+-g[0-9a-f]+(-dirty)?$$"; then \
  base=$$(printf "%s" "$$d" | sed -E "s/^v([0-9]+\\.[0-9]+\\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$$/\\1/"); \
  cnt=$$(printf "%s" "$$d" | sed -E "s/^v([0-9]+\\.[0-9]+\\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$$/\\2/"); \
  sha=$$(printf "%s" "$$d" | sed -E "s/^v([0-9]+\\.[0-9]+\\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$$/\\3/"); \
  dirty=$$(printf "%s" "$$d" | sed -E "s/^v([0-9]+\\.[0-9]+\\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$$/\\4/"); \
  if [ "$$dirty" = "-dirty" ]; then printf "%s-dev.%s+g%s.dirty" "$$base" "$$cnt" "$$sha"; else printf "%s-dev.%s+g%s" "$$base" "$$cnt" "$$sha"; fi; \
else \
  if printf "%s" "$$d" | grep -Eq "dirty$$"; then printf "0.0.0-dev+g%s.dirty" "$$c"; else printf "0.0.0-dev+g%s" "$$c"; fi; \
fi')
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
AUTHOR ?= $(shell git config user.name 2>/dev/null || echo Ruslan\ Ovsyannikov)
LICENSE ?= MIT
GOFLAGS ?= -trimpath
LDFLAGS := -s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)' -X 'main.author=$(AUTHOR)' -X 'main.license=$(LICENSE)'
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: all build build-linux build-all package-release checksums release-manifest clean fmt go-fmt test vet check version

all: build

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)

build-all:
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		goos=$${platform%/*}; \
		goarch=$${platform#*/}; \
		out="$(DIST_DIR)/$(BINARY)-$(VERSION)-$${goos}-$${goarch}"; \
		if [ "$$goos" = "windows" ]; then out="$$out.exe"; fi; \
		echo "building $$out"; \
		CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o "$$out" $(CMD_PATH); \
	done

package-release: build-all
	@rm -rf $(PKG_DIR)
	@mkdir -p $(PKG_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		goos=$${platform%/*}; \
		goarch=$${platform#*/}; \
		tmpdir=$$(mktemp -d); \
		pkgroot="$$tmpdir/socks2proxy"; \
		mkdir -p "$$pkgroot/docs" "$$pkgroot/examples" "$$pkgroot/platform"; \
		cp README.md LICENSE "$$pkgroot/"; \
		cp docs/*.md "$$pkgroot/docs/"; \
		cp examples/config.example.yaml "$$pkgroot/examples/"; \
		cp -R "platform/$$goos" "$$pkgroot/platform/"; \
		bin_src="$(DIST_DIR)/$(BINARY)-$(VERSION)-$${goos}-$${goarch}"; \
		bin_dst="$$pkgroot/$(BINARY)"; \
		if [ "$$goos" = "windows" ]; then \
			bin_src="$$bin_src.exe"; \
			bin_dst="$$pkgroot/$(BINARY).exe"; \
		fi; \
		cp "$$bin_src" "$$bin_dst"; \
		if [ "$$goos" = "windows" ]; then \
			archive="$(PKG_DIR)/$(BINARY)_$(VERSION)_$${goos}_$${goarch}.zip"; \
			( cd "$$tmpdir" && zip -qr "$$(pwd)/../$$(basename "$$archive")" socks2proxy ); \
			mv "$$tmpdir/../$$(basename "$$archive")" "$$archive"; \
		else \
			archive="$(PKG_DIR)/$(BINARY)_$(VERSION)_$${goos}_$${goarch}.tar.gz"; \
			tar -C "$$tmpdir" -czf "$$archive" socks2proxy; \
		fi; \
		rm -rf "$$tmpdir"; \
		echo "packaged $$archive"; \
	done

checksums: package-release
	@cd $(PKG_DIR) && sha256sum $(BINARY)_$(VERSION)_* > SHA256SUMS
	@echo "wrote $(PKG_DIR)/SHA256SUMS"

release-manifest: checksums
	@set -e; \
	manifest="$(PKG_DIR)/release-manifest.json"; \
	printf '{\n  "project": "%s",\n  "version": "%s",\n  "commit": "%s",\n  "date": "%s",\n  "artifacts": [\n' "$(BINARY)" "$(VERSION)" "$(COMMIT)" "$(DATE)" > "$$manifest"; \
	first=1; \
	while read -r sum file; do \
		[ "$$file" = "SHA256SUMS" ] && continue; \
		size=$$(wc -c < "$(PKG_DIR)/$$file"); \
		if [ $$first -eq 0 ]; then printf ',\n' >> "$$manifest"; fi; \
		first=0; \
		printf '    {"name":"%s","sha256":"%s","size_bytes":%s}' "$$file" "$$sum" "$$size" >> "$$manifest"; \
	done < "$(PKG_DIR)/SHA256SUMS"; \
	printf '\n  ]\n}\n' >> "$$manifest"; \
	echo "wrote $$manifest"

version:
	@echo VERSION=$(VERSION)
	@echo COMMIT=$(COMMIT)
	@echo DATE=$(DATE)
	@echo AUTHOR=$(AUTHOR)
	@echo LICENSE=$(LICENSE)

fmt:
	gofmt -w ./cmd ./internal

go-fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

check: fmt go-fmt vet test

clean:
	rm -f $(BUILD_DIR)/$(BINARY)
	rm -rf $(DIST_DIR)
	rm -rf $(PKG_DIR)
