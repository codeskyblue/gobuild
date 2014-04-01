// Copyright 2013 cae authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package zip enables you to transparently read or write ZIP compressed archives and the files inside them.
package zip

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// File represents a file in archive.
type File struct {
	*zip.FileHeader
	oldName    string
	oldComment string
	absPath    string
}

// ZipArchive represents a file archive, compressed with Zip.
type ZipArchive struct {
	*zip.ReadCloser
	FileName   string
	Comment    string
	NumFiles   int
	Flag       int
	Permission os.FileMode

	files        []*File
	isHasChanged bool

	// For supporting to flush to io.Writer.
	writer      io.Writer
	isHasWriter bool
}

// Create creates the named zip file, truncating
// it if it already exists. If successful, methods on the returned
// ZipArchive can be used for I/O; the associated file descriptor has mode
// O_RDWR.
// If there is an error, it will be of type *PathError.
func Create(fileName string) (zip *ZipArchive, err error) {
	os.MkdirAll(path.Dir(fileName), os.ModePerm)
	return OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Open opens the named zip file for reading.  If successful, methods on
// the returned ZipArchive can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func Open(fileName string) (zip *ZipArchive, err error) {
	return OpenFile(fileName, os.O_RDONLY, 0)
}

// OpenFile is the generalized open call; most users will use Open
// instead. It opens the named zip file with specified flag
// (O_RDONLY etc.) if applicable. If successful,
// methods on the returned ZipArchive can be used for I/O.
// If there is an error, it will be of type *PathError.
func OpenFile(fileName string, flag int, perm os.FileMode) (zip *ZipArchive, err error) {
	zip = &ZipArchive{}
	err = zip.Open(fileName, flag, perm)
	return zip, err
}

// New accepts a variable that implemented interface io.Writer
// for write-only purpose operations.
func New(w io.Writer) (zip *ZipArchive) {
	return &ZipArchive{
		writer:      w,
		isHasWriter: true,
	}
}

func hasPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// ListName returns a string slice of files' name in ZipArchive.
func (z *ZipArchive) ListName(prefixes ...string) []string {
	isHasPrefix := len(prefixes) > 0
	names := make([]string, 0, z.NumFiles)
	for _, f := range z.files {
		if isHasPrefix {
			if hasPrefix(f.Name, prefixes) {
				names = append(names, f.Name)
			}
			continue
		}
		names = append(names, f.Name)
	}
	return names
}

// AddEmptyDir adds a directory entry to ZipArchive,
// it returns false when directory already existed.
func (z *ZipArchive) AddEmptyDir(dirPath string) bool {
	if !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}

	for _, f := range z.files {
		if dirPath == f.Name {
			return false
		}
	}

	dirPath = strings.TrimSuffix(dirPath, "/")
	if strings.Contains(dirPath, "/") {
		// Auto add all upper level directory.
		tmpPath := path.Dir(dirPath)
		z.AddEmptyDir(tmpPath)
	}
	z.files = append(z.files, &File{
		FileHeader: &zip.FileHeader{
			Name:             dirPath + "/",
			UncompressedSize: 0,
		},
	})
	z.updateStat()
	return true
}

// AddFile adds a directory and subdirectories entries to ZipArchive,
func (z *ZipArchive) AddDir(dirPath, absPath string) error {
	dir, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer dir.Close()

	z.AddEmptyDir(dirPath)

	// Get file info slice
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, fi := range fis {
		curPath := strings.Replace(absPath+"/"+fi.Name(), "\\", "/", -1)
		tmpRecPath := strings.Replace(filepath.Join(dirPath, fi.Name()), "\\", "/", -1)
		if fi.IsDir() {
			err = z.AddDir(tmpRecPath, curPath)
			if err != nil {
				return err
			}
		} else {
			err = z.AddFile(tmpRecPath, curPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// AddFile adds a file entry to ZipArchive,
func (z *ZipArchive) AddFile(fileName, absPath string) error {
	if globalFilter(absPath) {
		return nil
	}

	f, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	file := new(File)
	file.FileHeader, err = zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	file.Name = fileName
	file.absPath = absPath

	z.AddEmptyDir(path.Dir(fileName))

	isExist := false
	for _, f := range z.files {
		if fileName == f.Name {
			f = file
			isExist = true
			break
		}
	}

	if !isExist {
		z.files = append(z.files, file)
	}
	z.updateStat()
	return nil
}

// DeleteIndex deletes an entry in the archive using its index.
func (z *ZipArchive) DeleteIndex(index int) error {
	if index >= z.NumFiles {
		return errors.New("index out of range of number of files")
	}

	z.files = append(z.files[:index], z.files[index+1:]...)
	return nil
}

// DeleteName deletes an entry in the archive using its name.
func (z *ZipArchive) DeleteName(name string) error {
	for i, f := range z.files {
		if f.Name == name {
			return z.DeleteIndex(i)
		}
	}
	return errors.New("entry with given name not found")
}

func (z *ZipArchive) updateStat() {
	z.NumFiles = len(z.files)
	z.isHasChanged = true
}

// copy copies file from source to target path.
// It returns false and error when error occurs in underlying functions.
func copy(destPath, srcPath string) error {
	sf, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sf.Close()

	si, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}

	df, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer df.Close()

	// Symbolic link.
	if si.Mode()&os.ModeSymlink != 0 {
		// target, err := os.Readlink(srcPath)
		// if err != nil {
		// 	return err
		// }

		// _, stderr, _ := com.ExecCmd("ln", "-s", target, destPath)
		// if len(stderr) > 0 {
		// 	return errors.New(stderr)
		// }
		// return nil
		return errors.New("Go hasn't support read symbolic link itself yet!")
	}

	// buffer reader, do chunk transfer
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := sf.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		// write a chunk
		if _, err := df.Write(buf[:n]); err != nil {
			return err
		}
	}

	return os.Chmod(destPath, si.Mode())
}

func globalFilter(name string) bool {
	if strings.Contains(name, ".DS_Store") {
		return true
	}
	return false
}
