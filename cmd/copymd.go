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
	"strings"
	"time"

	"github.com/itrepablik/itrlog"
	"github.com/itrepablik/kopy"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// copymdCmd represents the copymd command
var copymdCmd = &cobra.Command{
	Use:   "copymd",
	Short: "Copy the latest files from a specified directory based on the modified date and time",
	Long: `copymd command will copy the latest files including the sub-folders files based on the modified date and time from the specified folder.

Open the "config.yaml" configuration file, you can change the following default settings such as:

default:
	copy_mod_files_num_days: -7 # This must be a negative value interpreted as the previous days to start copying the files.
	
	ignore:
		file_type_or_folder_name: .db, folder_name # You can specify file extentions or folder name seperated with comma.

Example of a valid directory path in Windows:
"C:\source_folder_to_compress" "D:\backup_destination"

Or using the network directories, example:

"\\hostname_or_ip\source_folder_to_compress" "\\hostname_or_ip\backup_destination"

Or in Linux:
"/root/src" "/root/dst"`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		NumFilesCopied = 0 // Reset this variable

		// Get the default value for the "copy_mod_files_num_days" setting.
		modDays := viper.Get("default.copy_mod_files_num_days")
		mDays := modDays.(int)
		if _, ok := modDays.(int); !ok {
			mDays = -1
		}

		// Get the list of ignored file types.
		IgnoreFileTypes = viper.Get("ignore.file_type_or_folder_name")
		IGFT := fmt.Sprint(IgnoreFileTypes)
		IgnoreFT = strings.Split(IGFT, ",")

		// To make directory path separator a universal, in Linux "/" and in Windows "\" to auto change
		// depends on the user's OS using the filepath.FromSlash organic Go's library.
		src := filepath.FromSlash(args[0])
		dst := filepath.FromSlash(args[1])

		msg := `Starts copying the latest files from:`
		fmt.Println(msg, src)
		Sugar.Infow(msg, "src", src, "dst", dst, "log_time", time.Now().Format(itrlog.LogTimeFormat))

		// Starts copying the latest files from.
		if err := kopy.WalkDIRModLatest(src, dst, mDays, IsLogCopiedFile, IgnoreFT, Sugar); err != nil {
			fmt.Println(err)
			Sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return
		}

		// Give some info back to the user's console and the logs as well.
		msg = `Successfully copied the latest files from:`
		fmt.Println(msg, src, " Number of Files Copied: ", NumFilesCopied)
		Sugar.Infow(msg, "src", src, "dst", dst, "copied_files", NumFilesCopied, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	},
}

func init() {
	rootCmd.AddCommand(copymdCmd)
}
