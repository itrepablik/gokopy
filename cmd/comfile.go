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
	"fmt"
	"gokopy/itrlog"
	"gokopy/kopy"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// comfileCmd represents the kom command
var comfileCmd = &cobra.Command{
	Use:   "comfile",
	Short: "Compress any single file",
	Long: `comfile command will compress any single file using .zip compression format.

Example of a valid directory path in Windows:
"C:\source_folder\filename.txt" "D:\backup_destination"

Or using the network directories, example:
"\\hostname_or_ip\source_folder\filename.txt" "\\hostname_or_ip\backup_destination"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// Use this function to auto detect file path structure.
		src := filepath.FromSlash(args[0])
		dst := filepath.FromSlash(args[1])

		// Start the process.
		msg := `Start compressing the file:`
		fmt.Println(msg, src)
		Sugar.Errorw(msg, "src", src, "dst", dst, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		// Compose the zip filename
		fnWOext := kopy.FileNameWOExt(filepath.Base(args[0])) // Returns a filename without an extension.
		zipFileName := fnWOext + ".zip"

		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		zipDest := filepath.FromSlash(path.Join(args[1], zipFileName))

		// List of Files to compressed.
		files := []string{src}

		os.MkdirAll(dst, os.ModePerm) // Create the root folder first
		if err := kopy.ComFiles(zipDest, files); err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		msg = `Done compressing the file:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "dst", zipDest, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(comfileCmd)
}
