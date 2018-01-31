package config

import (
	"fmt"
	"io/ioutil"
)

func ReadFileContents(fileToEncrypt string) (string, error) {
	var dat, err = ioutil.ReadFile(fileToEncrypt)
	if err != nil {
		return "", fmt.Errorf("Error opening file at path %s : %s", fileToEncrypt, err)
	}
	return string(dat), nil
}
