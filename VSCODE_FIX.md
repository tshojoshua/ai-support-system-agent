# VS Code File Association Fix

## Problem
VS Code incorrectly identifies Debian maintainer scripts (`postinst`, `prerm`, `postrm`) and macOS installer scripts (`preinstall`, `postinstall`) as Dockerfiles, showing false "compile errors."

## Solution
A `.vscode/settings.json` file has been created with the correct file associations. Since `.vscode/` is gitignored, you need to keep this file locally.

The file is located at: `.vscode/settings.json`

## Manual Fix (if needed)
If the file still shows as red/Dockerfile:

1. Click on the language mode in the bottom-right of VS Code (it might say "Dockerfile")
2. Select "Configure File Association for 'postinst'..."
3. Choose "Shell Script"

Or use Command Palette (Ctrl+Shift+P):
- Type: "Change Language Mode"
- Select "Shell Script (Bash)"

## Verification
After fixing the association, the Dockerfile errors will disappear and you'll get proper bash syntax highlighting instead.

## Files Affected
- `packaging/linux/debian/postinst`
- `packaging/linux/debian/prerm`
- `packaging/linux/debian/postrm`
- `packaging/macos/scripts/preinstall`
- `packaging/macos/scripts/postinstall`

All these files are valid bash scripts with correct shebangs (`#!/bin/bash`) and executable permissions.
