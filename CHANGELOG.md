# Changelog

All notable changes to ThemeTime are documented here. The project follows
[Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.2] - 2026-07-22

### Fixed

- Prevented Smart Video Wallpaper Reborn from displaying a black desktop when
  its remembered video is no longer in ThemeTime's single-video playlist.
- Recreated stuck video wallpaper players during startup and manual apply so
  Plasma reliably loads and plays the selected video.

## [0.1.1] - 2026-07-22

### Added

- GoReleaser v2 release automation with native Arch Linux packages, checksums,
  SBOMs, and GitHub artifact attestations.

### Changed

- Consolidated the README, reference documentation, and task guides into one
  canonical documentation tree.

## [0.1.0] - 2026-07-20

### Added

- Solar Observatory desktop interface and system tray.
- Layered solar and fixed-time schedules for KDE Plasma 6 appearance actions.
- Static and Smart Video Wallpaper Reborn wallpaper support.
- Optional restricted privileged scheduler for SDDM and Plymouth themes.
- User service installer, diagnostics, configuration snapshots, and full docs.
- Reproducible Linux release archives and GitHub release automation.

[Unreleased]: https://github.com/RedUndercover/themetime/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/RedUndercover/themetime/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/RedUndercover/themetime/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/RedUndercover/themetime/releases/tag/v0.1.0
