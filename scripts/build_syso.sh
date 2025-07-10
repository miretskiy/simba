#!/usr/bin/env bash
set -euo pipefail

# Build libsimba static archives for both macOS targets and copy them as .syso
# so that the Go linker picks them up automatically.

readonly TOOLCHAIN="${1:-nightly}"
readonly MANIFEST="$(dirname "$0")/../rust/Cargo.toml"

for target in x86_64-apple-darwin aarch64-apple-darwin; do
  rustup target add "$target" --toolchain "$TOOLCHAIN" >/dev/null 2>&1 || true
  cargo +"$TOOLCHAIN" rustc --manifest-path "$MANIFEST" --release --lib --target "$target" -- -C relocation-model=pic
  if [[ "$target" == x86_64-apple-darwin ]]; then
    goarch=amd64
  else
    goarch=arm64
  fi
  cp "$(dirname "$MANIFEST")/target/$target/release/libsimba.a" "$(dirname "$0")/../internal/ffi/libsimba_darwin_${goarch}.syso"
  echo "Generated libsimba_darwin_${goarch}.syso"
done 