package network

import (
	"bytes"
	"fmt"
	"os/exec"
)

func CheckUrl(url string) (string, error) {
	cmd := exec.Command("curl", "-sSf", url)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	if out.String() == "" {
		return "", fmt.Errorf("URL %s returned empty response", url)
	}
	return out.String(), nil
}
