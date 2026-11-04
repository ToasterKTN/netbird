//go:build !qrcode

package cmd

import (
	"fmt"
	"io"
)

// printQRCode explains that QR output is not compiled in; the URL itself is
// still printed by the caller.
func printQRCode(w io.Writer, url string) {
	if url == "" {
		return
	}
	_, _ = fmt.Fprintln(w, "QR code output not compiled in, rebuild with -tags qrcode")
}
