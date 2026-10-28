// Package printing hands finished PDFs to the operating system: CUPS (lp)
// on Linux and macOS, or the default PDF viewer anywhere.
package printing

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Printer struct {
	Name    string `json:"name"`
	Default bool   `json:"default"`
}

// DirectSupported reports whether this machine can send jobs with lp.
func DirectSupported() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	_, err := exec.LookPath("lp")
	return err == nil
}

// Printers lists CUPS destinations.
func Printers() ([]Printer, error) {
	if !DirectSupported() {
		return nil, nil
	}
	out, err := exec.Command("lpstat", "-e").Output()
	if err != nil {
		return nil, fmt.Errorf("list printers: %w", err)
	}
	def := defaultPrinter()
	var list []Printer
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		if name := strings.TrimSpace(sc.Text()); name != "" {
			list = append(list, Printer{Name: name, Default: name == def})
		}
	}
	return list, nil
}

func defaultPrinter() string {
	out, err := exec.Command("lpstat", "-d").Output()
	if err != nil {
		return ""
	}
	// "system default destination: Canon_G6000_series"
	_, name, ok := strings.Cut(string(out), ":")
	if !ok {
		return ""
	}
	return strings.TrimSpace(name)
}

// Job is one submission to the printer.
type Job struct {
	Printer  string
	File     string
	Media    string // CUPS media name, e.g. "Legal"
	TwoSided bool   // flip on the short edge
}

// Print sends the job with lp. The PDF already has the exact sheet layout, so
// scaling is disabled.
func Print(j Job) error {
	args := []string{
		"-o", "media=" + j.Media,
		"-o", "print-scaling=none",
		"-o", "fit-to-page=false",
	}
	if j.TwoSided {
		args = append(args, "-o", "sides=two-sided-short-edge")
	} else {
		args = append(args, "-o", "sides=one-sided")
	}
	if j.Printer != "" {
		args = append(args, "-d", j.Printer)
	}
	args = append(args, "--", j.File)

	out, err := exec.Command("lp", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("lp: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Open shows the file in the system's default application.
func Open(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// MediaName maps a paper id to its CUPS media keyword.
func MediaName(paperID string) string {
	switch paperID {
	case "letter":
		return "Letter"
	case "legal":
		return "Legal"
	case "a4":
		return "A4"
	case "a5":
		return "A5"
	}
	return ""
}
