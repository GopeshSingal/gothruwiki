package wiki

import (
	"compress/bzip2"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type DecompressMode string

const (
	DecompressAuto   DecompressMode = "auto"
	DecompressStdlib DecompressMode = "stdlib"
	DecompressLbzip2 DecompressMode = "lbzip2"
	// DecompressPbzip2 DecompressMode = "pbzip2"
)

func OpenBZ2(path string, mode DecompressMode) (r io.ReadCloser, label string, err error) {
	switch mode {
	case DecompressStdlib:
		return openBZ2Stdlib(path)
	case DecompressLbzip2:
		return openBZ2External("lbzip2", path)
	// case DecompressPbzip2:
	// 	return openBZ2External("pbzip2", path)
	case DecompressAuto, "":
		if _, e := exec.LookPath("lbzip2"); e == nil {
			return openBZ2External("lbzip2", path)
		}
		// if _, e := exec.LookPath("pbzip2"); e == nil {
		// 	return openBZ2External("pbzip2", path)
		// }
		return openBZ2Stdlib(path)
	default:
		return nil, "", fmt.Errorf("unknown -decompress mode %q", mode)
	}
}

func openBZ2Stdlib(path string) (io.ReadCloser, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	return &stdlibBZ2{f: f, r: bzip2.NewReader(f)}, "stdlib compress/bzip2", nil
}

type stdlibBZ2 struct {
	f *os.File
	r io.Reader
}

func (s *stdlibBZ2) Read(p []byte) (int, error) { return s.r.Read(p) }

func (s *stdlibBZ2) Close() error { return s.f.Close() }

func openBZ2External(bin, path string) (io.ReadCloser, string, error) {
	cmd := exec.Command(bin, "-dc", path)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", err
	}
	if err := cmd.Start(); err != nil {
		return nil, "", err
	}
	return &externalBZ2{cmd: cmd, ReadCloser: stdout}, bin + " -dc", nil
}

type externalBZ2 struct {
	cmd *exec.Cmd
	io.ReadCloser
}

func (e *externalBZ2) Close() error {
	var err error
	if e.ReadCloser != nil {
		err = e.ReadCloser.Close()
	}
	if e.cmd != nil {
		_ = e.cmd.Wait()
	}
	return err
}
