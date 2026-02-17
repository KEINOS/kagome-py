/*
 * Kagome C Wrapper Implementation
 *
 * Wraps Go-exported functions to provide a stable API for FFI users.
 * Built with -fvisibility=hidden so only kagome_* functions are visible.
 */

#include "../build/libkagome.h"
#include "kagome_wrapper.h"

KAGOME_API
void* kagome_init(void) {
    return KagomeInit();
}

KAGOME_API
void kagome_destroy(void* handle) {
    KagomeDestroy(handle);
}

KAGOME_API
TokenArray* kagome_tokenize(void* handle, const char* input) {
    return KagomeTokenizeStruct(handle, (char*)input);
}

KAGOME_API
void kagome_free_token_array(TokenArray* arr) {
    KagomeFreeTokenArray(arr);
}

KAGOME_API
char* kagome_echo(const char* input) {
    return Echo((char*)input);
}

KAGOME_API
void kagome_echo_free(char* str) {
    EchoFree(str);
}
