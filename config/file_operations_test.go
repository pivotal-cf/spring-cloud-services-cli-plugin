package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Encrypt file", func() {

	var (
		testFile string
		testDir  string
	)

	JustBeforeEach(func() {
		testDir = CreateTempDir()
		testFile = CreateFile(testDir, "file-to-encrypt.txt")
	})

	It("should read file with absolute path", func() {
		contents, err := ReadFileContents(testFile)
		Expect(contents).To(Equal("Hello\nWorld\n"))
		Expect(err).To(BeNil())
	})

	It("should read file with relative path", func() {
		relPath := GetRelativePath(testFile)
		contents, err := ReadFileContents(relPath)
		Expect(contents).To(Equal("Hello\nWorld\n"))
		Expect(err).To(BeNil())
	})

	It("should fail for non-existent file", func() {
		contents, err := ReadFileContents("bogus.txt")
		Expect(contents).To(Equal(""))
		Expect(err).To(Equal(fmt.Errorf("Error opening file at path bogus.txt : open bogus.txt: no such file or directory")))
	})

	It("should fail for directory", func() {
		contents, err := ReadFileContents(testDir)
		Expect(contents).To(Equal(""))
		Expect(err).To(Equal(fmt.Errorf("Error opening file at path %s : read %s: is a directory", testDir, testDir)))
	})

})

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
