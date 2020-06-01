// +build linux

package disk

import (
	"fmt"
	"../utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// Partitions returns disk partitions. returns physical devices only
// (e.g. hard disks, cd-rom drives, USB keys)
func Partitions() ([]PartitionStat, error) {
	useMounts := false

	filename := "/proc/self/mountinfo"
	lines, err := utils.ReadLines(filename)
	if err != nil {
		if err != err.(*os.PathError) {
			return nil, err
		}
		// if kernel does not support self/mountinfo, fallback to self/mounts (<2.6.26)
		useMounts = true
		filename = "/proc/self/mounts"
		lines, err = utils.ReadLines(filename)
		if err != nil {
			return nil, err
		}
	}

	fs, err := getFileSystems()
	if err != nil {
		return nil, err
	}

	ret := make([]PartitionStat, 0, len(lines))

	for _, line := range lines {
		var d PartitionStat
		if useMounts {
			fields := strings.Fields(line)

			d = PartitionStat{
				Device:     fields[0],
				Mountpoint: unescapeFstab(fields[1]),
				Fstype:     fields[2],
				Opts:       fields[3],
			}

			if d.Device == "none" || !utils.StringsHas(fs, d.Fstype) {
				continue
			}
		} else {
			// a line of self/mountinfo has the following structure:
			// 36  35  98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue
			// (1) (2) (3)   (4)   (5)      (6)      (7)   (8) (9)   (10)         (11)

			// split the mountinfo line by the separator hyphen
			parts := strings.Split(line, " - ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("found invalid mountinfo line in file %s: %s ", filename, line)
			}

			fields := strings.Fields(parts[0])
			blockDeviceID := fields[2]
			mountPoint := fields[4]
			mountOpts := fields[5]

			fields = strings.Fields(parts[1])
			fstype := fields[0]
			device := fields[1]

			d = PartitionStat{
				Device:     device,
				Mountpoint: mountPoint,
				Fstype:     fstype,
				Opts:       mountOpts,
			}

			if d.Device == "none" || !utils.StringsHas(fs, d.Fstype) {
				continue
			}

			// /dev/root is not the real device name
			// so we get the real device name from its major/minor number
			if d.Device == "/dev/root" {
				devpath, err := os.Readlink("/sys/dev/block/" + blockDeviceID)
				if err != nil {
					return nil, err
				}
				d.Device = strings.Replace(d.Device, "root", filepath.Base(devpath), 1)
			}
		}
		ret = append(ret, d)
	}

	return ret, nil
}

// getFileSystems returns supported filesystems from /proc/filesystems
func getFileSystems() ([]string, error) {
	filename := "/proc/filesystems"
	lines, err := utils.ReadLines(filename)
	if err != nil {
		return nil, err
	}
	var ret []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "nodev") {
			ret = append(ret, strings.TrimSpace(line))
			continue
		}
		t := strings.Split(line, "\t")
		if len(t) != 2 || t[1] != "zfs" {
			continue
		}
		ret = append(ret, strings.TrimSpace(t[1]))
	}

	return ret, nil
}

// Usage returns a file system usage. path is a filesystem path such
// as "/", not device file path like "/dev/vda1".  If you want to use
// a return value of disk.Partitions, use "Mountpoint" not "Device".
func Usage(path string) (*UsageStat, error) {
	stat := unix.Statfs_t{}
	err := unix.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	bsize := stat.Bsize

	ret := &UsageStat{
		Path:   unescapeFstab(path),
		Fstype: getFsType(stat),
		Total:  uint64(stat.Blocks) * uint64(bsize),
		Free:   uint64(stat.Bavail) * uint64(bsize),
	}

	ret.Used = (uint64(stat.Blocks) - uint64(stat.Bfree)) * uint64(bsize)

	if (ret.Used + ret.Free) == 0 {
		ret.UsedPercent = 0
	} else {
		ret.UsedPercent = (float64(ret.Used) / float64(ret.Used+ret.Free)) * 100.0
	}

	return ret, nil
}

// Unescape escaped octal chars (like space 040, ampersand 046 and backslash 134) to their real value in fstab fields issue#555
func unescapeFstab(path string) string {
	escaped, err := strconv.Unquote(`"` + path + `"`)
	if err != nil {
		return path
	}
	return escaped
}

func getFsType(stat unix.Statfs_t) string {
	t := int64(stat.Type)
	ret, ok := fsTypeMap[t]
	if !ok {
		return ""
	}
	return ret
}
