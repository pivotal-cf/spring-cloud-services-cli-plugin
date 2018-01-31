package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func GetRelativePath(path string) string {
	cwd, err := os.Getwd()
	relPath, err := filepath.Rel(cwd, path)
	check(err)
	return relPath
}

func CreateFile(path string, fileName string) string {
	return CreateFileWithMode(path, fileName, os.FileMode(0666))
}

func CreateFileWithMode(path string, fileName string, mode os.FileMode) string {
	fp := filepath.Join(path, fileName)
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	check(err)
	defer f.Close()
	_, err = f.WriteString("Hello\nWorld\n")
	check(err)
	return fp
}

func CreateTempDir() string {
	tempDir, err := ioutil.TempDir("/tmp", "scs-cli-")
	check(err)
	return tempDir
}
