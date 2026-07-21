#!/usr/bin/env bash
set -euo pipefail

repo_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
version="${1:-$(tr -d '\n' < "$repo_dir/VERSION")}"
go_arch="${GOARCH:-$(go env GOARCH)}"

case "$version" in
  v*) version="${version#v}" ;;
esac

if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  printf 'error: version must resemble 1.2.3 or 1.2.3-rc.1 (got %q)\n' "$version" >&2
  exit 2
fi

case "$go_arch" in
  amd64) release_arch="x86_64" ;;
  arm64) release_arch="aarch64" ;;
  *) release_arch="$go_arch" ;;
esac

for binary in themetime themetime-wails themetime-rootctl themetime-rootd; do
  if [[ ! -x "$repo_dir/bin/$binary" ]]; then
    printf 'error: bin/%s is missing; run make build first\n' "$binary" >&2
    exit 1
  fi
done

package_name="themetime-${version}-linux-${release_arch}"
dist_dir="$repo_dir/dist"
stage_dir="$dist_dir/$package_name"
archive="$dist_dir/$package_name.tar.gz"

rm -rf -- "$stage_dir"
mkdir -p -- \
  "$stage_dir/bin" \
  "$stage_dir/libexec" \
  "$stage_dir/lib/systemd/system" \
  "$stage_dir/share/applications" \
  "$stage_dir/share/icons/hicolor/scalable/apps" \
  "$stage_dir/share/polkit-1/actions" \
  "$stage_dir/docs" \
  "$stage_dir/wiki"

install -m755 "$repo_dir/bin/themetime" "$stage_dir/bin/themetime"
install -m755 "$repo_dir/bin/themetime-wails" "$stage_dir/bin/themetime-wails"
install -m755 "$repo_dir/bin/themetime-rootctl" "$stage_dir/libexec/themetime-rootctl"
install -m755 "$repo_dir/bin/themetime-rootd" "$stage_dir/libexec/themetime-rootd"
install -m644 "$repo_dir/assets/desktop/io.github.themetime.ThemeTime.desktop" "$stage_dir/share/applications/"
install -m644 "$repo_dir/assets/icons/io.github.themetime.ThemeTime.svg" "$stage_dir/share/icons/hicolor/scalable/apps/"
install -m644 "$repo_dir/assets/polkit/io.github.themetime.rootctl.policy" "$stage_dir/share/polkit-1/actions/"
install -m644 "$repo_dir/assets/systemd/themetime-rootd.service" "$stage_dir/lib/systemd/system/"
install -m644 "$repo_dir/README.md" "$repo_dir/LICENSE" "$repo_dir/CHANGELOG.md" "$stage_dir/"
install -m755 "$repo_dir/scripts/install-release.sh" "$stage_dir/install.sh"
install -m755 "$repo_dir/scripts/uninstall-release.sh" "$stage_dir/uninstall.sh"
cp -a -- "$repo_dir/docs/." "$stage_dir/docs/"
cp -a -- "$repo_dir/wiki/." "$stage_dir/wiki/"

printf '%s\n' "$version" > "$stage_dir/VERSION"

epoch="${SOURCE_DATE_EPOCH:-0}"
tar --sort=name --mtime="@$epoch" --owner=0 --group=0 --numeric-owner \
  -C "$dist_dir" -cf - "$package_name" | gzip -n > "$archive"
(
  cd -- "$dist_dir"
  sha256sum "$(basename -- "$archive")"
) > "$archive.sha256.tmp"
mv -- "$archive.sha256.tmp" "$archive.sha256"

printf 'Created %s\n' "$archive"
printf 'Created %s\n' "$archive.sha256"
