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
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// dcfileCmd represents the dcfile command
var dcfileCmd = &cobra.Command{
	Use:   "dcfile",
	Short: "Decompress any single zip file",
	Long: `dcfile command will decompress any single zip compression file format.
Only the .zip files will be decompressed by this command.

Example of a valid directory path in Windows:
"C:\source_folder\filename.zip"

Or using the network directories, example:
"\\hostname_or_ip\source_folder\filename.zip"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		src := filepath.FromSlash(args[0])

		msg := `Start decompressing the file:`
		fmt.Println(msg, src)
		Sugar.Errorw(msg, "src", src, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		if err := kopy.Unzip(src, IsLogCopiedFile); err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		msg = `Done decompressing the file:`
		fmt.Println(msg, src)
		Sugar.Errorw(msg, "src", src, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(dcfileCmd)
}
