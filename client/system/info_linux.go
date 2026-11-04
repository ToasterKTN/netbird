//go:build !android

package system

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"

	"github.com/netbirdio/netbird/version"
)

var (
	// it is override in tests
	getSystemInfo = defaultSysInfoImplementation
)

func UpdateStaticInfoAsync() {
	go updateStaticInfo()
}

// GetInfo retrieves and parses the system information
func GetInfo(ctx context.Context) *Info {
	kernelName, kernelVersion, kernelPlatform := kernelInfo()

	osName, osVersion := readOsReleaseFile()
	if osName == "" {
		osName = kernelName
	}

	systemHostname, _ := os.Hostname()

	addrs, err := networkAddresses()
	if err != nil {
		log.Warnf("failed to discover network addresses: %s", err)
	}

	start := time.Now()
	si := getStaticInfo()
	if time.Since(start) > 1*time.Second {
		log.Warnf("updateStaticInfo took %s", time.Since(start))
	}

	gio := &Info{
		Kernel:             kernelName,
		Platform:           kernelPlatform,
		OS:                 osName,
		OSVersion:          osVersion,
		Hostname:           extractDeviceName(ctx, systemHostname),
		GoOS:               runtime.GOOS,
		CPUs:               runtime.NumCPU(),
		NetbirdVersion:     version.NetbirdVersion(),
		UIVersion:          extractUserAgent(ctx),
		KernelVersion:      kernelVersion,
		NetworkAddresses:   addrs,
		SystemSerialNumber: si.SystemSerialNumber,
		SystemProductName:  si.SystemProductName,
		SystemManufacturer: si.SystemManufacturer,
		Environment:        si.Environment,
	}

	return gio
}

func kernelInfo() (string, string, string) {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return "", "", ""
	}
	return unix.ByteSliceToString(uts.Sysname[:]), unix.ByteSliceToString(uts.Release[:]), unix.ByteSliceToString(uts.Machine[:])
}

func sysInfo() (string, string, string) {
	isascii := regexp.MustCompile("^[[:ascii:]]+$")
	si := getSystemInfo()
	serials := []string{si.ChassisSerial, si.ProductSerial}
	serial := ""

	for _, s := range serials {
		if isascii.MatchString(s) {
			serial = s
			if s != "Default string" {
				break
			}
		}
	}

	if serial == "" && isascii.MatchString(si.BoardSerial) {
		serial = si.BoardSerial
	}

	var name string
	for _, n := range []string{si.ProductName, si.BoardName} {
		if isascii.MatchString(n) {
			name = n
			break
		}
	}

	var manufacturer string
	if isascii.MatchString(si.ProductVendor) {
		manufacturer = si.ProductVendor
	}
	return serial, name, manufacturer
}

func defaultSysInfoImplementation() SysInfo {
	return readSysInfo(dmiReader("/sys/class/dmi/id", "/proc/device-tree"))
}

// dmiReader returns a reader for DMI attribute files, falling back to the
// device tree (ARM boards without DMI) for the product name and serial.
func dmiReader(dmiDir, deviceTreeDir string) func(dmiName, dtName string) string {
	return func(dmiName, dtName string) string {
		if v := readTrimmed(filepath.Join(dmiDir, dmiName)); v != "" || dtName == "" {
			return v
		}
		return readTrimmed(filepath.Join(deviceTreeDir, dtName))
	}
}

func readSysInfo(read func(dmiName, dtName string) string) SysInfo {
	return SysInfo{
		ChassisSerial: read("chassis_serial", ""),
		ProductSerial: read("product_serial", "serial-number"),
		BoardSerial:   read("board_serial", ""),
		ProductName:   read("product_name", "model"),
		BoardName:     read("board_name", ""),
		ProductVendor: read("sys_vendor", ""),
	}
}

// readTrimmed returns the file content without surrounding whitespace, NUL or
// U+FFFF, or "" when it cannot be read (e.g. root-only serial files).
func readTrimmed(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.Trim(string(data), " \r\n\t\u0000\uffff")
}
