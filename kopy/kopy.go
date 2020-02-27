// Package kopy stored all the copy operations for gokopy.
package kopy

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"gokopy/itrlog"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// BKMaxLogFileSizeInMB gets the max log file size value in megabytes.
var BKMaxLogFileSizeInMB int = 100 // mb

// BKMaxAgeLogInDays get the max age of a log files in days.
var BKMaxAgeLogInDays int = 0 // 0 days means, it won't delete older backup logs

// NumFilesCopied counts the number of files copied.
var NumFilesCopied int = 0

// NumFoldersCopied counts the number of folders copied.
var NumFoldersCopied int = 0

var cfgFile string
var ignoreFT []string

var logger *zap.Logger
var sugar *zap.SugaredLogger

// ComFileFormat compression file extentions.
const ComFileFormat = ".tar.gz"

// ComSingleFileFormat use zip compression format for any single file need to be compressed.
const ComSingleFileFormat = ".zip"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gokopy" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gokopy")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// LoadInitVars is to load common configurations during init method.
func LoadInitVars() {
	cobra.OnInitialize(InitConfig)
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory

	// Handle errors reading the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create the "config.yaml" asap.
			f, err := os.OpenFile("config.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				log.Fatalf("error opening file: %v", err)
			}
			defer f.Close()

			log.Fatalf("config.yaml file is not found or empty, please copy exactly the example configurations from the documentation")
		} else {
			// Config file was found but another error was produced
			log.Fatalf("fatal error config file: %v", err)
		}
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory

	// Handle errors reading the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; create the "config.yaml" asap.
			f, err := os.OpenFile("config.yaml", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			if err != nil {
				log.Fatalf("error opening file: %v", err)
			}
			defer f.Close()
		} else {
			// Config file was found but another error was produced
			log.Fatalf("fatal error config file: %v", err)
		}
	}

	// Get the default value for the "max_log_file_size_in_mb" setting.
	maxLogFileSize := viper.Get("logging.max_log_file_size_in_mb")
	BKMaxLogFileSizeInMB = maxLogFileSize.(int)
	if _, ok := maxLogFileSize.(int); !ok {
		BKMaxLogFileSizeInMB = 100 // default: mb
	}

	// Get the default value for the "max_age_in_days" setting.
	maxLogAge := viper.Get("logging.max_age_in_days")
	BKMaxAgeLogInDays = maxLogAge.(int)
	if _, ok := maxLogAge.(int); !ok {
		BKMaxAgeLogInDays = 0 // default: days
	}

	// Get the list of ignored file types.
	ignoreFileTypes := viper.Get("ignore.file_type_or_folder_name")
	IGFT := fmt.Sprint(ignoreFileTypes)
	ignoreFT = strings.Split(IGFT, ",")

	// Zap / Lamberjack Logger initialization
	logger = itrlog.InitLog(BKMaxLogFileSizeInMB, BKMaxAgeLogInDays)
	sugar = logger.Sugar()
}

func init() {
	LoadInitVars()
}

// ComFiles compresses one or many files into a single zip archive file.
func ComFiles(dest string, files []string) error {
	newZipFile, err := os.Create(dest)
	if err != nil {
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = AddFileToZip(zipWriter, file); err != nil {
			sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			return err
		}
	}
	return err
}

// AddFileToZip where to
func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}

// FileNameWOExt gets the filename without its file extension.
func FileNameWOExt(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// CompressDIR compressed the entire folder or directory.
func CompressDIR(src string, buf io.Writer, ignoreFT []string) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		_continue := false
		for _, i := range ignoreFT {
			if strings.Index(file, strings.TrimSpace(i)) != -1 {
				_continue = true // Ignore files and folders here
			}
		}

		if _continue == false {
			// if not a dir, write the file content
			header.Name = filepath.ToSlash(file)
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
				defer data.Close()
			}
		}
		return nil
	})

	// produce tar container first
	if err := tw.Close(); err != nil {
		return err
	}

	// finally compress the tar container to gzip.
	if err := zr.Close(); err != nil {
		return err
	}
	return nil
}

// CopyDir copies a whole directory recursively and its sub-directories.
func CopyDir(src, dst string, isLogCopiedFile bool) (int, int, error) {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return int(NumFoldersCopied), int(NumFilesCopied), err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return int(NumFoldersCopied), int(NumFilesCopied), err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return int(NumFoldersCopied), int(NumFilesCopied), err
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		_continue := false
		for _, i := range ignoreFT {
			if strings.Index(srcfp, strings.TrimSpace(i)) != -1 {
				_continue = true // Ignore files and folders here
			}
		}

		if _continue == false {
			if fd.IsDir() {
				if _, _, err = CopyDir(srcfp, dstfp, isLogCopiedFile); err != nil {
					fmt.Println(err)
					sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
				} else {
					NumFoldersCopied++
					// Only log when it's true
					if isLogCopiedFile {
						sugar.Infow("copied_folder", "name", fd.Name(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
						fmt.Println("copied folder: ", fd.Name())
					}
				}
			} else {
				if err = CopyFile(srcfp, dstfp, dst); err != nil {
					fmt.Println(err)
					sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
				} else {
					NumFilesCopied++
					// Only log when it's true
					if isLogCopiedFile {
						sugar.Infow("copied_file", "file", fd.Name(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
						fmt.Println("copied file: ", fd.Name())
					}
				}
			}
		}
	}
	return int(NumFoldersCopied), int(NumFilesCopied), err
}

// CopyFile copy a single file from the source to the destination.
func CopyFile(src, dst, bareDst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	defer srcfd.Close()

	os.MkdirAll(bareDst, os.ModePerm) // Create the dst folder if not exist
	if dstfd, err = os.Create(dst); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// DIRCopyFiles copy a single file from the source to the destination.
func DIRCopyFiles(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		fmt.Println(err)
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// WalkDIRModLatest copies the latest modified files based on the modified date and time.
func WalkDIRModLatest(src, dst string, modDays int, logCopiedFile bool) error {
	os.MkdirAll(dst, os.ModePerm) // Create the root folder first

	//Look for any sub sub-directories and its contents.
	var files []string
	folders := make(map[string]os.FileInfo)
	var startTime int64 = time.Now().AddDate(0, 0, modDays).Unix() // Behind "x" days modified date and time to start the copy operation.
	var endTime int64 = time.Now().Unix()                          // Current date and time
	var err error

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		_continue := false
		for _, i := range ignoreFT {
			if strings.Index(path, strings.TrimSpace(i)) != -1 {
				_continue = true // Loop : ignore files and folders here.
			}
		}

		if _continue == false {
			ft, _ := os.Stat(path)
			if info.IsDir() {
				folders[path] = ft
			} else {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	// Create all the folder structures first before any files will be added later.
	for p, _ := range folders {
		srcFile := filepath.FromSlash(p)
		dstBareDir := filepath.FromSlash(strings.Replace(srcFile, src, dst, -1))
		os.MkdirAll(dstBareDir, os.ModePerm) // Create the dst folder if not exist
	}

	// Now, add all the contents
	for _, f := range files {
		ff, err1 := os.Stat(f)
		if err1 != nil {
			sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			fmt.Println(err1)
		}
		fModTime := ff.ModTime().Unix()

		if fModTime >= startTime && fModTime <= endTime {
			srcFile := filepath.FromSlash(f)
			dstBareDir := filepath.FromSlash(strings.Replace(srcFile, filepath.FromSlash(src), filepath.FromSlash(dst), -1))
			if err = DIRCopyFiles(srcFile, dstBareDir); err != nil {
				fmt.Println(err)
				sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			} else {
				NumFilesCopied++
				// Only log when it's true
				if logCopiedFile {
					fmt.Println("copied file: ", ff.Name())
					sugar.Errorw("copied_file", "name", ff.Name(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
				}
			}
		}
	}
	return err
}

// ExtractTarGz extracts the tar.gz compressed file.
func ExtractTarGz(gzipStream io.Reader, src string, isLogCopiedFile bool) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		fmt.Println("new reader failed")
		sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
	}

	tarReader := tar.NewReader(uncompressedStream)
	fnExtract, fnExtractRoot, fnExtractCounter := "", "", 0
	if b := strings.Contains(src, ComFileFormat); b {
		fnExtract = strings.Replace(src, ComFileFormat, "", -1)
	}

	os.MkdirAll(fnExtract, os.ModePerm) // Create a new dst dir first
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err.Error())
			sugar.Errorw("error", "err", err.Error(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
			log.Fatalf(err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			fnExtractCounter++
			if fnExtractCounter == 1 {
				fnExtractRoot = filepath.FromSlash(header.Name) // Gets the root dir only e.g C:\a
			}

			folderPath := filepath.FromSlash(header.Name) //full dir path
			extractFileTo := ""

			// Replace the original folder root directory of the compressed folder to a new dst location.
			if b := strings.Contains(folderPath, fnExtractRoot); b {
				extractFileTo = strings.Replace(folderPath, fnExtractRoot, fnExtract, -1)
			}
			if err := os.MkdirAll(filepath.FromSlash(extractFileTo), os.ModePerm); err != nil {
				fmt.Println(err.Error())
				sugar.Errorw("error", "err", err.Error(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
				log.Fatalf(err.Error())
			}
		case tar.TypeReg:

			folderPath := filepath.FromSlash(header.Name) //full file path
			extractFileTo := ""

			// Replace the original file path of each files from the compressed folder to a new dst location.
			if b := strings.Contains(folderPath, fnExtractRoot); b {
				extractFileTo = strings.Replace(folderPath, fnExtractRoot, fnExtract, -1)
			}

			outFile, err := os.Create(extractFileTo)
			if err != nil {
				fmt.Println(err.Error())
				sugar.Errorw("error", "err", err.Error(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
				log.Fatalf(err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				fmt.Println(err.Error())
				sugar.Errorw("error", "err", err.Error(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
				log.Fatalf(err.Error())
			}
			// Only log when it's true
			if isLogCopiedFile {
				fmt.Println("extracting to: ", extractFileTo)
				sugar.Infow("extracting to: ", "dst", extractFileTo, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			}
			defer outFile.Close()
		default:
			fmt.Println("unknown type:", filepath.FromSlash(header.Name))
			sugar.Errorw("unknown type", "file_type", filepath.FromSlash(header.Name), "log_time", time.Now().Format(itrlog.LogTimeFormat))
		}
	}
	return err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, isLogCopiedFile bool) error {
	// Read the compressed file's original source folder or directory.
	zipReader, err := zip.OpenReader(src)

	fnExtract := ""
	if b := strings.Contains(src, ComSingleFileFormat); b {
		fnExtract = strings.Replace(src, ComSingleFileFormat, "", -1)
	}

	os.MkdirAll(fnExtract, os.ModePerm) // Create a new dst dir first
	for _, file := range zipReader.Reader.File {
		zippedFile, err := file.Open()
		if err != nil {
			fmt.Println(err)
			sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
		}
		defer zippedFile.Close()

		folderPath := filepath.FromSlash(file.Name) //full file path
		extractFileTo := ""

		// Replace the original folder root directory of the compressed folder to a new dst location.
		if b := strings.Contains(folderPath, filepath.Dir(folderPath)); b {
			extractFileTo = strings.Replace(folderPath, filepath.Dir(folderPath), fnExtract, -1)
		}
		extractedFilePath := filepath.Join(extractFileTo, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(filepath.FromSlash(extractedFilePath), os.ModePerm); err != nil {
				fmt.Println(err.Error())
				sugar.Errorw("error", "err", err.Error(), "log_time", time.Now().Format(itrlog.LogTimeFormat))
				log.Fatalf(err.Error())
			}
		} else {
			folderPath := filepath.FromSlash(file.Name) //full file path
			extractFileTo := ""

			// Replace the original folder root directory of the compressed folder to a new dst location.
			if b := strings.Contains(folderPath, filepath.Dir(folderPath)); b {
				extractFileTo = strings.Replace(folderPath, filepath.Dir(folderPath), fnExtract, -1)
			}
			outputFile, err := os.OpenFile(extractFileTo, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())

			if err != nil {
				fmt.Println(err)
				sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			}
			defer outputFile.Close()

			if _, err = io.Copy(outputFile, zippedFile); err != nil {
				fmt.Println(err)
				sugar.Errorw("error", "err", err, "log_time", time.Now().Format(itrlog.LogTimeFormat))
				return err
			}

			// Only log when it's true
			if isLogCopiedFile {
				fmt.Println("extracting to: ", extractFileTo)
				sugar.Infow("extracting to: ", "dst", extractFileTo, "log_time", time.Now().Format(itrlog.LogTimeFormat))
			}
		}
	}
	return err
}
