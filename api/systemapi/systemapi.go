package systemapi

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/EntropyPool/entropy-logger"
	"github.com/moby/sys/mountinfo"
)

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func FilePerm2Int(file string) (int, error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, err
	}
	strMode := fmt.Sprintf("%o", fi.Mode().Perm())
	mode, _ := strconv.ParseInt(strMode, 64)
	return mode, nil
}

func MountpointWrittable(mountpoint string) (bool, error) {
	var info *mountinfo.Info = nil

	_, err := mountinfo.GetMounts(func(i *mountinfo.Info) (skip, stop bool) {
		if info.Mountpoint == mountpoint {
			info = i
			return false, true
		}
		return false, false
	})

	if err != nil {
		return false, err
	}

	if info != nil {
		if strings.TrimSpace(strings.Split(info.Options, ",")[0]) == "rw" {
			return true, nil
		}
		return false, nil
	}

	return false, xerrors.Errorf("mountpoint %v not found", mountpoint)
}

func StatSubDirs(dir string, sublevel int) map[string]error {
	stat := map[string]error{}
	mySlashes := strings.Count(dir, "/")

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if strings.Count(path, "/") == sublevel+mySlashes {
			stat[path] = err
		}
		return nil
	})

	return stat
}
