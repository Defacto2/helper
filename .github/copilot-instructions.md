# Copilot Instructions for helper

## Quick Start

**Build & Test Commands:**
- Run all tests: `task test` (or `go test -count 1 ./...`)
- Run tests with race detection: `task testr` (or `go test -count 1 -race ./...`)
- Run a single test: `go test -run TestName -v`
- Lint code: `task lint` (uses gofumpt + golangci-lint)
- Check nil dereferences: `task nil` (uses `go tool nilaway`)

**Project Setup:**
- Go 1.25.1 or later required
- Uses `Task` runner (https://taskfile.dev/installation/) for common operations
- View all available tasks: `task --list-all`

## Architecture

This is a **utility library** providing helper functions for the Defacto2 server project. The package is organized into logical modules by functionality:

### Core Modules

1. **helper.go** (545 lines) - Main module containing:
   - Character encoding detection (`Determine`, `DetermineS`) - analyzes byte sequences to detect text encodings (Unicode, UTF-8, CP-437, CP-1252, etc.)
   - Date/time helpers (`TimeDistance`, `Day`, `Year`, `Latency`)
   - Network utilities (`Ping`, `LocalIPs`, `LocalHosts`, `LocalHostPing`)
   - Cryptography helpers (`CookieStore`)
   - Metadata (`Add1`, `Logger`)

2. **string.go** (780 lines) - String manipulation:
   - Cleaning and normalizing text
   - Slug generation
   - Truncation and padding
   - Case conversion utilities
   - UUID generation wrappers

3. **os.go** (538 lines) - File system operations:
   - File existence/type checking (`IsFile`, `IsDir`, `IsStat`)
   - File hashing (`Integrity`, `IntegrityFile`, `IntegrityBytes`)
   - Directory operations (`Count`, `Files`, `Lines`)
   - File matching and pattern searching (`FileMatch`, `Finds`)

4. **root.go** (148 lines) - Constrained file operations using `os.Root`:
   - Duplicate files with root constraints (`Duplicater`, `DuplicaterOW`)
   - Rename files (`RenameFile`, `RenameFileOW`)
   - Cross-device move detection (`RenameCrossDevice`)

### Platform-Specific Overrides

- **os_unix.go** (35 lines) - Unix/Linux implementations
- **os_win.go** (18 lines) - Windows-specific implementations

### Testing

- **helper_test.go** (331 lines) - Main test suite with embedded testdata
- **os_test.go** (265 lines) - File system operation tests
- **root_test.go** (124 lines) - Root operation tests
- **string_test.go** (696 lines) - String operation tests
- **testdata/** - Embedded test fixtures (includes TEST.BMP, TEST.DOC, and text samples)

Tests use Go's standard `embed` package to include testdata files directly in the binary.

## Key Conventions

### Character Encoding Detection
- `Determine()` is the public wrapper; `determine()` is the internal implementation with detailed logging support
- Detection prioritizes context: Unicode BOM detection → UTF-8 validation → character analysis → fallback to Windows-1252
- The `suppliment()` function checks for extended ASCII and special control sequences
- Amiga-specific handling for ANSI files (handles Alt-Esc and bell characters)

### Error Handling
- Functions that interact with the file system return `error` as the last return value
- Network operations (`Ping`, `LocalHostPing`) return status code and latency in nanoseconds
- Encoding detection functions return nil for unknown encodings (not an error case)

### Logging
- Functions with `S` suffix accept `*slog.Logger` for detailed logging (e.g., `DetermineS()`)
- Logger is passed as first parameter after `*slog.Logger`
- Available via context using `helper.Logger()` when needed

### File Modes & Permissions
- Use constants: `WriteWriteRead` (0o664 - RW for owner/group, R for others)
- Directory permissions: `DirWriteReadRead` (0o755)
- Avoid hardcoding octal values; use these constants instead

### Testing with testdata
- Testdata files are embedded in the test binary using `//go:embed testdata`
- Access embedded files via `testdataFS` (embed.FS) in tests
- This ensures tests work without external file dependencies

### Cross-Platform Considerations
- `os_unix.go` and `os_win.go` provide platform-specific implementations
- Use the pattern: check platform in function implementations or use build tags if complex
- The `RenameCrossDevice` function detects cross-device rename failures (Unix-specific issue)

### Code Style
- Use `golangci-lint` configuration (.golangci.yaml) with specific disabled linters
- Format with `gofumpt` (more opinionated than gofmt)
- Use `goimports` for import organization
- Nil dereference checking is enabled via `nilaway` - fix any tool errors

## Dependencies

- **golang.org/x/text** - Text encoding detection and Unicode handling
- **golang.org/x/sys** - Platform-specific syscalls
- **github.com/google/uuid** - UUID generation
- **github.com/nalgeon/be** - Bit manipulation utilities
- **go.uber.org/nilaway** - Nil dereference static analysis (dev tool)

## Common Pitfalls

- Don't hardcode file paths or permissions - use constants from `os.go`
- Encoding detection is probabilistic; don't assume 100% accuracy for ambiguous files
- Network operations have a 5-second timeout (`Timeout` constant)
- The `os.Root` constraint functions require careful use of the root handle; refer to Go 1.22+ documentation
