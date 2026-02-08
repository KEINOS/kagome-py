# TODO: Add pip distribution packaging for kagome-py

## Directory Restructure
- [ ] Move `libkagome/go_wrapper/` -> `_src_c/go/`
- [ ] Move `libkagome/c_wrapper/` -> `_src_c/c/`
- [ ] Create `_src_c/build/.gitkeep`
- [ ] Remove old `libkagome/` directory

## Python Package
- [ ] Create `src/libkagome/__init__.py`
- [ ] Create `src/libkagome/_wrapper.py` (adapted from python_wrapper/libkagome.py)
- [ ] Create `src/libkagome/lib/.gitkeep`

## Packaging Files
- [ ] Create `pyproject.toml`
- [ ] Create `setup.py` (bdist_wheel tag override + sdist build hook)
- [ ] Create `MANIFEST.in`

## Update Build Files
- [ ] Update `Makefile` (new paths, stage-lib/wheel/sdist targets)
- [ ] Update `.gitignore` (packaging artifacts)
- [ ] Update `_src_c/c/kagome_wrapper.c` (#include path)

## Update Docker Files
- [ ] Update `Dockerfile` (paths)
- [ ] Update `docker-compose.yml` (volume mounts)

## Documentation
- [ ] Update `README.md` (min Python 3.10+)

## Verify
- [ ] `make clean && make build && make test-python`
- [ ] `make wheel` and inspect
- [ ] `make sdist` and inspect

## CI/CD (after local verification)
- [ ] Create `.github/workflows/build-and-test.yml`
- [ ] Create `.github/workflows/publish.yml`
