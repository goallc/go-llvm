// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception

package llvm

import (
	"strings"
	"testing"
)

func TestGlobalMetadataOperand(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("metadata")
	defer mod.Dispose()
	fn := AddFunction(mod, "f", FunctionType(ctx.VoidType(), nil, false))
	kind := ctx.MDKindID("test.identity")
	if !fn.GlobalMetadata(kind).IsNil() {
		t.Fatal("missing attachment is not nil")
	}
	name := ctx.MDString("storage")
	index := ConstInt(ctx.Int32Type(), 7, false).ConstantAsMetadata()
	fn.SetGlobalMetadata(kind, ctx.MDNode([]Metadata{name, index}))
	got := fn.GlobalMetadata(kind)
	if got.Operand(0) != name || got.Operand(1) != index {
		t.Fatal("attachment operands lost their metadata identity")
	}
	fn.EraseGlobalMetadata(kind)
	if !fn.GlobalMetadata(kind).IsNil() {
		t.Fatal("erased attachment is not nil")
	}
}

func TestFunctionAnnotations(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("annotations")
	defer mod.Dispose()
	d := NewDIBuilder(mod)
	defer d.Destroy()
	d.CreateCompileUnit(DICompileUnit{Language: DW_LANG_Go, File: "test.go", Producer: "test"})
	file := d.CreateFile("test.go", "/tmp")
	typ := d.CreateSubroutineType(DISubroutineType{File: file})
	annotations := ctx.TemporaryMDNode(nil)
	sp := d.CreateFunction(file, DIFunction{
		Name: "f", LinkageName: "f", File: file, Line: 1,
		Type: typ, IsDefinition: true, Annotations: annotations,
	})
	fn := AddFunction(mod, "f", FunctionType(ctx.VoidType(), nil, false))
	fn.SetSubprogram(sp)
	b := ctx.NewBuilder()
	defer b.Dispose()
	b.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
	b.SetCurrentDebugLocation(1, 1, sp, Metadata{})
	b.CreateRetVoid()
	annotations.ReplaceAllUsesWith(ctx.MDNode([]Metadata{
		ctx.MDNode([]Metadata{ctx.MDString("identity"), ctx.MDString("storage")}),
	}))
	d.Finalize()
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	ir := mod.String()
	if !strings.Contains(ir, "annotations: !") || !strings.Contains(ir, `!{!"identity", !"storage"}`) {
		t.Fatalf("function annotation was not resolved:\n%s", ir)
	}
}
