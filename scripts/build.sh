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
    cargo +nightly build --release --lib --crate-type staticlib --crate-type cdylib
  else
    echo "[WARN] rustup not found; attempting to build with default cargo (requires nightly default)" >&2
    cargo build --release
  fi
)

echo "[SIMBA build] Copying library into internal/ffi..."
cp "${ROOT_DIR}/rust/target/release/libsimba.a" "${ROOT_DIR}/internal/ffi/" || true
cp "${ROOT_DIR}/rust/target/release/libsimba.dylib" "${ROOT_DIR}/internal/ffi/" || true
cp "${ROOT_DIR}/rust/target/release/libsimba.so" "${ROOT_DIR}/internal/ffi/" || true

echo "[SIMBA build] Done." 