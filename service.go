package main

import (
	_ "embed"
	"fmt"
	"github.com/cross-cpm/go-shutil"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed init-file
var INIT_FILE string

func fatalError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func InstallService() {
	targetPath := "/usr/local/bin/rc2http"
	initPath := "/etc/init.d/rc2http"

	err := os.MkdirAll(filepath.Dir(targetPath), 0755)
	fatalError(err)

	executable, err := os.Executable()
	fatalError(err)
	exePath, err := filepath.Abs(executable)
	fatalError(err)
	if exePath != targetPath {
		_, err := shutil.CopyFile(exePath, targetPath, &shutil.CopyOptions{})
		fatalError(err)
		err = os.Chmod(targetPath, 0755)
		fatalError(err)
	}

	err = os.WriteFile(initPath, []byte(INIT_FILE), 0644)
	fatalError(err)
	err = os.Chmod(initPath, 0755)
	fatalError(err)

	cmd := exec.Command("update-rc.d", "rc2http", "defaults")
	err = cmd.Run()
	fatalError(err)
}
