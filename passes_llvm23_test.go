//go:build llvm23

package llvm

// LLVM 23 removed default<Os>. Keep this test focused on the LLVMRunPasses API:
// the current arm64 target-machine pipeline asserts on this legacy recursive
// fixture at O2, while O0 exercises the same C API without target transforms.
const defaultTestPipeline = "default<O0>"

func prepareDefaultTestPipeline(Context, Value) {}
