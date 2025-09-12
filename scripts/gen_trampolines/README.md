# gen_trampolines

Automatically generates assembly trampolines for SIMBA FFI functions by scanning Go source files for `//simba:trampoline` annotations.

## Usage

The script can now be invoked from any directory within the SIMBA project:

```bash
# From project root (recommended)
go run ./scripts/gen_trampolines

# From anywhere within the project
go run path/to/scripts/gen_trampolines

# With explicit paths
go run ./scripts/gen_trampolines -pkg ./internal/ffi -output ./internal/ffi
```

## Command Line Options

- `-pkg string`: Path to package containing trampoline declarations (default: auto-detect)
- `-output string`: Directory to write generated .s files (default: same as pkg)

## How It Works

1. **Auto-detection**: If no `-pkg` is specified, the script walks up from the current directory to find `go.mod`, then looks for `internal/ffi/` relative to the project root.

2. **Source scanning**: Parses all `.go` files in the target package looking for functions with `//simba:trampoline` comments.

3. **Assembly generation**: For each annotated function, generates platform-specific assembly trampolines:
   - `syso_amd64.s` - x86-64 assembly
   - `syso_arm64.s` - ARM64 assembly

4. **Output**: Files are written to the output directory (defaults to the same directory as the source package).

## Integration with go generate

The script is designed to work seamlessly with `go generate`:

```go
//go:generate go run ../../scripts/gen_trampolines
```

This can be placed in any Go file within the project, and the script will automatically find the correct paths regardless of where `go generate` is invoked from.

## Directory Structure Requirements

The script expects a standard Go module layout:
- `go.mod` at project root
- FFI package at `internal/ffi/`
- Generated files written to the same directory as the source

## Supported Function Signatures

The generator supports these parameter and return types:
- `*byte` (pointer to data)
- `uintptr` (sizes, offsets)
- `uint8`, `uint32`, `uint64` (scalar values)
- `float32`, `float64` (floating point)
- `*uint16`, `*uint32`, `*uint64` (output pointers)

## Example Trampoline Declaration

```go
//simba:trampoline amd64 arm64
//go:noescape
func sum_u8_32_raw(ptr *byte, n uintptr) uint32
```

This generates trampolines for both amd64 and arm64 architectures that:
1. Load parameters from Go ABI stack frame
2. Move parameters to System-V/AAPCS64 calling convention
3. Call the Rust function `sum_u8_32` (without `_raw` suffix)
4. Move return value back to Go ABI location
