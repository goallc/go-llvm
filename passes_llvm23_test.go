//go:build llvm23 || (!byollvm && !llvm14 && !llvm15 && !llvm16 && !llvm17 && !llvm18 && !llvm19 && !llvm20 && !llvm21 && !llvm22)

package llvm

// LLVM 23 removed default<Os>. Keep this test focused on the LLVMRunPasses API:
// the current arm64 target-machine pipeline asserts on this legacy recursive
// fixture at O2, while O0 exercises the same C API without target transforms.
const defaultTestPipeline = "default<O0>"

func prepareDefaultTestPipeline(Context, Value) {}
