package output

import (
	"bytes"
	"io"
	"testing"
)

func TestBulletPrint(t *testing.T) {
	tests := []struct {
		name      string
		verbosity Verbosity
		msg       string
		want      string
	}{
		{
			name:      "quiet mode prints bullet message",
			verbosity: Quiet,
			msg:       "Using cached image: littlebox/base:0.1.0",
			want:      "● Using cached image: littlebox/base:0.1.0\r\n",
		},
		{
			name:      "debug mode prints nothing",
			verbosity: Debug,
			msg:       "Using cached image: littlebox/base:0.1.0",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			BulletPrint(&buf, tt.verbosity, tt.msg)
			if got := buf.String(); got != tt.want {
				t.Errorf("BulletPrint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatusWriter_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStatusWriter(&buf, Quiet)

	sw.Start("Building image: littlebox/base:0.1.0, this may take a while")
	if got := buf.String(); got != "● Building image: littlebox/base:0.1.0, this may take a while..." {
		t.Errorf("Start() = %q, want %q", got, "● Building image: littlebox/base:0.1.0, this may take a while...")
	}

	sw.Done()
	want := "● Building image: littlebox/base:0.1.0, this may take a while... done\r\n"
	if got := buf.String(); got != want {
		t.Errorf("Done() = %q, want %q", got, want)
	}
}

func TestStatusWriter_DebugMode(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStatusWriter(&buf, Debug)

	sw.Start("Building image: littlebox/base:0.1.0, this may take a while")
	sw.Done()

	if got := buf.String(); got != "" {
		t.Errorf("StatusWriter in debug mode printed %q, want empty string", got)
	}
}

func TestStatusWriter_MultipleCycles(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStatusWriter(&buf, Quiet)

	// First cycle
	sw.Start("Setting up the sandbox")
	sw.Done()

	// Second cycle
	sw.Start("Building image")
	sw.Done()

	want := "● Setting up the sandbox... done\r\n● Building image... done\r\n"
	if got := buf.String(); got != want {
		t.Errorf("Multiple cycles = %q, want %q", got, want)
	}
}

func TestStatusWriter_DoneWithoutStart(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStatusWriter(&buf, Quiet)

	// Done without Start should do nothing
	sw.Done()

	if got := buf.String(); got != "" {
		t.Errorf("Done() without Start() printed %q, want empty string", got)
	}
}

func TestWriter(t *testing.T) {
	tests := []struct {
		name        string
		verbosity   Verbosity
		wantDiscard bool
	}{
		{
			name:        "quiet mode returns discard",
			verbosity:   Quiet,
			wantDiscard: true,
		},
		{
			name:        "debug mode returns original writer",
			verbosity:   Debug,
			wantDiscard: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := Writer(&buf, tt.verbosity)

			// Write something to test
			_, _ = w.Write([]byte("test"))

			if tt.wantDiscard {
				// In quiet mode, buffer should be empty (output discarded)
				if buf.Len() != 0 {
					t.Errorf("Writer() in quiet mode wrote to buffer, want discard")
				}
				// Verify it's actually io.Discard
				if w != io.Discard {
					t.Errorf("Writer() in quiet mode = %v, want io.Discard", w)
				}
			} else {
				// In debug mode, buffer should contain the write
				if got := buf.String(); got != "test" {
					t.Errorf("Writer() in debug mode = %q, want %q", got, "test")
				}
			}
		})
	}
}
