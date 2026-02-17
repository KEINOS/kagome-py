// Package cgotest provides test helpers for cgo-based Kagome tokenizer testing.
// It is an internal package and should not be imported outside of tests.
package cgotest

/*
#include <stdint.h>
#include <stdlib.h>

// Declare C functions from main.go
void* KagomeInit(void);
void KagomeDestroy(void* handle);

typedef struct {
	char* surface;
	char* pos1;
	char* pos2;
	char* pos3;
	char* pos4;
	char* base_form;
	char* conj_type;
	char* conj_form;
	char* reading;
	char* pronunciation;
	int   start;
	int   end;
} Token;

typedef struct {
	Token* tokens;
	int    length;
} TokenArray;

TokenArray* KagomeTokenizeStruct(void* handle, const char* input);
void KagomeFreeTokenArray(TokenArray* arr);
*/
import "C"

import "unsafe"

// TokenizeString tokenizes a Go string using the Kagome tokenizer.
//
//nolint:nlreturn // Guarded flow keeps CGO cleanup simple.
func TokenizeString(handle unsafe.Pointer, text string) int {
	count := 0
	if handle != nil && text != "" {
		cStr := C.CString(text)
		if cStr != nil {
			defer C.free(unsafe.Pointer(cStr))

			arr := C.KagomeTokenizeStruct(handle, cStr)
			if arr != nil {
				count = int(arr.length)
				C.KagomeFreeTokenArray(arr)
			}
		}
	}

	return count
}

// HandleExists checks if a handle is still valid.
//
//nolint:nlreturn // Guarded flow keeps CGO cleanup simple.
func HandleExists(handle unsafe.Pointer) bool {
	exists := false
	if handle != nil {
		cStr := C.CString("a")
		if cStr != nil {
			defer C.free(unsafe.Pointer(cStr))

			arr := C.KagomeTokenizeStruct(handle, cStr)
			if arr != nil {
				C.KagomeFreeTokenArray(arr)
				exists = true
			}
		}
	}

	return exists
}
