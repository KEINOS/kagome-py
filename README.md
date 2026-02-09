# kagome-py

[![gopher_and_python_enjoying_kagome_patterned_pie](./.github/images/kagome-py-logo.png)](https://github.com/KEINOS/kagome-py)

Python bindings for [kagome](https://github.com/ikawaha/kagome), a Japanese morphological analyzer.

> [!NOTE]
> This repository is a spin-off project from the [C shared library example of Kagome](https://github.com/ikawaha/kagome/tree/v2/_examples/clib) to **allow easy installation via `pip` and usage of `libkagome`** (the shared library) in Python.

## Usage

### Installation

```sh
pip install kagome-py
```

### Public API

The `libkagome` library provides the following public API:

```python
class Kagome:
    tokenize(self, text: str) -> list[Token]
    wakati(self, text: str) -> list[str]
```

- `tokenize`: Analyze text and return a list of POS-tagged `Token` objects.
- `wakati`: Split text into words and return them as a list of strings.

### Examples

```shellsession
>>> from libkagome import Kagome
>>> kagome = Kagome()
>>> kagome.wakati("すもももももももものうち")
['すもも', 'も', 'もも', 'も', 'もも', 'の', 'うち']
>>>
>>> tokens = kagome.tokenize("すもももももももものうち")
>>> for token in tokens:
...     print(token)
surface=すもも, pos=['名詞', '一般', '*', '*'], base_form=すもも, conj_type=*, conj_form=*, reading=スモモ, pronunciation=スモモ, start=0, end=3
surface=も, pos=['助詞', '係助詞', '*', '*'], base_form=も, conj_type=*, conj_form=*, reading=モ, pronunciation=モ, start=3, end=4
surface=もも, pos=['名詞', '一般', '*', '*'], base_form=もも, conj_type=*, conj_form=*, reading=モモ, pronunciation=モモ, start=4, end=6
surface=も, pos=['助詞', '係助詞', '*', '*'], base_form=も, conj_type=*, conj_form=*, reading=モ, pronunciation=モ, start=6, end=7
surface=もも, pos=['名詞', '一般', '*', '*'], base_form=もも, conj_type=*, conj_form=*, reading=モモ, pronunciation=モモ, start=7, end=9
surface=の, pos=['助詞', '連体化', '*', '*'], base_form=の, conj_type=*, conj_form=*, reading=ノ, pronunciation=ノ, start=9, end=10
surface=うち, pos=['名詞', '非自立', '副詞可能', '*'], base_form=うち, conj_type=*, conj_form=*, reading=ウチ, pronunciation=ウチ, start=10, end=12
>>>
```

```python
import sys
from libkagome import Kagome

def main() -> int:
    kagome = Kagome()

    text = "すもももももももものうち"

    expect = [
        "surface=すもも, pos=['名詞', '一般', '*', '*'], base_form=すもも, conj_type=*, conj_form=*, reading=スモモ, pronunciation=スモモ, start=0, end=3",
        "surface=も, pos=['助詞', '係助詞', '*', '*'], base_form=も, conj_type=*, conj_form=*, reading=モ, pronunciation=モ, start=3, end=4",
        "surface=もも, pos=['名詞', '一般', '*', '*'], base_form=もも, conj_type=*, conj_form=*, reading=モモ, pronunciation=モモ, start=4, end=6",
        "surface=も, pos=['助詞', '係助詞', '*', '*'], base_form=も, conj_type=*, conj_form=*, reading=モ, pronunciation=モ, start=6, end=7",
        "surface=もも, pos=['名詞', '一般', '*', '*'], base_form=もも, conj_type=*, conj_form=*, reading=モモ, pronunciation=モモ, start=7, end=9",
        "surface=の, pos=['助詞', '連体化', '*', '*'], base_form=の, conj_type=*, conj_form=*, reading=ノ, pronunciation=ノ, start=9, end=10",
        "surface=うち, pos=['名詞', '非自立', '副詞可能', '*'], base_form=うち, conj_type=*, conj_form=*, reading=ウチ, pronunciation=ウチ, start=10, end=12",
    ]

    actual = [str(t) for t in kagome.tokenize(text)]

    for line in actual:
        print(line)

    if actual == expect:
        print("PASS")
        return 0

    print("FAIL")
    return 1

if __name__ == "__main__":
    sys.exit(main())
```

## Supported OS and Architectures

- Python: 3.10+
- OS:
  - macOS (x86_64, Arm64)
  - Linux (x86_64, Arm64) + `glibc` (`manylinux`. No support for `musl`)
  - Windows (x86_64)

> [!IMPORTANT]
> Alpine Linux is not supported due to `musl` limitations.
> For alternative approaches, see the [Dockerfile](./Dockerfile) and [docker-compose.yml](./docker-compose.yml) for development environments.

## Contributing

Any contributions for the betterment are welcome!

- [CONTRIBUTING.md](./.github/CONTRIBUTING.md)

## License

- [MIT License](./LICENSE)
  - kagome: [MIT License](https://github.com/ikawaha/kagome/blob/v2/LICENSE)
