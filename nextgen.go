package main

import (
	"archive/zip"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/otiai10/copy"
)

//go:embed banner.txt
var banner string

func main()  {
	fmt.Println(banner)
	prettyPrint("Let's get you set up with the Next Template.", "special")

	downloadTemplate()
	unZip()
	copyFiles()
	cleanUp()

}

func cleanUp() {
	prettyPrint("Cleaning up...", "info")
	os.Remove("next.zip")
	os.RemoveAll("nextgen-output")
	prettyPrint("Cleaned up.", "success")
}

func copyFiles() {
	prettyPrint("Copying files...", "info")

	// for all files in src directory copy to dst
	err := copy.Copy("nextgen-output/next-template-main/", ".")
	if err != nil {
		prettyPrint("Error copying files: " + err.Error(), "error")
		os.Exit(1)
	}


	prettyPrint("Files copied.", "success")
}

func unZip() {
	prettyPrint("Unzipping template...", "info")
	
	dst := "nextgen-output/"
    archive, err := zip.OpenReader("next.zip")
    if err != nil {
        panic(err)
    }
    defer archive.Close()

    for _, f := range archive.File {
        filePath := filepath.Join(dst, f.Name)
        if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
            prettyPrint("Illegal file path.", "error")
            os.Exit(3)
        }
        if f.FileInfo().IsDir() {
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            prettyPrint("Error creating directory.", "error")
			panic(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
			prettyPrint("Error opening file.", "error")
            panic(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
			prettyPrint("Error opening file in archive.", "error")
            panic(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			prettyPrint("Error copying file.", "error")
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
	prettyPrint("File unzipped.", "success")
}

func downloadTemplate() {
	prettyPrint("Downloading template...", "info")

	resp, err := http.Get("https://github.com/AndreCox/next-template/archive/main.zip")
	if err != nil {
		prettyPrint("HTTP get error: " + err.Error(), "error")
	}
	if resp.StatusCode != 200 {
		prettyPrint("Error: " + resp.Status, "error")
		os.Exit(1)
	}

	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	prettyPrint("Downloading " + strconv.Itoa(size)  + " bytes.", "📦")

	if size < 1000000 {
		prettyPrint("Size is suspiciously small.", "error")
		os.Exit(2)
	}

	defer resp.Body.Close()
    // Create the file
    out, err := os.Create("next.zip")
    if err != nil {
        prettyPrint("Error creating file: " + err.Error(), "error")
		os.Exit(2)
    }

    defer out.Close()

    // Write the body to file
    _, err = io.Copy(out, resp.Body)
    if err != nil {
		prettyPrint("Error downloading template: " + err.Error(), "error")
		os.Exit(5)
	}
	prettyPrint("Template downloaded.", "success")
}

func prettyPrint(text string, level string) {

	var Reset  = "\033[0m"
	var Red    = "\033[31m"
	var Green  = "\033[32m"
	//var Yellow = "\033[33m"
	//var Blue   = "\033[34m"
	var Purple = "\033[35m"
	//var Cyan   = "\033[36m"
	//var Gray   = "\033[37m"
	//var White  = "\033[97m"
	var Orange = "\033[38;5;208m"

	levelIcon := level

	switch level {
	case "info":
		levelIcon = "📝"
	case "success":
		levelIcon = "✅"
		text = Green + text + Reset
	case "error":
		levelIcon = "❌"
		text = Red + text + Reset
	case "warning":
		levelIcon = "⚠️"
		text = Orange + text + Reset
	case "special":
		levelIcon = "✨"
		text = Purple + text + Reset
	}
	



	logtime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(levelIcon + " [" + logtime + "] " + text)
}