package console

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/ericogr/ads1115-to-mqtt/pkg/sensor"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()
	f()
	_ = w.Close()
	os.Stdout = stdout
	return <-outC
}

func TestConsolePublish(t *testing.T) {
	c := NewConsole()
	ts := time.Date(2025, 9, 19, 14, 41, 54, 0, time.UTC)
	readings := []sensor.Reading{{Channel: 0, Raw: 123, Value: 1.234567, Timestamp: ts}}
	out := captureStdout(func() { _ = c.Publish(readings) })
	want := "2025-09-19T14:41:54Z channel=0 raw=123 value=1.234567\n"
	if out != want {
		t.Fatalf("console output mismatch:\n got: %q\nwant: %q", out, want)
	}
}
