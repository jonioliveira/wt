package ui_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jonioliveira/wt/internal/ui"
)

func TestSyncResult_PrintSummary_CopiedOnly(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 5}.PrintSummary(&buf)
	out := buf.String()

	if !strings.Contains(out, "5 copied") {
		t.Errorf("want '5 copied' in output, got: %q", out)
	}
	if strings.Contains(out, "skipped") {
		t.Errorf("want no 'skipped' in output, got: %q", out)
	}
	if strings.Contains(out, "failed") {
		t.Errorf("want no 'failed' in output, got: %q", out)
	}
}

func TestSyncResult_PrintSummary_WithSkipped(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 3, Skipped: 2}.PrintSummary(&buf)
	out := buf.String()

	if !strings.Contains(out, "3 copied") {
		t.Errorf("want '3 copied' in output, got: %q", out)
	}
	if !strings.Contains(out, "2 skipped") {
		t.Errorf("want '2 skipped' in output, got: %q", out)
	}
	if strings.Contains(out, "failed") {
		t.Errorf("want no 'failed' in output, got: %q", out)
	}
}

func TestSyncResult_PrintSummary_WithFailed(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 1, Failed: 3}.PrintSummary(&buf)
	out := buf.String()

	if !strings.Contains(out, "1 copied") {
		t.Errorf("want '1 copied' in output, got: %q", out)
	}
	if !strings.Contains(out, "3 failed") {
		t.Errorf("want '3 failed' in output, got: %q", out)
	}
	if strings.Contains(out, "skipped") {
		t.Errorf("want no 'skipped' in output, got: %q", out)
	}
}

func TestSyncResult_PrintSummary_AllThree(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 4, Skipped: 2, Failed: 1}.PrintSummary(&buf)
	out := buf.String()

	for _, want := range []string{"4 copied", "2 skipped", "1 failed"} {
		if !strings.Contains(out, want) {
			t.Errorf("want %q in output, got: %q", want, out)
		}
	}
}

func TestSyncResult_PrintSummary_ZeroSkippedOmitted(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 2, Skipped: 0}.PrintSummary(&buf)

	if strings.Contains(buf.String(), "skipped") {
		t.Errorf("want '0 skipped' omitted, got: %q", buf.String())
	}
}

func TestSyncResult_PrintSummary_ZeroFailedOmitted(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 2, Failed: 0}.PrintSummary(&buf)

	if strings.Contains(buf.String(), "failed") {
		t.Errorf("want '0 failed' omitted, got: %q", buf.String())
	}
}

func TestSyncResult_PrintSummary_EndsWithNewline(t *testing.T) {
	var buf bytes.Buffer
	ui.SyncResult{Copied: 1}.PrintSummary(&buf)

	if !strings.HasSuffix(buf.String(), "\n") {
		t.Errorf("PrintSummary should end with a newline, got: %q", buf.String())
	}
}
