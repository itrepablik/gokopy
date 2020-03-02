/*
Copyright © 2020 ITRepablik <support@itrepablik.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/itrepablik/itrlog"
	"github.com/itrepablik/kopy"

	"github.com/spf13/cobra"
)

// comdirCmd represents the comdir command
var comdirCmd = &cobra.Command{
	Use:   "comdir",
	Short: "Compress the entire directory or a folder",
	Long: `comdir command will compress the specified directory or a folder including the sub-folders and its contents as well
using .tar.gz compression format.

Example of a valid directory path in Windows:
"C:\source_folder_to_compress" "D:\backup_destination"

Or using the network directories, example:
"\\hostname_or_ip\source_folder_to_compress" "\\hostname_or_ip\backup_destination"

Or in Linux:
"/root/src" "/root/dst"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		src := filepath.FromSlash(args[0])
		dst := filepath.FromSlash(args[1])

		msg := `Start compressing the directory or a folder:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "dst", dst, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		// Compose the zip filename
		fnWOext := kopy.FileNameWOExt(filepath.Base(args[0])) // Returns a filename without an extension.
		zipDir := fnWOext + kopy.ComFileFormat

		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		zipDest := filepath.FromSlash(path.Join(args[1], zipDir))

		// Start compressing the entire directory or a folder using the tar + gzip
		var buf bytes.Buffer
		if err := kopy.CompressDIR(src, &buf, IgnoreFT); err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		// write the .tar.gzip
		os.MkdirAll(dst, os.ModePerm) // Create the root folder first
		fileToWrite, err := os.OpenFile(zipDest, os.O_CREATE|os.O_RDWR, os.FileMode(600))
		if err != nil {
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			panic(err)
		}
		if _, err := io.Copy(fileToWrite, &buf); err != nil {
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			panic(err)
		}
		defer fileToWrite.Close()

		msg = `Done compressing the directory or a folder:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "dst", zipDest, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(comdirCmd)
}
