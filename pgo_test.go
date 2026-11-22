package llvm

import (
	"strings"
	"testing"
)

func TestRunPassesWithPGO(t *testing.T) {
	InitializeNativeTarget()
	triple := DefaultTargetTriple()
	target, err := GetTargetFromTriple(triple)
	if err != nil {
		t.Fatal(err)
	}
	tm := target.CreateTargetMachine(triple, "", "", CodeGenLevelDefault, RelocDefault, CodeModelDefault)
	defer tm.Dispose()
	for _, tc := range []struct {
		passes  string
		wantErr string
	}{
		{"default<O2>", ""},
		{"function(simplifycfg)", ""},
		{"module(pgo-instr-gen,instrprof)", ""},
		{"default<O2>,pgo-instr-gen,instrprof", "conflicting LLVM PGO passes"},
		{"instrprof,pgo-instr-gen", "requires preceding"},
		{"coro-cond(pgo-instr-gen,instrprof)", "nested LLVM PGO passes"},
	} {
		t.Run(tc.passes, func(t *testing.T) {
			ctx := NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("profile")
			defer mod.Dispose()
			mod.SetTarget(triple)
			fn := AddFunction(mod, "entry", FunctionType(ctx.VoidType(), nil, false))
			b := ctx.NewBuilder()
			defer b.Dispose()
			b.SetInsertPointAtEnd(ctx.AddBasicBlock(fn, "entry"))
			b.CreateRetVoid()
			err := mod.RunPassesWithPGO(tc.passes, tm, "")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("got %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(mod.String(), "atomicrmw add") {
				t.Error("missing compiler-rt counters")
			}
		})
	}
}
