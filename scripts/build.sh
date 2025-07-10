#!/usr/bin/env bash
set -euo pipefail

# Determine repository root relative to this script.
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
ROOT_DIR="${SCRIPT_DIR}/.."

echo "[SIMBA build] Building Rust static and shared libraries..."
(
  cd "${ROOT_DIR}/rust"
  # Build with nightly to enable portable_simd feature
  if command -v rustup >/dev/null 2>&1; then
    cargo +nightly build --release --lib
  else
    echo "[WARN] rustup not found; attempting to build with default cargo (requires nightly default)" >&2
    cargo build --release --lib
  fi
)

# Remove any outdated generic syso archive and legacy cgo libraries
rm -f "${ROOT_DIR}/internal/ffi/libsimba.syso" \
       "${ROOT_DIR}/internal/ffi/libsimba.a" \
       "${ROOT_DIR}/internal/ffi/libsimba.dylib"

# ---------------------------------------------------------------------------
# Build .syso archives for the no-cgo backend (darwin/amd64 and darwin/arm64)
# ---------------------------------------------------------------------------
echo "[SIMBA build] Building Mach-O .syso archives for simba_syso tag..."
(
  cd "${ROOT_DIR}/rust"
  if command -v rustup >/dev/null 2>&1; then
    for target in x86_64-apple-darwin aarch64-apple-darwin; do
      rustup target add ${target} --toolchain nightly >/dev/null 2>&1 || true
      cargo +nightly rustc --release --lib --target ${target} -- -C relocation-model=pic
      arch=${target%%-*}; os=${target##*-darwin}; # arch part first segment until '-' maybe x86_64 or aarch64
      goarch=$( [ "$arch" = "x86_64" ] && echo amd64 || echo arm64 )
      cp "${ROOT_DIR}/rust/target/${target}/release/libsimba.a" \
         "${ROOT_DIR}/internal/ffi/libsimba_darwin_${goarch}.syso" || true
    done
  else
    echo "[WARN] rustup not found; skipping .syso build" >&2
  fi
)

echo "[SIMBA build] syso builds complete."

echo "[SIMBA build] Done." 