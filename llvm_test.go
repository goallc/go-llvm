package llvm_test

import "github.com/goallc/go-llvm"

import "testing"

// Dummy test function.
// All it does is test whether we can use LLVM at all.
func TestLLVM(t *testing.T) {
	t.Log("LLVM version:", llvm.Version)
}
