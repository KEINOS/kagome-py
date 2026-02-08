"""
Python bindings for kagome, a Japanese morphological analyzer.

Usage:
    from libkagome import Kagome

    kagome = Kagome()
    tokens = kagome.tokenize("すもももももももものうち")
    words = kagome.wakati("すもももももももものうち")
"""

from libkagome._wrapper import Kagome, Token

__all__ = ["Kagome", "Token"]
__version__ = "2.10.3"
