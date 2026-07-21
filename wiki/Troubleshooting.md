# Troubleshooting

## Three-command triage

```sh
themetime doctor
themetime show-config
systemctl --user status themetime.service
```

Then test the exact phase:

```sh
themetime apply --phase <phase-id>
```

The JSON result identifies the failed or skipped action.

## Quick symptom map

| Symptom | First check |
| --- | --- |
| GUI not found | Keep `themetime-wails` beside `themetime` or set `THEMETIME_GUI`. |
| Blank/failed GUI | Run `themetime-wails` in a terminal; check WebKitGTK and rebuild frontend. |
| Window seems impossible to close | Use **Quit** from the system tray. |
| Wrong solar time | Verify coordinate signs and IANA timezone together. |
| Service runs but makes no changes | Inspect `journalctl --user -u themetime.service`. |
| Phase repeats every poll | One action is probably failing, so state is not finalized. |
| Theme ID is rejected | Use the installed package ID rather than its display name. |
| Image is not found | Use an absolute path or `~/...`; `$HOME` is not expanded. |
| Video does not play | Check exact Smart Video plugin ID and `qdbus6`. |
| Command works only in terminal | Use absolute executables and avoid interactive shell assumptions. |
| SDDM/Plymouth does nothing | Re-export, then inspect the root service journal. |

## Logs

User scheduler:

```sh
journalctl --user -u themetime.service -n 100 --no-pager
```

Privileged scheduler:

```sh
sudo journalctl -u themetime-rootd.service -n 100 --no-pager
```

## Configuration recovery

Stop the daemon before repairing JSON:

```sh
systemctl --user stop themetime.service
themetime show-config
```

Restore your manual backup or correct the reported validation error. ThemeTime
also keeps snapshots of selected KDE files under
`~/.local/state/themetime/snapshots/`, but it does not automatically restore
them.

## Detailed playbook

The comprehensive [troubleshooting reference](../docs/troubleshooting.md) covers
each symptom, expected KDE caching behavior, service environment differences,
wallpaper screen targeting, plugin discovery, privileged state, and safe bug
report collection.
