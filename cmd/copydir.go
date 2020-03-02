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

// copydirCmd represents the copydir command
var copydirCmd = &cobra.Command{
	Use:   "copydir",
	Short: "Copy the entire folder or a directory without a compression",
	Long: `copydir command is to copy the entire directory or folder including its sub-folders and sub-directories contents.
Take note that, it will replace any existing files and its contents to the destination directory or a folder.

It must have a valid and absolute path for the source and its destination folder or directory.
The Source and Destination paths should contains the "" space "" characters with one space in between to separate them.

Example of a valid directory path in Windows:
"C:\source_folder" "D:\backup_destination"

Or using the network directories, example:
"\\hostname_or_ip\source_folder" "\\hostname_or_ip\backup_destination"

Or in Linux:
"/root/src" "/root/dst"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		src := filepath.FromSlash(args[0])
		dst := filepath.FromSlash(args[1])

		msg := `Starts copying the entire directory or a folder: `
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		// Starts copying the entire directory or a folder.
		filesCopied, foldersCopied, err := kopy.CopyDir(src, dst, IsLogCopiedFile, IgnoreFT, Sugar)
		if err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		// Give some info back to the user's console and the logs as well.
		msg = `Successfully copied the entire directory or a folder: `
		fmt.Println(msg, src, ", Number of Folders Copied: ", filesCopied, " Number of Files Copied: ", foldersCopied)
		Sugar.Infow(msg, "src", src, "dst", dst, "folder_copied", filesCopied, "files_copied", foldersCopied, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(copydirCmd)
}
