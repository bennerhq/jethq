#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
bundle=${1:-"$root/jethq.app"}
temporary_dir=$(mktemp -d)
iconset="$temporary_dir/JetHQ.iconset"
mkdir "$iconset"
trap 'rm -rf "$temporary_dir"' EXIT

mkdir -p "$bundle/Contents/MacOS" "$bundle/Contents/Resources"
rm -f "$bundle/Contents/MacOS/jetkvm"
cp "$root/packaging/macos/Info.plist" "$bundle/Contents/Info.plist"

icon="$root/images/JetHQ-icon-bg.png"
sips -z 16 16 "$icon" --out "$iconset/icon_16x16.png" >/dev/null
sips -z 32 32 "$icon" --out "$iconset/icon_16x16@2x.png" >/dev/null
sips -z 32 32 "$icon" --out "$iconset/icon_32x32.png" >/dev/null
sips -z 64 64 "$icon" --out "$iconset/icon_32x32@2x.png" >/dev/null
sips -z 128 128 "$icon" --out "$iconset/icon_128x128.png" >/dev/null
sips -z 256 256 "$icon" --out "$iconset/icon_128x128@2x.png" >/dev/null
sips -z 256 256 "$icon" --out "$iconset/icon_256x256.png" >/dev/null
sips -z 512 512 "$icon" --out "$iconset/icon_256x256@2x.png" >/dev/null
sips -z 512 512 "$icon" --out "$iconset/icon_512x512.png" >/dev/null
sips -z 1024 1024 "$icon" --out "$iconset/icon_512x512@2x.png" >/dev/null
iconutil -c icns "$iconset" -o "$bundle/Contents/Resources/JetHQ.icns"

cd "$root"
go build -o "$bundle/Contents/MacOS/jethq" ./cmd/jetkvm-desktop
