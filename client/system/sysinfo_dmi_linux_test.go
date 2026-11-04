//go:build !android

package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadSysInfoFromDMI(t *testing.T) {
	dmi, dt := t.TempDir(), t.TempDir()
	write := func(dir, name, v string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(v), 0o600))
	}
	write(dmi, "chassis_serial", "CH-1\n")
	write(dmi, "board_serial", " BS-2 \n")
	write(dmi, "board_name", "Board\x00")
	write(dmi, "sys_vendor", "Vendor\n")
	write(dt, "model", "Raspberry Pi 4\x00")
	write(dt, "serial-number", "10000000abc\x00")

	got := readSysInfo(dmiReader(dmi, dt))
	require.Equal(t, SysInfo{
		ChassisSerial: "CH-1",
		ProductSerial: "10000000abc",
		BoardSerial:   "BS-2",
		ProductName:   "Raspberry Pi 4",
		BoardName:     "Board",
		ProductVendor: "Vendor",
	}, got)

	write(dmi, "product_name", "ThinkPad\n")
	require.Equal(t, "ThinkPad", readSysInfo(dmiReader(dmi, dt)).ProductName)
}
