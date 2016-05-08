package msources

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func cbzify(folderPath string) error {
	cbzFile, err := os.Create(folderPath + ".cbz")
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(cbzFile)
	err = filepath.Walk(folderPath,
		func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil || fileInfo.IsDir() {
				return err
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			zipFileWriter, err := zipWriter.Create(filePath)
			if err != nil {
				return err
			}

			_, err = io.Copy(zipFileWriter, file)
			if err != nil {
				return err
			}

			return nil
		})
	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	err = os.RemoveAll(folderPath)
	if err != nil {
		fmt.Printf("Couldn't %v after creating cbz\n", folderPath)
		return nil
	}

	return nil
}
