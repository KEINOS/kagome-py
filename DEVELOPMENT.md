# Development Guide: kagome-py

## Installation

### For Users

Install from PyPI:

```bash
pip install kagome-py
```

### Build from Source

Requires: Go 1.25+, C compiler (cc/gcc)

```bash
pip install kagome-py --no-binary :all:
```

### Usage

```python
from libkagome import Kagome

kagome = Kagome()

# Full morphological analysis
tokens = kagome.tokenize("すもももももももものうち")
for token in tokens:
    print(token)

# Word segmentation only
words = kagome.wakati("すもももももももものうち")
print(words)  # ['すもも', 'も', 'もも', 'も', 'もも', 'の', 'うち']
```

## For Developers

### Local Development

**Setup**:

```bash
git clone https://github.com/KEINOS/kagome-py.git
cd kagome-py
```

**Build and test**:

```bash
make clean           # Clean build artifacts
make build           # Build shared library for current platform
make test-python     # Test Python bindings
make test-go         # Run Go tests (also in Makefile test-python)
```

**Build wheels locally** (current platform only):

```bash
make wheel           # Creates dist/kagome_py-*.whl
pip install dist/*.whl  # Test install
```

**Build source distribution**:

```bash
make sdist           # Creates dist/kagome_py-*.tar.gz
```

### Release to PyPI

**Prerequisites**:

1. Configure GitHub → PyPI trusted publisher (OIDC):
   - Go to [pypi.org](https://pypi.org) → Project settings → Publishing
     - Add trusted publisher: GitHub repo `KEINOS/kagome-py`,
         workflow `publish.yml`, environment `pypi`

2. Update version in `pyproject.toml` and `src/libkagome/__init__.py`
    to match `_src_c/go/go.mod` (kagome version)

**Release**:

```bash
git tag v2.10.3  # Match version in pyproject.toml
git push origin v2.10.3

# Alternative: Create release via GitHub UI
# go to https://github.com/KEINOS/kagome-py/releases/new
# The publish.yml workflow will automatically:
#   1. Build wheels for all 5 platforms
#   2. Build source distribution
#   3. Publish to PyPI
```

**Verify** (after ~5 min):

```bash
pip install kagome-py==2.10.3
python -c "from libkagome import Kagome; print(Kagome().wakati('すもも...'))"
```

### Project Structure

```sh
kagome-py/
├── _src_c/                   # Native source (Go + C, NOT installed)
│   ├── go/                   # Go FFI bridge (main.go, tests, go.mod)
│   ├── c/                    # C wrapper (kagome_wrapper.c/h)
│   └── build/                # Build artifacts (.a, .h, .dylib/.so/.dll)
├── src/libkagome/            # Python package (installed)
│   ├── __init__.py           # Public API
│   ├── _wrapper.py           # ctypes FFI bindings
│   └── lib/                  # Shared libraries at install time
├── tests/                    # Test suite
├── .github/workflows/        # CI/CD
│   ├── build-and-test.yml    # Test on 5 platforms
│   └── publish.yml           # Build wheels + publish
├── Makefile                  # Build automation
├── pyproject.toml            # Package metadata (PEP 621)
├── setup.py                  # Build customizations
├── MANIFEST.in               # Sdist contents
└── README.md                 # User documentation
```

### Supported Platforms

| OS | Arch | Wheel | Sdist | Status |
| :-- | :-- | :--: | :--: | :-- |
| macOS | arm64 | ✅ | ✅ | Ready |
| macOS | x86_64 | ✅ | ✅ | Ready |
| Linux (glibc) | x86_64 | ✅ | ✅ | Ready |
| Linux (glibc) | arm64 | ✅ | ✅ | Ready |
| Windows | x86_64 | ✅ | ✅ | Ready |
| Windows | arm64 | ❌ | ❌ | Deferred |
| Alpine (musl) | any | ❌ | ❌ | Not supported |

### Key Files

| File | Purpose |
| :-- | :-- |
| `pyproject.toml` | Package metadata (name, version, dependencies) |
| `setup.py` | Wheel platform tagging + source build hook |
| `MANIFEST.in` | Control what goes into sdist (Go/C sources) |
| `src/libkagome/__init__.py` | Public API entry point |
| `src/libkagome/_wrapper.py` | ctypes FFI to C functions |
| `.github/workflows/build-and-test.yml` | CI on push/PR |
| `.github/workflows/publish.yml` | Build + publish on release |

### Testing Wheels vs Sdist

**Test wheel install**:

```bash
python3 -m venv /tmp/test-whl
source /tmp/test-whl/bin/activate
pip install dist/kagome_py-*.whl
python3 tests/libkagome_test.py
```

**Test sdist install** (builds from source):

```bash
rm src/libkagome/lib/libkagome.*  # Clear pre-built lib
python3 -m venv /tmp/test-src
source /tmp/test-src/bin/activate
pip install dist/kagome_py-*.tar.gz
python3 tests/libkagome_test.py
```

### Troubleshooting

**"Shared library not found"** → Wheel missing from wheel or install failed

- Verify: `unzip -l dist/*.whl | grep libkagome.dylib`
- Reinstall: `pip install --force-reinstall --no-cache-dir dist/*.whl`

**"Go not found"** → Source install (sdist) but Go not installed

- Install Go: <https://golang.org/doc/install>
- Retry: `pip install --no-binary :all: kagome-py`

**Tests fail with strange errors** → Using wrong import path

- ✅ Correct: `from libkagome import Kagome`
- ❌ Wrong: `from _src_c.go import ...` (Go sources, not installed)

---

## Implementation Details

See `IMPLEMENTATION_SUMMARY.md` for:

- Full directory restructure rationale
- Detailed build process
- Architecture decisions
- All changes made
- Verification results

---

## Next Steps

1. **Configure PyPI trusted publisher** (one-time setup)
2. **Tag a release** in GitHub
3. **Verify** the package installs and works
4. **Announce** on relevant channels
