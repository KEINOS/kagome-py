"""
Python bindings for kagome, a Japanese morphological analyzer.

Usage:
    from libkagome import Kagome

    kagome = Kagome()
    tokens = kagome.tokenize("すもももももももものうち")
    words = kagome.wakati("すもももももももものうち")
"""

from importlib.metadata import PackageNotFoundError, version

from libkagome._wrapper import Kagome, Token

__all__ = ["Kagome", "Token"]

try:
    __version__ = version("kagome-py")
except PackageNotFoundError:
    __version__ = "0.0.0-dev"
