//===- ir_test.go - Tests for ir ------------------------------------------===//
//
// Part of the LLVM Project, under the Apache License v2.0 with LLVM Exceptions.
// See https://llvm.org/LICENSE.txt for license information.
// SPDX-License-Identifier: Apache-2.0 WITH LLVM-exception
//
//===----------------------------------------------------------------------===//
//
// This file tests bindings for the ir component.
//
//===----------------------------------------------------------------------===//

package llvm

import (
	"strconv"
	"strings"
	"testing"
)

func TestCreateCallBrIntrinsic(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("callbr-intrinsic")
	defer mod.Dispose()
	b := ctx.NewBuilder()
	defer b.Dispose()

	intrinsicID := LookupIntrinsicID("llvm.go.defer.edge")
	if intrinsicID == 0 {
		t.Fatal("llvm.go.defer.edge intrinsic is unavailable")
	}
	deferEdge := GetIntrinsicDeclaration(mod, intrinsicID, nil)
	fn := AddFunction(mod, "f", FunctionType(ctx.VoidType(), nil, false))
	entry := ctx.AddBasicBlock(fn, "entry")
	normal := ctx.AddBasicBlock(fn, "normal")
	recover := ctx.AddBasicBlock(fn, "recover")

	b.SetInsertPointAtEnd(entry)
	b.CreateCallBr(deferEdge.GlobalValueType(), deferEdge, nil, normal, []BasicBlock{recover}, "")
	b.SetInsertPointAtEnd(normal)
	b.CreateRetVoid()
	b.SetInsertPointAtEnd(recover)
	b.CreateRetVoid()

	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatalf("module verification failed: %v\n%s", err, mod.String())
	}
	if got := mod.String(); !strings.Contains(got, "callbr void @llvm.go.defer.edge()") ||
		!strings.Contains(got, "to label %normal [label %recover]") {
		t.Fatalf("module does not contain defer callbr:\n%s", got)
	}
}

func TestCreateCallWithOperandBundles(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("operand-bundle")
	defer mod.Dispose()
	b := ctx.NewBuilder()
	defer b.Dispose()

	donothingID := LookupIntrinsicID("llvm.donothing")
	if donothingID == 0 {
		t.Fatal("llvm.donothing intrinsic is unavailable")
	}
	donothing := GetIntrinsicDeclaration(mod, donothingID, nil)
	fnType := FunctionType(ctx.VoidType(), []Type{PointerType(ctx.Int8Type(), 0)}, false)
	fn := AddFunction(mod, "f", fnType)
	entry := ctx.AddBasicBlock(fn, "entry")

	b.SetInsertPointAtEnd(entry)
	bundle := NewOperandBundle("go.keepalive", []Value{fn.Param(0)})
	b.CreateCallWithOperandBundles(donothing.GlobalValueType(), donothing, nil,
		[]OperandBundle{bundle}, "")
	bundle.Dispose()
	b.CreateRetVoid()

	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatalf("module verification failed: %v\n%s", err, mod.String())
	}
	if got := mod.String(); !strings.Contains(got,
		`call void @llvm.donothing() [ "go.keepalive"(ptr %0) ]`) {
		t.Fatalf("module does not contain operand bundle:\n%s", got)
	}
}

func TestReplaceIncomingBlock(t *testing.T) {
	ctx := NewContext()
	defer ctx.Dispose()
	mod := ctx.NewModule("phi-incoming-block")
	defer mod.Dispose()
	b := ctx.NewBuilder()
	defer b.Dispose()

	fn := AddFunction(mod, "f", FunctionType(ctx.Int32Type(), nil, false))
	entry := ctx.AddBasicBlock(fn, "entry")
	old := ctx.AddBasicBlock(fn, "old")
	replacement := ctx.AddBasicBlock(fn, "replacement")
	merge := ctx.AddBasicBlock(fn, "merge")

	b.SetInsertPointAtEnd(entry)
	b.CreateBr(replacement)
	b.SetInsertPointAtEnd(old)
	b.CreateRet(ConstInt(ctx.Int32Type(), 0, false))
	b.SetInsertPointAtEnd(replacement)
	b.CreateBr(merge)
	b.SetInsertPointAtEnd(merge)
	phi := b.CreatePHI(ctx.Int32Type(), "value")
	phi.AddIncoming([]Value{ConstInt(ctx.Int32Type(), 7, false)}, []BasicBlock{old})
	b.CreateRet(phi)

	phi.ReplaceIncomingBlock(old, replacement)
	if got := phi.IncomingBlock(0); got != replacement {
		t.Fatalf("incoming block = %v, want replacement", got)
	}
	if err := VerifyModule(mod, ReturnStatusAction); err != nil {
		t.Fatalf("module verification failed: %v\n%s", err, mod.String())
	}
}

func testAttribute(t *testing.T, name string) {
	ctx := NewContext()
	mod := ctx.NewModule("")
	defer mod.Dispose()

	ftyp := FunctionType(ctx.VoidType(), nil, false)
	fn := AddFunction(mod, "foo", ftyp)

	kind := AttributeKindID(name)
	attr := mod.Context().CreateEnumAttribute(kind, 0)

	fn.AddFunctionAttr(attr)
	newattr := fn.GetEnumFunctionAttribute(kind)
	if attr != newattr {
		t.Errorf("got attribute %p, want %p", newattr.C, attr.C)
	}

	text := mod.String()
	if !strings.Contains(text, " "+name+" ") {
		t.Errorf("expected attribute '%s', got:\n%s", name, text)
	}

	fn.RemoveEnumFunctionAttribute(kind)
	newattr = fn.GetEnumFunctionAttribute(kind)
	if !newattr.IsNil() {
		t.Errorf("got attribute %p, want 0", newattr.C)
	}
}

func TestAttributes(t *testing.T) {
	// Tests that our attribute constants haven't drifted from LLVM's.
	attrTests := []string{
		"sanitize_address",
		"alwaysinline",
		"builtin",
		"convergent",
		"inlinehint",
		"inreg",
		"jumptable",
		"minsize",
		"naked",
		"nest",
		"noalias",
		"nobuiltin",
		"noduplicate",
		"noimplicitfloat",
		"noinline",
		"nonlazybind",
		"nonnull",
		"noredzone",
		"noreturn",
		"nounwind",
		"optnone",
		"optsize",
		"readnone",
		"readonly",
		"returned",
		"returns_twice",
		"signext",
		"safestack",
		"ssp",
		"sspreq",
		"sspstrong",
		"sanitize_thread",
		"sanitize_memory",
		"uwtable",
		"zeroext",
		"cold",
		"nocf_check",
	}

	for _, name := range attrTests {
		majorVersion, err := strconv.Atoi(strings.SplitN(Version, ".", 2)[0])
		if err != nil {
			// sanity check, should be unreachable
			t.Errorf("could not parse LLVM version: %v", err)
		}
		if majorVersion >= 15 && name == "uwtable" {
			// This changed from an EnumAttr to an IntAttr in LLVM 15, and testAttribute doesn't work on such attributes.
			continue
		}
		testAttribute(t, name)
	}
}

func TestDebugLoc(t *testing.T) {
	ctx := NewContext()
	mod := ctx.NewModule("")
	defer mod.Dispose()

	b := ctx.NewBuilder()
	defer b.Dispose()

	d := NewDIBuilder(mod)
	defer func() {
		d.Destroy()
	}()
	file := d.CreateFile("dummy_file", "dummy_dir")
	voidInfo := d.CreateBasicType(DIBasicType{Name: "void"})
	typeInfo := d.CreateSubroutineType(DISubroutineType{
		File:       file,
		Parameters: []Metadata{voidInfo},
		Flags:      0,
	})
	scope := d.CreateFunction(file, DIFunction{
		Name:         "foo",
		LinkageName:  "foo",
		Line:         10,
		ScopeLine:    10,
		Type:         typeInfo,
		File:         file,
		IsDefinition: true,
	})

	b.SetCurrentDebugLocation(10, 20, scope, Metadata{})
	loc := b.GetCurrentDebugLocation()
	if loc.Line != 10 {
		t.Errorf("Got line %d, though wanted 10", loc.Line)
	}
	if loc.Col != 20 {
		t.Errorf("Got column %d, though wanted 20", loc.Col)
	}
	if loc.Scope.C != scope.C {
		t.Errorf("Got metadata %v as scope, though wanted %v", loc.Scope.C, scope.C)
	}
}

func TestSubtypes(t *testing.T) {
	cont := NewContext()
	defer cont.Dispose()

	st_pointer := cont.StructType([]Type{cont.Int32Type(), cont.Int8Type()}, false)
	st_inner := st_pointer.Subtypes()
	if len(st_inner) != 2 {
		t.Errorf("Got size %d, though wanted 2", len(st_inner))
	}
	if st_inner[0] != cont.Int32Type() {
		t.Errorf("Expected first struct field to be int32")
	}
	if st_inner[1] != cont.Int8Type() {
		t.Errorf("Expected second struct field to be int8")
	}
}
