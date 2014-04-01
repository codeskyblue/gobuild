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

package zip

import (
	"archive/zip"
	"os"
	"strings"
)

// OpenFile is the generalized open call; most users will use Open
// instead. It opens the named zip file with specified flag
// (O_RDONLY etc.) if applicable. If successful,
// methods on the returned ZipArchive can be used for I/O.
// If there is an error, it will be of type *PathError.
func (z *ZipArchive) Open(fileName string, flag int, perm os.FileMode) error {
	if flag&os.O_CREATE != 0 {
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}
		zw := zip.NewWriter(f)
		zw.Close()
	}

	rc, err := zip.OpenReader(fileName)
	if err != nil {
		return err
	}

	z.ReadCloser = rc
	z.FileName = fileName
	z.Comment = rc.Comment
	z.NumFiles = len(rc.File)
	z.Flag = flag
	z.Permission = perm
	z.isHasChanged = false

	z.files = make([]*File, z.NumFiles)
	for i, f := range rc.File {
		z.files[i] = &File{}
		z.files[i].FileHeader, err = zip.FileInfoHeader(f.FileInfo())
		if err != nil {
			return err
		}
		z.files[i].Name = strings.Replace(f.Name, "\\", "/", -1)
		if f.FileInfo().IsDir() && !strings.HasSuffix(z.files[i].Name, "/") {
			z.files[i].Name += "/"
		}
	}
	return nil
}
