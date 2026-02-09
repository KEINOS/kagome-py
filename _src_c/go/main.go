// Package main provides C FFI bindings for Kagome, a Japanese morphological analyzer.
// It exports a C API for initializing, tokenizing, and managing Kagome instances.
package main

/*
#include <stdint.h>
#include <stdlib.h>

// ------------------------------------------------------------------
// C ABI structures
// ------------------------------------------------------------------

// Token represents one morphological token.
// All strings are UTF-8, null-terminated, and allocated with malloc.
typedef struct {
	char* surface;        // Surface form (表層形)
	char* pos1;           // Part-of-speech hierarchy 0 (major class, 大分類)
	char* pos2;           // Part-of-speech hierarchy 1 (middle class, 中分類)
	char* pos3;           // Part-of-speech hierarchy 2 (small class, 小分類)
	char* pos4;           // Part-of-speech hierarchy 3 (fine class, 再分類)
	char* base_form;      // Base form / dictionary form (原形・基本形)
	char* conj_type;      // Conjugation type (活用型)
	char* conj_form;      // Conjugation form (活用形)
	char* reading;        // Reading in katakana (読み)
	char* pronunciation;  // Pronunciation (発音)
	int   start;          // Start position (開始位置)
	int   end;            // End position (終了位置)
} Token;

// TokenArray is an owned array returned to foreign languages.
// Both the array itself and all nested strings must be freed
// by calling KagomeFreeTokenArray.
typedef struct {
	Token* tokens;
	int    length;
} TokenArray;
*/
import "C"

import (
	"sync"
	"unsafe"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// ------------------------------------------------------------------
// Constants for token feature indices
// ------------------------------------------------------------------

const (
	posIndex0 = iota // POS hierarchy 0 (major class)
	posIndex1        // POS hierarchy 1 (middle class)
	posIndex2        // POS hierarchy 2 (small class)
	posIndex3        // POS hierarchy 3 (fine class)
)

const (
	conjTypeIndex      = 4 // Conjugation type
	conjFormIndex      = 5 // Conjugation form
	baseFormIndex      = 6 // Base form
	readingIndex       = 7 // Reading in katakana
	pronunciationIndex = 8 // Pronunciation
)

// ------------------------------------------------------------------
// Internal state management
// ------------------------------------------------------------------

// instances maps opaque C handles to Kagome tokenizers.
//
// IMPORTANT (cgo rule):
//
//	Go pointers must never be passed to C.
//	Therefore, we allocate a dummy pointer using C.malloc()
//	and use that pointer as the external handle.
//
//nolint:gochecknoglobals // Required for FFI handle management.
var (
	instanceMutex = sync.Mutex{}
	instances     = make(map[unsafe.Pointer]*tokenizer.Tokenizer)
)

// ------------------------------------------------------------------
// Helper utilities
// ------------------------------------------------------------------

// getOrEmpty returns arr[index] if it exists, otherwise an empty string.
// This avoids bounds checks at every call site.
func getOrEmpty(arr []string, index int) string {
	if index >= 0 && index < len(arr) {
		return arr[index]
	}

	return ""
}

// wouldOverflowTokenAllocation checks if allocating count tokens would cause integer overflow.
// Returns true if the allocation would be unsafe.
func wouldOverflowTokenAllocation(count int) bool {
	if count < 0 {
		return true
	}

	//nolint:exhaustruct // C struct size calculation, fields not needed.
	tokenSize := C.size_t(unsafe.Sizeof(C.Token{}))
	maxSafeTokens := (^C.size_t(0)) / tokenSize

	return C.size_t(count) > maxSafeTokens
}

// freeStrings safely frees multiple C strings.
// Checks for nil before freeing (safe to pass nil pointers).
func freeStrings(strs ...*C.char) {
	for _, s := range strs {
		if s != nil {
			C.free(unsafe.Pointer(s))
		}
	}
}

// ------------------------------------------------------------------
// Exported C API
// ------------------------------------------------------------------

//export KagomeInit
func KagomeInit() unsafe.Pointer {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	tok, err := tokenizer.New(
		ipa.Dict(),
		tokenizer.OmitBosEos(),
	)
	if err != nil {
		return nil
	}

	// Allocate an opaque handle in C memory.
	// This pointer is safe to pass across FFI boundaries.
	handle := C.malloc(1)
	if handle == nil {
		return nil
	}

	instances[handle] = tok

	return handle
}

//export KagomeDestroy
func KagomeDestroy(handle unsafe.Pointer) {
	if handle == nil {
		return
	}

	instanceMutex.Lock()
	delete(instances, handle)
	instanceMutex.Unlock()

	// Free the dummy handle allocated in KagomeInit.
	C.free(handle)
}

// tokenStrings holds all C strings for a single token.
type tokenStrings struct {
	surface, pos1, pos2, pos3, pos4           *C.char
	conjType, conjForm, baseForm, reading     *C.char
	pronunciation                             *C.char
}

// allocateTokenStrings allocates C strings for a single token.
// Returns all allocated strings or nil for any, indicating allocation failure.
func allocateTokenStrings(tok *tokenizer.Token) tokenStrings {
	pos := tok.POS()
	features := tok.Features()

	return tokenStrings{
		surface:       C.CString(tok.Surface),
		pos1:          C.CString(getOrEmpty(pos, posIndex0)),
		pos2:          C.CString(getOrEmpty(pos, posIndex1)),
		pos3:          C.CString(getOrEmpty(pos, posIndex2)),
		pos4:          C.CString(getOrEmpty(pos, posIndex3)),
		conjType:      C.CString(getOrEmpty(features, conjTypeIndex)),
		conjForm:      C.CString(getOrEmpty(features, conjFormIndex)),
		baseForm:      C.CString(getOrEmpty(features, baseFormIndex)),
		reading:       C.CString(getOrEmpty(features, readingIndex)),
		pronunciation: C.CString(getOrEmpty(features, pronunciationIndex)),
	}
}

// allocationsFailed checks if any string allocation in tokenStrings failed.
func (ts tokenStrings) allocationsFailed() bool {
	return ts.surface == nil || ts.pos1 == nil || ts.pos2 == nil ||
		ts.pos3 == nil || ts.pos4 == nil || ts.conjType == nil ||
		ts.conjForm == nil || ts.baseForm == nil || ts.reading == nil ||
		ts.pronunciation == nil
}

// checkAndStoreToken verifies string allocations and stores token in slice.
// Returns false if any string allocation failed, indicating cleanup is needed.
func checkAndStoreToken(index int, slice []C.Token, tok *tokenizer.Token, strings tokenStrings) bool {
	if strings.allocationsFailed() {
		return false
	}

	slice[index] = C.Token{
		surface:       strings.surface,
		pos1:          strings.pos1,
		pos2:          strings.pos2,
		pos3:          strings.pos3,
		pos4:          strings.pos4,
		conj_type:     strings.conjType,
		conj_form:     strings.conjForm,
		base_form:     strings.baseForm,
		reading:       strings.reading,
		pronunciation: strings.pronunciation,
		start:         C.int(tok.Start),
		end:           C.int(tok.End),
	}

	return true
}

// cleanupTokens frees all strings in tokens up to the given count.
func cleanupTokens(slice []C.Token, count int) {
	for index := range count {
		token := slice[index]

		freeStrings(
			token.surface,
			token.pos1,
			token.pos2,
			token.pos3,
			token.pos4,
			token.conj_type,
			token.conj_form,
			token.base_form,
			token.reading,
			token.pronunciation,
		)
	}
}

// free frees all strings in a tokenStrings struct.
func (ts tokenStrings) free() {
	freeStrings(
		ts.surface, ts.pos1, ts.pos2, ts.pos3, ts.pos4,
		ts.conjType, ts.conjForm, ts.baseForm, ts.reading, ts.pronunciation,
	)
}

//export KagomeTokenizeStruct
func KagomeTokenizeStruct(handle unsafe.Pointer, input *C.char) *C.TokenArray {
	if handle == nil || input == nil {
		return nil
	}

	// Lock ONLY for map access.
	instanceMutex.Lock()
	tokenizer := instances[handle]
	instanceMutex.Unlock() // early unlock after getting the instance

	if tokenizer == nil {
		return nil
	}

	text := C.GoString(input)
	tokens := tokenizer.Tokenize(text)
	count := len(tokens)

	// Check for integer overflow in allocation.
	if wouldOverflowTokenAllocation(count) {
		return nil
	}

	// Allocate TokenArray (always owned by caller).
	//nolint:exhaustruct // C struct size calculation, fields not needed.
	arr := (*C.TokenArray)(C.malloc(C.size_t(unsafe.Sizeof(C.TokenArray{}))))
	if arr == nil {
		return nil
	}

	arr.length = C.int(count)

	if count == 0 {
		arr.tokens = nil

		return arr
	}

	// Allocate contiguous Token array.
	//nolint:exhaustruct // C struct size calculation, fields not needed.
	cTokens := (*C.Token)(C.malloc(
		C.size_t(count) * C.size_t(unsafe.Sizeof(C.Token{})),
	))
	if cTokens == nil {
		C.free(unsafe.Pointer(arr))

		return nil
	}

	slice := unsafe.Slice(cTokens, count)

	for index, tok := range tokens {
		strings := allocateTokenStrings(&tok)

		if !checkAndStoreToken(index, slice, &tok, strings) {
			// Free strings we just allocated for current token.
			strings.free()

			// Free all previously completed tokens.
			cleanupTokens(slice, index)

			C.free(unsafe.Pointer(cTokens))
			C.free(unsafe.Pointer(arr))

			return nil
		}
	}

	arr.tokens = cTokens

	return arr
}

//export KagomeFreeTokenArray
func KagomeFreeTokenArray(arr *C.TokenArray) {
	if arr == nil {
		return
	}

	if arr.tokens != nil {
		slice := unsafe.Slice(arr.tokens, int(arr.length))
		for _, token := range slice {
			C.free(unsafe.Pointer(token.surface))
			C.free(unsafe.Pointer(token.pos1))
			C.free(unsafe.Pointer(token.pos2))
			C.free(unsafe.Pointer(token.pos3))
			C.free(unsafe.Pointer(token.pos4))
			C.free(unsafe.Pointer(token.base_form))
			C.free(unsafe.Pointer(token.conj_type))
			C.free(unsafe.Pointer(token.conj_form))
			C.free(unsafe.Pointer(token.reading))
			C.free(unsafe.Pointer(token.pronunciation))
		}
		C.free(unsafe.Pointer(arr.tokens))
	}

	// Always free the container itself.
	C.free(unsafe.Pointer(arr))
}

// ------------------------------------------------------------------
// Test utilities (wrapped as kagome_echo/kagome_echo_free)
// ------------------------------------------------------------------

// Echo copies a string and returns it.
// Used for testing FFI setup (string passing, memory allocation).
//
// FFI users should call kagome_echo() from the C wrapper, not this directly.
//
//export Echo
func Echo(input *C.char) *C.char {
	if input == nil {
		return nil
	}

	return C.CString(C.GoString(input))
}

// EchoFree frees a string returned by Echo.
// FFI users should call kagome_echo_free() from the C wrapper, not this directly.
//
//export EchoFree
func EchoFree(p *C.char) {
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}

// ------------------------------------------------------------------
// Test Helpers (in main.go due to Go cgo limitation)
// ------------------------------------------------------------------
//
// NOTE: These functions are exclusively for testing and are placed in main.go
// because Go does not support cgo imports in *_test.go files. Even though they
// are helper functions, they cannot be moved to a separate test file without
// breaking the cgo preamble requirement.
//
// These functions are only called from main_test.go and have no impact on
// the production FFI API.

// testTokenizeString is a test helper that tokenizes a Go string.
// It handles the C string conversion internally and returns the token count.
// This is only used by tests and is not part of the public API.
func testTokenizeString(handle unsafe.Pointer, text string) int {
	if handle == nil || text == "" {
		return 0
	}

	cStr := C.CString(text)
	defer C.free(unsafe.Pointer(cStr))

	arr := KagomeTokenizeStruct(handle, cStr)
	if arr == nil {
		return 0
	}

	count := int(arr.length)
	KagomeFreeTokenArray(arr)

	return count
}

// testHandleExists checks if a handle is still in the instances map.
// This is only used by tests and is not part of the public API.
func testHandleExists(handle unsafe.Pointer) bool {
	if handle == nil {
		return false
	}

	instanceMutex.Lock()
	_, exists := instances[handle]
	instanceMutex.Unlock()

	return exists
}

func main() {}