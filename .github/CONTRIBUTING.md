# Information for Contributors

## Branch to Pull Request

- `main` branch

## Versioning

Versions of `kagome-py` are aligned with the versions of `kagome`. For example, `kagome-py 2.0.0` corresponds to upstream `kagome v2.0.0`.

However, if there are any fixes in `kagome-py` without changes in `kagome`, we append a suffix to the version.

For example, `kagome-py 2.0.0.20260210` corresponds to:

- `kagome` v2.0.0
- `kagome-py` fix released on 2026-02-10

In each release note, we specify the Python versions that are tested as compatible.

## Fail-Fast Policy/Backward Compatibility

We follow a [fail-fast](https://en.wikipedia.org/wiki/Fail-fast_system) policy. **We support only the latest released version of `kagome`**. For Python, we guarantee support only for the version(s) that successfully pass our CI tests.

We do not actively maintain backward compatibility with older Python versions or past releases of `kagome-py`. In those cases, please consider one of the following options:

- Refer to the corresponding older released versions of `kagome-py`
- Use Docker images provided for evaluation
- Build from source

## Dockerfiles

This repository aims to provide Python bindings for `kagome` that can be installed via `pip`.

However, for evaluation and testing purposes, we also provide Dockerfiles.

The Dockerfile in the root directory is the one we use in our CI tests. Which means it always uses the latest released version of `kagome-py` and the latest released version of Python.

The Dockerfiles in the `docker/` directory are for evaluation purposes. Each Dockerfile specifies a particular version of Python and `kagome-py`. You can use them to test compatibility with specific versions.
