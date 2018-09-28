package utils

import (
	"archive/tar"
	"io"
	"path/filepath"
	"fmt"
	"os"
)

func UnpackTarIntoDirectory(reader *tar.Reader, dir string) error {
	for {
		// Read the next entry in the file
		header, err := reader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		// Check that the entry is well-formed
		target := filepath.Join(dir, header.Name)
		absTarget, err := filepath.Abs(target)
		if !filepath.HasPrefix(absTarget, dir) {
			return fmt.Errorf("DirectoryUnpackSecurityError: expected prefix with '{}', have '{}'", dir, absTarget)
		}
		target = absTarget
		// Do something
		switch header.Typeflag {
		case tar.TypeDir:
			permBits := header.FileInfo().Mode() & 0x1F
			if err := os.MkdirAll(absTarget, permBits); err != nil {
				return fmt.Errorf("UnpackError: could not create directory at '{}' (error was '{}')", absTarget, err)
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, reader); err != nil {
				return fmt.Errorf("UnpackError: could not create file at '{}' (error was '{}')", absTarget, err)
			}
		}
	}
	return nil
}
