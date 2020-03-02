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
	"path/filepath"
	"time"

	"github.com/itrepablik/itrlog"
	"github.com/itrepablik/kopy"

	"github.com/spf13/cobra"
)

// copyfileCmd represents the copyfile command
var copyfileCmd = &cobra.Command{
	Use:   "copyfile",
	Short: "Copy a single file without a compression",
	Long: `copyfile command is to copy the individual or a specific file from a valid source folder or a directory.
Take note that, it will replace the existing file and its contents to the destination file.

It must have a valid and absolute path for the source and its destination folder or directory.
The Source and Destination paths should contains the "" space "" characters with one space in between to separate them.

Example of a valid directory path in Windows:
"C:\source_folder\filename.txt" "D:\backup_destination"

Or using the network directories, example:
"\\hostname_or_ip\source_folder\filename.txt" "\\hostname_or_ip\backup_destination"

Or in Linux:
"/root/src/file.txt" "/root/dst"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		src := filepath.FromSlash(args[0])
		dst := filepath.FromSlash(args[1])

		msg := `Starts copying the single file:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "dst", dst, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		dest := filepath.FromSlash(filepath.Join(args[1], filepath.Base(src)))

		// Starts copying the single file.
		if err := kopy.CopyFile(src, dest, dst, Sugar); err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		// Give some info back to the user's console and the logs as well.
		msg = `Successfully copied the file:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "dst", dest, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(copyfileCmd)
}
