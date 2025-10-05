package binarydist

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
)

// applyPatch is just a wrapper of Patch for testing
func applyPatch(old io.Reader, new io.Writer, patch io.Reader) error {
	return Patch(old, new, patch)
}

func testFunc(t *testing.T, fDiff func(old, new io.Reader, patch io.Writer) error) {
	t.Helper()

	diffT := []struct {
		name string
		old  *os.File
		new  *os.File
	}{
		{
			name: "sample data",
			old:  mustOpen("testdata/sample.old"),
			new:  mustOpen("testdata/sample.new"),
		},
		{
			name: "random data",
			old:  mustWriteRandFile("test.old", 1e3, 1),
			new:  mustWriteRandFile("test.new", 1e3, 2),
		},
	}

	for _, s := range diffT {
		t.Run(s.name, func(t *testing.T) {
			// Use bsdiff CLI to generate a patch file (expected patch)
			exp, err := os.CreateTemp("", "bspatch.")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(exp.Name())

			cmd := exec.Command("bsdiff", s.old.Name(), s.new.Name(), exp.Name())
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				t.Fatal(err)
			}

			// Use our `fDiff` implementation to generate a patch (got)
			got, err := os.CreateTemp("", "bspatch.")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(got.Name())

			if err = fDiff(s.old, s.new, got); err != nil {
				t.Fatalf("error running diff function: %s", err)
			}

			// Read in test data old/expected - so we can compare the results of applying the patch
			oldData := mustReadAll(mustOpen(s.old.Name()))
			newData := mustReadAll(mustOpen(s.new.Name())) // (the expected result)

			// Apply 'got' patch
			got.Seek(0, 0)
			appliedGot := new(bytes.Buffer)
			if err := applyPatch(bytes.NewReader(oldData), appliedGot, got); err != nil {
				t.Fatalf("patch apply failed (got): %s", err)
			}

			// Apply 'exp' patch
			exp.Seek(0, 0)
			appliedExp := new(bytes.Buffer)
			if err := applyPatch(bytes.NewReader(oldData), appliedExp, exp); err != nil {
				t.Fatalf("patch apply failed (exp): %s", err)
			}

			// appliedGot / appliedExp and newData should all be equal
			if !bytes.Equal(appliedExp.Bytes(), newData) {
				t.Fatalf("expected patch did not reproduce new file")
			}
			if !bytes.Equal(appliedGot.Bytes(), newData) {
				t.Fatalf("our patch did not reproduce new file")
			}
		})
	}
}

func TestDiff(t *testing.T) {
	testFunc(t, Diff)
}

func TestDiffWithSuf(t *testing.T) {
	fDiff := func(old, new io.Reader, patch io.Writer) error {
		oldSufStruct, err := ComputeSuf(old)
		if err != nil {
			return err
		}

		return DiffWithSuf(oldSufStruct, new, patch)
	}
	testFunc(t, fDiff)
}
