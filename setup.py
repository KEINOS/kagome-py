"""
Build support for kagome-py.

Responsibilities:
1. Force platform-specific wheel tags (not pure Python) since the package
   includes a native shared library loaded via ctypes.
2. Build the shared library from Go/C source when installing from sdist.

All metadata lives in pyproject.toml.
"""

import os
import platform
import shutil
import subprocess
from pathlib import Path

from setuptools import setup
from setuptools.command.build_py import build_py as _build_py

# ---------------------------------------------------------------------------
# Platform wheel tagging
# ---------------------------------------------------------------------------

try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

    class bdist_wheel(_bdist_wheel):
        def finalize_options(self):
            _bdist_wheel.finalize_options(self)
            self.root_is_pure = False

        def get_tag(self):
            _, _, plat = _bdist_wheel.get_tag(self)
            return "py3", "none", plat

except ImportError:
    bdist_wheel = None


# ---------------------------------------------------------------------------
# Source build hook (for sdist installs)
# ---------------------------------------------------------------------------

# Shared library extension per platform
_LIB_EXT = {
    "Darwin": "dylib",
    "Linux": "so",
    "Windows": "dll",
}


def _lib_ext() -> str:
    """Get the shared library extension for the current platform."""
    ext = _LIB_EXT.get(platform.system())
    if ext is None:
        raise RuntimeError(f"Unsupported platform: {platform.system()}")
    return ext


def _find_tool(name: str) -> str:
    """Find a build tool on PATH, raising a clear error if missing."""
    path = shutil.which(name)
    if path is None:
        raise RuntimeError(
            f"'{name}' not found. Building kagome-py from source requires "
            f"Go (>=1.25) and a C compiler (cc/gcc). "
            f"Install them or use a pre-built wheel: pip install kagome-py"
        )
    return path


def _build_native_library(root: Path) -> None:
    """Build the Go/C shared library if not already present."""
    lib_dir = root / "src" / "libkagome" / "lib"
    ext = _lib_ext()
    lib_file = lib_dir / f"libkagome.{ext}"

    # Skip if shared library already exists (e.g., pre-built wheel)
    if lib_file.exists():
        return

    go_dir = root / "_src_c" / "go"

    # Skip if no Go source (pre-built wheel, not sdist)
    if not go_dir.exists():
        return

    go = _find_tool("go")
    cc = _find_tool("cc") if platform.system() != "Windows" else _find_tool("gcc")

    build_dir = root / "_src_c" / "build"
    c_dir = root / "_src_c" / "c"

    build_dir.mkdir(parents=True, exist_ok=True)
    lib_dir.mkdir(parents=True, exist_ok=True)

    env = os.environ.copy()
    env["CGO_ENABLED"] = "1"

    # Step 1: Build Go c-archive
    archive = build_dir / "libkagome.a"
    subprocess.check_call(
        [
            go, "build",
            "-buildmode=c-archive",
            "-o", str(archive),
            "./main.go",
        ],
        cwd=str(go_dir),
        env=env,
    )

    # Step 2: Build shared library from C wrapper + Go archive
    shared_lib = build_dir / f"libkagome.{ext}"
    cc_cmd = [
        cc,
        "-fPIC",
        "-fvisibility=hidden",
        "-shared",
        str(c_dir / "kagome_wrapper.c"),
        str(archive),
        "-o", str(shared_lib),
    ]
    # Go runtime on Windows needs extra link flags
    if platform.system() == "Windows":
        cc_cmd.extend(["-lwinmm", "-lws2_32", "-lntdll"])

    subprocess.check_call(cc_cmd, cwd=str(root / "_src_c"), env=env)

    # Step 3: Copy to package lib directory
    shutil.copy2(str(shared_lib), str(lib_file))


class build_py(_build_py):
    """Build native library before collecting Python files and package_data."""

    def run(self):
        # Build Go/C shared library BEFORE build_py collects package_data.
        # This ensures the library is present when setuptools scans lib/*.
        root = Path(self.distribution.src_root or ".").resolve()
        _build_native_library(root)
        _build_py.run(self)


# ---------------------------------------------------------------------------
# Setup
# ---------------------------------------------------------------------------

cmdclass = {"build_py": build_py}
if bdist_wheel is not None:
    cmdclass["bdist_wheel"] = bdist_wheel

setup(cmdclass=cmdclass)
