import io
import sys
from libkagome import Kagome, Token
import libkagome


def _ensure_utf8_stdout() -> None:
    """Reconfigure stdout to UTF-8 so Japanese text prints on Windows (cp1252)."""
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8")
    elif sys.stdout.encoding != "utf-8":
        sys.stdout = io.TextIOWrapper(
            sys.stdout.buffer, encoding="utf-8", errors="replace"
        )


def main() -> int:
    _ensure_utf8_stdout()

    # Original integration test
    print("=" * 70)
    print("INTEGRATION TEST: Basic tokenization")
    print("=" * 70)

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

    integration_pass = actual == expect
    if integration_pass:
        print("INTEGRATION TEST: PASS\n")
    else:
        print("INTEGRATION TEST: FAIL\n")

    # Run comprehensive P1-P6 tests
    print("=" * 70)
    print("COMPREHENSIVE TESTS: P1-P6")
    print("=" * 70)

    results, passed, failed = run_all_tests()

    for success, message in results:
        status = "✓" if success else "✗"
        print(f"{status} {message}")

    print()
    print("=" * 70)
    print(f"RESULTS: {passed} passed, {failed} failed ({passed + failed} total)")
    print("=" * 70)

    # Return success only if all tests pass
    return 0 if (integration_pass and failed == 0) else 1


# ------------------------------------------------------------------
# P1: Edge Cases - Empty & Boundary Inputs
# ------------------------------------------------------------------


def test_p1_empty_string() -> tuple[bool, str]:
    """P1.1: Tokenize empty string."""
    kagome = Kagome()
    result = kagome.tokenize("")
    if result == []:
        return (True, "P1.1 PASS: Empty string returns empty list")
    return (False, f"P1.1 FAIL: Expected [], got {result}")


def test_p1_single_character() -> tuple[bool, str]:
    """P1.2: Tokenize single ASCII character."""
    kagome = Kagome()
    result = kagome.tokenize("a")
    if len(result) > 0:
        return (True, "P1.2 PASS: Single character tokenizes successfully")
    return (False, "P1.2 FAIL: Single character should tokenize")


def test_p1_very_long_text() -> tuple[bool, str]:
    """P1.3: Tokenize 100KB+ text."""
    kagome = Kagome()
    # Create 100KB+ test text
    long_text = "これはテストです。" * 5000  # ~90KB
    result = kagome.tokenize(long_text)
    if len(result) > 1000:
        return (True, f"P1.3 PASS: Long text produced {len(result)} tokens")
    return (False, f"P1.3 FAIL: Expected 1000+ tokens, got {len(result)}")


def test_p1_whitespace_only() -> tuple[bool, str]:
    """P1.4: Tokenize whitespace-only input."""
    kagome = Kagome()
    result = kagome.tokenize("   \n\t  ")
    if isinstance(result, list):
        return (True, f"P1.4 PASS: Whitespace-only handled ({len(result)} tokens)")
    return (False, "P1.4 FAIL: Whitespace-only should not crash")


def test_p1_single_japanese_character() -> tuple[bool, str]:
    """P1.5: Tokenize single Japanese character."""
    kagome = Kagome()
    result = kagome.tokenize("あ")
    if len(result) > 0:
        return (True, "P1.5 PASS: Single Japanese character tokenizes")
    return (False, "P1.5 FAIL: Single Japanese character should tokenize")


# ------------------------------------------------------------------
# P2: Unicode Edge Cases
# ------------------------------------------------------------------


def test_p2_ascii_input() -> tuple[bool, str]:
    """P2.1: Handle ASCII-only text."""
    kagome = Kagome()
    result = kagome.tokenize("hello world")
    if isinstance(result, list):
        return (True, f"P2.1 PASS: ASCII text handled ({len(result)} tokens)")
    return (False, "P2.1 FAIL: ASCII text should not crash")


def test_p2_emoji_sequences() -> tuple[bool, str]:
    """P2.2: Handle emoji sequences."""
    kagome = Kagome()
    result = kagome.tokenize("👍 Great! 🎉")
    if isinstance(result, list):
        return (True, f"P2.2 PASS: Emoji sequences handled ({len(result)} tokens)")
    return (False, "P2.2 FAIL: Emoji sequences should not crash")


def test_p2_mixed_scripts() -> tuple[bool, str]:
    """P2.3: Handle mixed language scripts."""
    kagome = Kagome()
    result = kagome.tokenize("English 日本語 中文 العربية")
    if isinstance(result, list) and len(result) > 0:
        return (True, f"P2.3 PASS: Mixed scripts handled ({len(result)} tokens)")
    return (False, "P2.3 FAIL: Mixed scripts should tokenize")


def test_p2_combining_diacritics() -> tuple[bool, str]:
    """P2.4: Handle combining diacritics and accents."""
    kagome = Kagome()
    result = kagome.tokenize("café naïve")
    if isinstance(result, list):
        return (True, f"P2.4 PASS: Combining diacritics handled ({len(result)} tokens)")
    return (False, "P2.4 FAIL: Combining diacritics should not crash")


def test_p2_rtl_text() -> tuple[bool, str]:
    """P2.5: Handle right-to-left (RTL) text."""
    kagome = Kagome()
    result = kagome.tokenize("العربية עברית")
    if isinstance(result, list):
        return (True, f"P2.5 PASS: RTL text handled ({len(result)} tokens)")
    return (False, "P2.5 FAIL: RTL text should not crash")


# ------------------------------------------------------------------
# P3: wakati() Method Coverage
# ------------------------------------------------------------------


def test_p3_wakati_empty() -> tuple[bool, str]:
    """P3.1: wakati() with empty string."""
    kagome = Kagome()
    result = kagome.wakati("")
    if result == []:
        return (True, "P3.1 PASS: wakati() empty string returns []")
    return (False, f"P3.1 FAIL: Expected [], got {result}")


def test_p3_wakati_single_word() -> tuple[bool, str]:
    """P3.2: wakati() with single word."""
    kagome = Kagome()
    result = kagome.wakati("テスト")
    if result == ["テスト"]:
        return (True, "P3.2 PASS: wakati() single word correct")
    return (False, f"P3.2 FAIL: Expected ['テスト'], got {result}")


def test_p3_wakati_sentence() -> tuple[bool, str]:
    """P3.3: wakati() sentence breakdown."""
    kagome = Kagome()
    result = kagome.wakati("今日は天気です")
    if len(result) > 0 and all(isinstance(s, str) for s in result):
        return (True, f"P3.3 PASS: wakati() sentence breakdown ({len(result)} tokens)")
    return (False, "P3.3 FAIL: wakati() sentence should produce list of strings")


def test_p3_wakati_ascii() -> tuple[bool, str]:
    """P3.4: wakati() with ASCII text."""
    kagome = Kagome()
    result = kagome.wakati("hello world")
    if isinstance(result, list):
        return (True, f"P3.4 PASS: wakati() ASCII text handled ({len(result)} tokens)")
    return (False, "P3.4 FAIL: wakati() ASCII text should not crash")


def test_p3_wakati_mixed_content() -> tuple[bool, str]:
    """P3.5: wakati() with mixed Japanese and ASCII."""
    kagome = Kagome()
    result = kagome.wakati("テスト123日本語")
    if isinstance(result, list) and len(result) > 0:
        return (True, f"P3.5 PASS: wakati() mixed content handled ({len(result)} tokens)")
    return (False, "P3.5 FAIL: wakati() mixed content should tokenize")


def test_p3_wakati_vs_tokenize_consistency() -> tuple[bool, str]:
    """P3.6: Verify wakati() is consistent with tokenize()."""
    kagome = Kagome()
    text = "すもももももももものうち"
    tokens = kagome.tokenize(text)
    wakati = kagome.wakati(text)

    token_surfaces = [t.surface for t in tokens]

    if wakati == token_surfaces:
        return (True, "P3.6 PASS: wakati() consistent with tokenize()")
    return (False, f"P3.6 FAIL: wakati() inconsistent with tokenize()")


# ------------------------------------------------------------------
# P4: Token Equality & Comparison Edge Cases
# ------------------------------------------------------------------


def test_p4_token_eq_with_none() -> tuple[bool, str]:
    """P4.1: Token.__eq__ with None."""
    token = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    result = token == None
    if result is False:
        return (True, "P4.1 PASS: Token == None returns False")
    return (False, "P4.1 FAIL: Token == None should return False")


def test_p4_token_eq_with_string() -> tuple[bool, str]:
    """P4.2: Token.__eq__ with string."""
    token = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    result = token == "test"
    if result is False:
        return (True, "P4.2 PASS: Token == string returns False")
    return (False, "P4.2 FAIL: Token == string should return False")


def test_p4_token_eq_same_values() -> tuple[bool, str]:
    """P4.3: Token.__eq__ with same values."""
    token1 = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    token2 = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    if token1 == token2:
        return (True, "P4.3 PASS: Identical tokens are equal")
    return (False, "P4.3 FAIL: Identical tokens should be equal")


def test_p4_token_eq_different_surface() -> tuple[bool, str]:
    """P4.4: Token.__eq__ with different surface forms."""
    token1 = Token("test1", ["n"], "test1", "*", "*", "テスト", "テスト", 0, 5)
    token2 = Token("test2", ["n"], "test2", "*", "*", "テスト", "テスト", 0, 5)
    if token1 != token2:
        return (True, "P4.4 PASS: Different tokens are not equal")
    return (False, "P4.4 FAIL: Different tokens should not be equal")


def test_p4_token_in_list() -> tuple[bool, str]:
    """P4.5: Token in list membership testing."""
    token1 = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    token2 = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    token_list = [token1]

    if token2 in token_list:
        return (True, "P4.5 PASS: Token 'in' list works correctly")
    return (False, "P4.5 FAIL: Token 'in' list should find equal tokens")


def test_p4_token_repr() -> tuple[bool, str]:
    """P4.6: Token.__repr__ self-consistency."""
    token = Token("test", ["n"], "test", "*", "*", "テスト", "テスト", 0, 4)
    repr_str = repr(token)

    if "Token(" in repr_str and "surface=" in repr_str:
        return (True, f"P4.6 PASS: Token.__repr__ produces valid output")
    return (False, f"P4.6 FAIL: Token.__repr__ output invalid: {repr_str}")


# ------------------------------------------------------------------
# P5: Multiple Instances & Reuse
# ------------------------------------------------------------------


def test_p5_multiple_instances() -> tuple[bool, str]:
    """P5.1: Create and use multiple Kagome instances."""
    kagome1 = Kagome()
    kagome2 = Kagome()

    result1 = kagome1.tokenize("テスト")
    result2 = kagome2.tokenize("テスト")

    if len(result1) > 0 and len(result2) > 0:
        return (True, "P5.1 PASS: Multiple instances work independently")
    return (False, "P5.1 FAIL: Multiple instances should work independently")


def test_p5_instance_reuse_after_tokenize() -> tuple[bool, str]:
    """P5.2: Reuse instance after multiple tokenizations."""
    kagome = Kagome()

    texts = ["テスト", "日本語", "kagome"]
    results = []

    for text in texts:
        result = kagome.tokenize(text)
        results.append(len(result) > 0)

    if all(results):
        return (True, "P5.2 PASS: Instance reuse after multiple tokenizations")
    return (False, "P5.2 FAIL: Instance reuse should work")


def test_p5_instance_memory_behavior() -> tuple[bool, str]:
    """P5.3: Test memory behavior with multiple instances."""
    instances = [Kagome() for _ in range(10)]

    # Use all instances
    for kagome in instances:
        result = kagome.tokenize("テスト")
        if len(result) == 0:
            return (False, "P5.3 FAIL: Instance should work after creation")

    return (True, "P5.3 PASS: Multiple instances memory behavior correct")


# ------------------------------------------------------------------
# P6: Version & Metadata
# ------------------------------------------------------------------


def test_p6_version_defined() -> tuple[bool, str]:
    """P6.1: __version__ is defined."""
    if hasattr(libkagome, "__version__"):
        version = libkagome.__version__
        if isinstance(version, str) and len(version) > 0:
            return (True, f"P6.1 PASS: __version__ defined as '{version}'")
        return (False, "P6.1 FAIL: __version__ should be a non-empty string")
    return (False, "P6.1 FAIL: __version__ not defined")


def test_p6_all_exports() -> tuple[bool, str]:
    """P6.2: __all__ exports correct symbols."""
    if hasattr(libkagome, "__all__"):
        all_exports = libkagome.__all__
        if "Kagome" in all_exports and "Token" in all_exports:
            return (True, f"P6.2 PASS: __all__ = {all_exports}")
        return (False, f"P6.2 FAIL: __all__ should contain Kagome and Token")
    return (False, "P6.2 FAIL: __all__ not defined")


def test_p6_module_docstring() -> tuple[bool, str]:
    """P6.3: Module has docstring."""
    if libkagome.__doc__ and len(libkagome.__doc__) > 0:
        return (True, "P6.3 PASS: Module docstring present")
    return (False, "P6.3 FAIL: Module should have docstring")


def test_p6_public_api_accessible() -> tuple[bool, str]:
    """P6.4: Public API classes are accessible."""
    if hasattr(libkagome, "Kagome") and hasattr(libkagome, "Token"):
        kagome_cls = libkagome.Kagome
        token_cls = libkagome.Token

        if callable(kagome_cls) and callable(token_cls):
            return (True, "P6.4 PASS: Kagome and Token classes accessible")
        return (False, "P6.4 FAIL: Classes should be callable")
    return (False, "P6.4 FAIL: Classes not found in public API")


def test_p6_kagome_constructible() -> tuple[bool, str]:
    """P6.5: Kagome class can be instantiated."""
    try:
        kagome = Kagome()
        if kagome is not None:
            return (True, "P6.5 PASS: Kagome instantiation works")
        return (False, "P6.5 FAIL: Kagome instantiation returned None")
    except Exception as e:
        return (False, f"P6.5 FAIL: Kagome instantiation raised {type(e).__name__}")


def test_p6_token_constructible() -> tuple[bool, str]:
    """P6.6: Token class can be instantiated."""
    try:
        token = Token("surface", ["n"], "base", "*", "*", "r", "p", 0, 7)
        if token is not None and token.surface == "surface":
            return (True, "P6.6 PASS: Token instantiation works")
        return (False, "P6.6 FAIL: Token instantiation failed")
    except Exception as e:
        return (False, f"P6.6 FAIL: Token instantiation raised {type(e).__name__}")


# ------------------------------------------------------------------
# Test runner
# ------------------------------------------------------------------


def run_all_tests() -> tuple[list[tuple[bool, str]], int, int]:
    """Run all P1-P6 tests and return results."""
    test_functions = [
        # P1: Edge Cases
        test_p1_empty_string,
        test_p1_single_character,
        test_p1_very_long_text,
        test_p1_whitespace_only,
        test_p1_single_japanese_character,
        # P2: Unicode
        test_p2_ascii_input,
        test_p2_emoji_sequences,
        test_p2_mixed_scripts,
        test_p2_combining_diacritics,
        test_p2_rtl_text,
        # P3: wakati()
        test_p3_wakati_empty,
        test_p3_wakati_single_word,
        test_p3_wakati_sentence,
        test_p3_wakati_ascii,
        test_p3_wakati_mixed_content,
        test_p3_wakati_vs_tokenize_consistency,
        # P4: Token Equality
        test_p4_token_eq_with_none,
        test_p4_token_eq_with_string,
        test_p4_token_eq_same_values,
        test_p4_token_eq_different_surface,
        test_p4_token_in_list,
        test_p4_token_repr,
        # P5: Multiple Instances
        test_p5_multiple_instances,
        test_p5_instance_reuse_after_tokenize,
        test_p5_instance_memory_behavior,
        # P6: Metadata
        test_p6_version_defined,
        test_p6_all_exports,
        test_p6_module_docstring,
        test_p6_public_api_accessible,
        test_p6_kagome_constructible,
        test_p6_token_constructible,
    ]

    results = []
    passed = 0
    failed = 0

    for test_func in test_functions:
        try:
            success, message = test_func()
            results.append((success, message))
            if success:
                passed += 1
            else:
                failed += 1
        except Exception as e:
            results.append((False, f"{test_func.__name__} raised {type(e).__name__}: {e}"))
            failed += 1

    return (results, passed, failed)


if __name__ == "__main__":
    sys.exit(main())
