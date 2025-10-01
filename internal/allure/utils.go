package allure

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (a *Allure) createOutputDir() {
	isExists, err := a.existsDir(a.path)
	if err != nil {
		a.ProcessError("Error create allure json " + err.Error())
	}

	if !isExists {
		_ = os.MkdirAll(a.path, os.ModePerm)
	}

}

func (a *Allure) existsDir(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (a *Allure) deleteFiles() {
	var directory string
	var err error

	if isAbsolutePath(a.path) {
		directory = a.path
	} else {
		directory, err = os.Getwd()
		if err != nil {
			a.ProcessError(err)
			return
		}
		directory = directory + "/" + a.path
	}

	readDirectory, err := os.Open(directory)
	if err != nil {
		a.ProcessError(err)
		return
	}

	allFiles, err := readDirectory.Readdir(0)
	if err != nil {
		a.ProcessError(err)
		return
	}

	for f := range allFiles {
		file := allFiles[f]
		fileName := file.Name()
		if strings.HasSuffix(a.path+fileName, ".json") {
			if err := os.Remove(a.path + fileName); err != nil {
				if os.IsNotExist(err) {
					a.ProcessWarm(fmt.Sprintf("File %s not exist, skip\n", fileName))
					continue
				}
				continue
			}
		}
	}
}

func (a *Allure) addFileToZip(zipWriter *zip.Writer, filename string) error {
	if !strings.HasSuffix(filename, ".json") {
		return nil
	}

	file, err := os.Open(a.path + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (a *Allure) createZip(zipFileName string) {
	directory, err := os.Getwd()
	if err != nil {
		a.ProcessError(err)
		return
	}
	readDirectory, err := os.Open(directory + "/" + a.path)
	if err != nil {
		a.ProcessError(err)
		return
	}

	zipFile, err := os.Create(directory + "/" + a.path + zipFileName)
	if err != nil {
		a.ProcessError(err)
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	allFiles, err := readDirectory.Readdir(0)
	if err != nil {
		a.ProcessError(err)
		return
	}

	for _, file := range allFiles {
		if err := a.addFileToZip(zipWriter, file.Name()); err != nil {
			a.ProcessError(err)
		}
	}
}

// Функция для определения Content-Type по расширению файла
func (a *Allure) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".zip":
		return "application/zip"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	case ".rar":
		return "application/vnd.rar"
	case ".7z":
		return "application/x-7z-compressed"
	case ".tar.gz", ".tgz":
		return "application/gzip"
	case ".tar.bz2":
		return "application/x-bzip2"
	case ".jar":
		return "application/java-archive"
	case ".war":
		return "application/java-archive"
	default:
		return "application/octet-stream"
	}
}

// HasTrailingSeparator проверяет наличие разделителя в конце пути
func HasTrailingSeparator(path string) bool {
	if path == "" {
		return false
	}
	return strings.HasSuffix(path, string(filepath.Separator)) ||
		strings.HasSuffix(path, "/") // для URL и универсальности
}

// EnsureTrailingSeparator добавляет разделитель в конце пути
func EnsureTrailingSeparator(path string) string {
	if path == "" {
		return string(filepath.Separator)
	}
	if !HasTrailingSeparator(path) {
		return path + string(filepath.Separator)
	}
	return path
}

func isAbsolutePath(path string) bool {
	// Используем стандартную функцию
	if filepath.IsAbs(path) {
		return true
	}

	// Дополнительная проверка для Windows
	if runtime.GOOS == "windows" {
		// Проверяем пути вида C:\ или \\server\share
		if len(path) >= 2 {
			// Диск с двоеточием (C:\, D:\ и т.д.)
			if path[1] == ':' && (path[2] == '\\' || path[2] == '/') {
				return true
			}
			// UNC путь (\\server\share)
			if strings.HasPrefix(path, "\\\\") {
				return true
			}
		}
	}

	return false
}
