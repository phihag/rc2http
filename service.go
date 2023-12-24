package main

import (
	_ "embed"
	"github.com/cross-cpm/go-shutil"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed init-file
var INIT_FILE string

func InstallService() {
	targetPath := "/usr/local/bin/rc2http"
	initPath := "/etc/init.d/rc2http"

	err := os.MkdirAll(filepath.Dir(targetPath), 0755)
	FatalError(err)

	executable, err := os.Executable()
	FatalError(err)
	exePath, err := filepath.Abs(executable)
	FatalError(err)
	if exePath != targetPath {
		_, err := shutil.CopyFile(exePath, targetPath, &shutil.CopyOptions{})
		FatalError(err)
		err = os.Chmod(targetPath, 0755)
		FatalError(err)
	}

	err = os.WriteFile(initPath, []byte(INIT_FILE), 0644)
	FatalError(err)
	err = os.Chmod(initPath, 0755)
	FatalError(err)

	cmd := exec.Command("update-rc.d", "rc2http", "defaults")
	err = cmd.Run()
	FatalError(err)
}
