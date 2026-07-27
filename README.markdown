# Go bindings for GoALLC LLVM

This repository contains the LLVM Go bindings used by GoALLC. It intentionally
targets the customized LLVM build maintained by the GoALLC project; a
system-installed LLVM is not supported.

## LLVM payload

The binding always reads headers and libraries below `${SRCDIR}/llvm`:

    llvm/include/llvm-c
    llvm/lib

`llvm` is not committed. The caller must create it as a symlink to an LLVM
payload with this layout. GoALLC's `cmd/dist` manages that symlink from its
`-llvm-dir` option, whose default is `$GOROOT/llvm`.

## Build tags

The LLVM API version and link mode are independent, mandatory build-tag axes:

* `llvm23` selects the LLVM 23 API. A future LLVM 24 port will add `llvm24`
  without changing how the payload path or link mode is selected.
* `dynamicllvm` links `llvm/lib/libLLVM`; it is the GoALLC default.
* `staticllvm` links the customized aggregate archive
  `llvm/lib/libLLVMGoALLC.a`.

For example:

    go test -tags='llvm23 dynamicllvm' ./...
    go test -tags='llvm23 staticllvm' ./...

Do not select multiple version tags or multiple link-mode tags in one build.

## License

These bindings originated in LLVM and remain licensed under the Apache License
2.0 with LLVM Exceptions. See `LICENSE.txt`.
