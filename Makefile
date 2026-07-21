APP=themetime
VERSION?=$(shell tr -d '\n' < VERSION)
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || printf unknown)
GO_BUILD_FLAGS=-trimpath -buildvcs=false
VERSION_LDFLAGS=-s -w -X github.com/themetime/themetime/internal/buildinfo.Version=$(VERSION) -X github.com/themetime/themetime/internal/buildinfo.Commit=$(COMMIT)
STRIP_LDFLAGS=-s -w
WAILS_BASE_TAGS=production,desktop
WAILS_WEBKIT_TAG=$(shell pkg-config --exists webkit2gtk-4.1 2>/dev/null && printf ',webkit2_41')
WAILS_TAGS=$(WAILS_BASE_TAGS)$(WAILS_WEBKIT_TAG)

.PHONY: build build-wails-frontend clean docs-check test fmt doctor package release-check install-user-assets install-root-assets

build-wails-frontend:
	npm --prefix cmd/themetime-wails/frontend ci
	npm --prefix cmd/themetime-wails/frontend run build

build: build-wails-frontend
	go build $(GO_BUILD_FLAGS) -ldflags "$(VERSION_LDFLAGS)" -o bin/themetime ./cmd/themetime
	go build $(GO_BUILD_FLAGS) -ldflags "$(STRIP_LDFLAGS)" -tags "$(WAILS_TAGS)" -o bin/themetime-wails ./cmd/themetime-wails
	go build $(GO_BUILD_FLAGS) -ldflags "$(STRIP_LDFLAGS)" -o bin/themetime-rootctl ./cmd/themetime-rootctl
	go build $(GO_BUILD_FLAGS) -ldflags "$(STRIP_LDFLAGS)" -o bin/themetime-rootd ./cmd/themetime-rootd

clean:
	rm -rf -- bin dist cmd/themetime-wails/frontend/dist

docs-check:
	node scripts/check-docs.mjs

test: docs-check build-wails-frontend
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

doctor:
	go run ./cmd/themetime doctor

package: build
	./scripts/package-release.sh "$(VERSION)"

release-check: test package
	(cd dist && sha256sum --check ./*.sha256)

install-user-assets: build
	install -Dm755 bin/themetime "$(HOME)/.local/bin/themetime"
	install -Dm755 bin/themetime-wails "$(HOME)/.local/bin/themetime-wails"
	install -Dm644 assets/desktop/io.github.themetime.ThemeTime.desktop "$(HOME)/.local/share/applications/io.github.themetime.ThemeTime.desktop"
	install -Dm644 assets/icons/io.github.themetime.ThemeTime.svg "$(HOME)/.local/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg"
	-command -v update-desktop-database >/dev/null && update-desktop-database "$(HOME)/.local/share/applications"
	-command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -f -t "$(HOME)/.local/share/icons/hicolor"
	-command -v kbuildsycoca6 >/dev/null && kbuildsycoca6

install-root-assets:
	test -x bin/themetime-rootctl
	test -x bin/themetime-rootd
	install -Dm755 bin/themetime-rootctl /usr/local/libexec/themetime-rootctl
	install -Dm755 bin/themetime-rootd /usr/local/libexec/themetime-rootd
	install -Dm644 assets/polkit/io.github.themetime.rootctl.policy /usr/share/polkit-1/actions/io.github.themetime.rootctl.policy
	install -Dm644 assets/systemd/themetime-rootd.service /etc/systemd/system/themetime-rootd.service
	systemctl daemon-reload
