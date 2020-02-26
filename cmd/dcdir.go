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
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// dcdirCmd represents the dcdir command
var dcdirCmd = &cobra.Command{
	Use:   "dcdir",
	Short: "Decompress any single tar.gz file",
	Long: `The dcdir command will decompress the specified directory or a folder including the sub-folders
and its contents as well. Only the .tar.gz compress files will be extracted by this dcdir command in relation
to the comdir command in which it will compress the entire folder.

Example of a valid directory path in Windows:
"C:\source_folder\foldername.tar.gz"

Or using the network directories, example:
"\\hostname_or_ip\source_folder\foldername.tar.gz"

Example of a valid directory path in Linux:
"/home/user/source_folder_to_compress/foldername.tar.gz".`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		src := filepath.FromSlash(args[0])

		msg := `Start decompressing the folder or a directory:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		r, err := os.Open(src)
		if err != nil {
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			fmt.Println("error")
			return
		}
		if err := kopy.ExtractTarGz(r, src, IsLogCopiedFile); err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		msg = `Done decompressing the folder or a directory:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(dcdirCmd)
}
