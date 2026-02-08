# Implementation Summary: pip Distribution for kagome-py

**Date**: February 8-9, 2026
**Status**: ✅ Complete and verified

---

## Overview

Implemented full pip distribution packaging for kagome-py with support for:
- **Wheels**: Pre-built binaries for 5 platform/arch combinations (macOS arm64+x86_64, Linux x86_64+arm64, Windows x86_64)
- **Sdist**: Source distribution with Go/C build hooks for source installs
- **CI/CD**: GitHub Actions workflows for automated testing and publishing to PyPI

---

## Directory Restructure

Reorganized the repository for clarity and maintainability:

```
BEFORE (mixed concerns):
  libkagome/
    ├── go_wrapper/
    ├── c_wrapper/
    ├── python_wrapper/
    ├── bin/ (build artifacts)
    └── dist/ (manual copy)

AFTER (clean separation):
  _src_c/ (native source, NOT installed)
    ├── go/ (Go FFI bridge)
    ├── c/ (C wrapper)
    └── build/ (artifacts)

  src/libkagome/ (Python package, pip installable)
    ├── __init__.py
    ├── _wrapper.py
    └── lib/ (shared libs placed here at build time)
```

**Rationale**: `src/` layout is the Python packaging standard. `_src_c/` keeps native source separate and internal-looking (underscore prefix signals "not for users").

---

## Files Created

### 1. **pyproject.toml**
- PEP 621 project metadata
- Declares setuptools + wheel as build system
- Package name: `kagome-py`, import: `from libkagome import Kagome`
- Minimum Python: 3.10+ (uses `match/case` syntax)
- `[tool.setuptools.package-data]`: includes `lib/*` in wheels

### 2. **setup.py** (minimal, ~155 lines)
Two responsibilities:

**a) Platform wheel tagging**:
- Custom `bdist_wheel` that sets `root_is_pure = False`
- Override `get_tag()` to return `("py3", "none", plat)` instead of `py3-none-any`
- Ensures wheels are platform-specific despite using ctypes (not a C extension)

**b) Source build hook**:
- Custom `build_py` that runs before package data collection
- Checks for Go and cc availability, raises clear error if missing
- Builds: `go build -buildmode=c-archive` → `cc -shared` → copies lib to `src/libkagome/lib/`
- Enables `pip install kagome-py --no-binary :all:` for users with Go installed

### 3. **MANIFEST.in**
Controls sdist contents:
- **Include**: Go/C source, Python package, build files
- **Exclude**: Build artifacts (`_src_c/build`), compiled libraries (`.so`, `.dylib`, `.dll`), cache
- Ensures sdist is self-contained but library-free (built on install)

### 4. **src/libkagome/__init__.py**
```python
from libkagome._wrapper import Kagome, Token
__all__ = ["Kagome", "Token"]
__version__ = "2.10.3"
```
Public API entry point.

### 5. **src/libkagome/_wrapper.py**
Adapted from `libkagome/python_wrapper/libkagome.py`:
- **Single change**: `_shared_library_path()` finds lib at `src/libkagome/lib/` (not `../bin/`)
- Everything else identical: ctypes structs, Token class, Kagome class
- ctypes FFI to Go-exported C functions

### 6. **CI/CD Workflows**

#### `.github/workflows/build-and-test.yml`
Runs on: `push` (main), `pull_request` (main), `workflow_dispatch`

**Matrix**: 5 platform/arch combinations
| Runner | GOOS | GOARCH | Output | Wheel Tag |
|--------|------|--------|--------|-----------|
| `macos-latest` | darwin | arm64 | .dylib | macosx_11_0_arm64 |
| `macos-13` | darwin | amd64 | .dylib | macosx_10_13_x86_64 |
| `ubuntu-latest` | linux | amd64 | .so | manylinux_2_17_x86_64 |
| `ubuntu-24.04-arm` | linux | arm64 | .so | manylinux_2_17_aarch64 |
| `windows-latest` | windows | amd64 | .dll | win_amd64 |

**Steps**:
1. Setup Go, Python
2. Build Go c-archive
3. Build shared lib (`cc -shared`)
4. Stage lib into `src/libkagome/lib/`
5. Build wheel with `pip wheel`
6. Re-tag with correct platform tag
7. Install wheel, run tests
8. Verify manylinux compliance (Linux only)
9. Upload artifact

#### `.github/workflows/publish.yml`
Runs on: `release` (published), `workflow_dispatch`

**Jobs**:
1. `build` - Same 5-platform matrix, builds wheels only (no tests)
2. `build-sdist` - Single job, builds sdist (no Go/C, just sources)
3. `publish` - Depends on both, publishes all artifacts to PyPI via trusted publisher (OIDC)

**Note**: Trusted publisher requires pre-configuration on PyPI (see "Next Steps").

---

## Files Modified

### 1. **Makefile**
- Updated paths: `libkagome/` → `_src_c/`, `.bin/` → `.build/`
- New targets:
  - `stage-lib`: Copy built lib to `src/libkagome/lib/`
  - `wheel`: Build wheel with `pip wheel`
  - `sdist`: Build source distribution
- Updated `clean`: Remove `_src_c/build/`, `src/libkagome/lib/`, `dist/`, etc.
- Updated `test-python`: Use `PYTHONPATH=./src`
- Updated `test-go`: Use `_src_c/go` path

### 2. **_src_c/c/kagome_wrapper.c**
- Include path: `#include "../build/libkagome.h"` (was `../bin/`)

### 3. **.gitignore**
Added:
```
_src_c/build/*
!_src_c/build/.gitkeep
src/libkagome/lib/libkagome.*
!src/libkagome/lib/.gitkeep
*.whl, *.tar.gz, *.egg-info, build/, dist/
```

### 4. **Dockerfile**
- Updated workdir: `/app/_src_c/go` (was `/app/libkagome/go_wrapper`)

### 5. **docker-compose.yml**
- Updated volume mounts to use `_src_c/go`, `_src_c/c`, `src/libkagome`

### 6. **README.md**
- Updated Python minimum version: 3.10+ (was 3.8+)

---

## Verification Results

### ✅ Wheel Build and Install
```bash
$ make clean && make build && make wheel
$ pip install dist/kagome_py-2.10.3-py3-none-macosx_26_0_arm64.whl
$ python -c "from libkagome import Kagome; print(Kagome().wakati('すもも...'))"
['すもも', 'も', 'もも', ...]
PASS
```

### ✅ Sdist Build and Install
```bash
$ make sdist
$ pip install dist/kagome_py-2.10.3.tar.gz  # Builds Go/C from source
$ python -c "from libkagome import Kagome; ..."
PASS
```

### ✅ Python Tests
All tests pass in both wheel and sdist install scenarios.

---

## Next Steps for PyPI Publishing

### 1. Register Project
Create account on [pypi.org](https://pypi.org), register project name `kagome-py`.

### 2. Configure Trusted Publisher (OIDC)
On PyPI, go to **Project Settings** → **Publishing** → **Add trusted publisher**:
- **GitHub repository owner**: KEINOS
- **Repository name**: kagome-py
- **Workflow name**: publish.yml
- **Environment name**: pypi

This allows the GitHub Actions `publish` job to authenticate without storing API tokens.

### 3. Create Release Trigger
```bash
git tag v2.10.3
git push origin v2.10.3
gh release create v2.10.3 --title "kagome-py 2.10.3"
```

The `publish.yml` workflow will automatically:
1. Build all wheels (5 platforms)
2. Build sdist
3. Publish to PyPI

### 4. Manual Test (Optional)
```bash
pip install kagome-py==2.10.3
from libkagome import Kagome
Kagome().wakati("すもも...")
```

---

## Architecture Decisions

### Why `src/` layout?
- Standard practice for Python packages
- Prevents accidental imports of uninstalled modules during development
- Clear separation: `src/` = what gets installed, root = build infrastructure

### Why `_src_c/` (not `native/` or `libkagome/`)?
- Underscore signals "internal build infrastructure, not for users"
- Avoids collision with the `src/libkagome/` package directory
- Clearer than generic `native/`

### Why both wheel and sdist?
- **Wheels**: Fast install, pre-built for each platform, recommended
- **Sdist**: Allows customization, builds from source with Go installed
- Users can choose: `pip install kagome-py` (wheel) or `pip install --no-binary :all: kagome-py` (source)

### Why `setup.py` + `pyproject.toml`?
- All metadata in `pyproject.toml` (PEP 621 standard)
- `setup.py` only provides build customizations (wheel tags, source build hook)
- Setuptools requires `setup.py` for command class overrides

### Why custom `build_py` (not `build_ext`)?
- `build_ext` runs AFTER `build_py`, too late for package_data collection
- `build_py` runs first, ensuring library is present when setuptools scans `lib/*`

---

## Known Limitations & Future Work

### Current (v1)
- Windows arm64: Deferred (would require cross-compiler setup)
- `-race` flag: Skipped in CI Go builds to avoid Go 1.25 linker warning (see initial issue analysis)

### For Future Releases
1. **Windows arm64**: Add when GitHub provides native runners or `llvm-mingw` cross-compilation is tested
2. **Automated version sync**: Script to extract version from `go.mod` and inject into Python files
3. **Release notes automation**: Generate changelog from commits
4. **PyPI badge**: Add to README after first publish

---

## Files Summary

| File | Type | Status |
|------|------|--------|
| `pyproject.toml` | NEW | ✅ |
| `setup.py` | NEW | ✅ |
| `MANIFEST.in` | NEW | ✅ |
| `src/libkagome/__init__.py` | NEW | ✅ |
| `src/libkagome/_wrapper.py` | NEW | ✅ |
| `.github/workflows/build-and-test.yml` | NEW | ✅ |
| `.github/workflows/publish.yml` | NEW | ✅ |
| `Makefile` | MODIFIED | ✅ |
| `.gitignore` | MODIFIED | ✅ |
| `_src_c/c/kagome_wrapper.c` | MODIFIED | ✅ |
| `Dockerfile` | MODIFIED | ✅ |
| `docker-compose.yml` | MODIFIED | ✅ |
| `README.md` | MODIFIED | ✅ |
| `_src_c/go/` | MOVED | ✅ |
| `_src_c/c/` | MOVED | ✅ |
| ~~`libkagome/`~~ | REMOVED | ✅ |

---

## Testing Checklist

- [x] `make clean && make build` — Go tests pass, shared lib built
- [x] `make test-python` — Python tests pass with `PYTHONPATH=./src`
- [x] `make wheel` — Wheel builds with platform tags
- [x] `make sdist` — Sdist builds with Go/C sources
- [x] Wheel install in clean venv — Works, tests pass
- [x] Sdist install in clean venv — Builds from source, tests pass
- [x] `auditwheel show` (Linux) — manylinux_2_17 compliance verified
- [x] No setuptools deprecation warnings (license format fixed)

---

## Questions or Issues?

Refer to the implementation plan: `/Users/keinos/.claude/plans/distributed-orbiting-waffle.md`

For PyPI publishing, contact [pypa/gh-action-pypi-publish](https://github.com/pypa/gh-action-pypi-publish).
