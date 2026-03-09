#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

APP_NAME="ClaudeUsages"
APP_BUNDLE="${APP_NAME}.app"
BUILD_DIR=".build/release"

echo "==> Building ${APP_NAME} (release)..."
swift build -c release

echo "==> Assembling ${APP_BUNDLE}..."
rm -rf "${APP_BUNDLE}"
mkdir -p "${APP_BUNDLE}/Contents/MacOS"

cp "${BUILD_DIR}/${APP_NAME}" "${APP_BUNDLE}/Contents/MacOS/${APP_NAME}"
cp "Resources/Info.plist" "${APP_BUNDLE}/Contents/"

echo "==> Code signing (ad-hoc)..."
codesign --force --sign - "${APP_BUNDLE}"

echo "==> Done: ${SCRIPT_DIR}/${APP_BUNDLE}"
echo ""
echo "To install: cp -r ${APP_BUNDLE} /Applications/"
echo "To run:     open ${APP_BUNDLE}"
