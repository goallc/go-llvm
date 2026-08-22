package llvm

import (
	"os"
	"strings"
	"testing"
)

func TestConfigureGoObjFromModule(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		wantErr  string
	}{
		{name: "absent"},
		{
			name:     "invalid field count",
			metadata: "!goobj.config = !{!0}\n!0 = !{!\"goallc.goobj\"}",
			wantErr:  "must contain twelve fields",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := "define void @f() { ret void }\n" + test.metadata + "\n"
			file, err := os.CreateTemp("", "goobj-config-*.ll")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(file.Name())
			if _, err := file.WriteString(source); err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}

			ctx := NewContext()
			defer ctx.Dispose()
			buf, err := NewMemoryBufferFromFile(file.Name())
			if err != nil {
				t.Fatal(err)
			}
			module, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatal(err)
			}
			defer module.Dispose()

			err = ConfigureGoObjFromModule(module)
			if test.wantErr == "" && err != nil {
				t.Fatal(err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("ConfigureGoObjFromModule error = %v, want substring %q", err, test.wantErr)
			}
		})
	}
}
