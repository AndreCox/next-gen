package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/buger/jsonparser"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/iancoleman/orderedmap"
	"github.com/jpillora/overseer"
	"github.com/otiai10/copy"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var version = "2.2.1"

//go:embed banner.txt
var banner string

func main()  {
	fmt.Println("\033[36m" + banner + "\033[0m")
	prettyPrint("Let's get you set up with the Next Template!", "special")

	// check if there is a new version
	prettyPrint("Checking for updates...", "info")
	update(version)


	if !folderCheck() {
		if existingProjectCheck() {
			prettyPrint("You already have a next-gen project in this folder. We can customize it for you", "info")
			customizeProject()
			prettyPrint("Done!", "success")
			os.Exit(0)
		}
		prettyPrint("This folder does not contain a Next Gen project. Please create a new folder and run this command again.", "error")
		os.Exit(1)
	}
	downloadTemplate()
	unZip()
	copyFiles()
	cleanUp()
	prettyPrint("Great I've fetched the latest version from you, now I just need your help finishing up.", "special")
	customizeProject()

	prettyPrint("Done Customization!", "success")
	prettyPrint("Now we are going to try to see if you have git installed. If you do, we will set up version control for you.", "info")

	if checkGit() {
		prettyPrint("Git is installed. Setting up version control.", "success")
		setupGit()
		prettyPrint("Great we set up version control for you. You can now run git commands to manage your project.", "success")
	} else {
		prettyPrint("Git is not installed. Skipping version control.", "info")
	}

	// check if yarn is installed
	prettyPrint("Now we are going to try to see if you have yarn installed. If you do, we will install dependencies for you.", "info")
	if checkYarn() {
		prettyPrint("Yarn is installed. Installing dependencies.", "success")
		installDependencies()
		prettyPrint("Great we installed dependencies for you. You can now run yarn commands to manage your project.", "success")
		os.Exit(0)
	} else {
	prettyPrint("Yarn is not installed. Skipping dependency installation.", "warning")
	prettyPrint("All done! You can now run 'yarn install' to install the dependencies.", "special")
	os.Exit(0)
	}


}

func setupGit() {
	exec.Command("git", "init").Run()
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "Initial commit").Run()
}

func installDependencies() {
	// run yarn install
	exec.Command("yarn", "install").Run()
}

func checkYarn() bool {
	if !isCommandAvailable("yarn") {
		prettyPrint("Yarn is not installed. We won't be installing dependencies.", "warning")
		return false
	}
		prettyPrint("Yarn is installed. We will install dependencies.", "success")
		return true
}

func checkGit() bool {
	// check if git is installed
	if !isCommandAvailable("git") {
		prettyPrint("Git is not installed. We won't be setting up version control.", "warning")
		return false
	}
		prettyPrint("Git is installed. We will set up version control.", "success")
		return true
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "--version")
	if err := cmd.Run(); err != nil {
			return false
	}
	return true
}

func existingProjectCheck()(bool) {
	// check if there is already a tauri.conf.json
	// if there is, ask if they want to overwrite
	// if not, continue
	files, err := os.ReadDir(".")
	if err != nil {
		prettyPrint("Error reading directory", "error")
	}
	// check if files contains package.json
	for _, file := range files {
		if file.Name() == "package.json" {
			prettyPrint("This folder already contains a package.json file. Checking if this is a next-gen project", "info")

			if checkPackagejson() {
				return true
			} else {
				prettyPrint("This is not a next-gen project. Please create a new folder and run this command again.", "error")
				os.Exit(1)
			}
		}

	}
	return false
}

func checkPackagejson()(bool) {
	// open package.json and check if it has the field next-gen
	file, err := os.Open("package.json")
	if err != nil {
		prettyPrint("Error opening package.json", "error")
		os.Exit(1)
	}

	// save file to string
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		prettyPrint("Error reading package.json", "error")
		os.Exit(1)
	}
	
	id, err := jsonparser.GetString(fileBytes, "next-gen", "id")
	if err != nil {
		prettyPrint("Error reading next-gen ID.", "error")
		os.Exit(1)
	}

	if id == "4326dec8a92b394498ebe4f542833e5a" {
		return true
	}

	return false



}

func folderCheck()(bool) {
	// make sure the folder is empty
	prettyPrint("Checking folder...", "info")
	files, err := os.ReadDir(".")
	if err != nil {
		prettyPrint("Error reading directory", "error")
	}

	if len(files) > 0 {
		prettyPrint("This folder is not empty. Checking if an existing project is here.", "info")
		return false
	}
	prettyPrint("Folder is empty.", "success")
	return true
}

func getInputs()(string, string, string, string, string) {
	var projectName string
	var projectPrettyName string
	var projectDescription string
	var projectAuthor string
	var projectID string

	correct := false
	scanner := bufio.NewScanner(os.Stdin)
	prettyPrint("What would you like to name your project?", "input")
	for !correct {
		if scanner.Scan() {
		    projectName = scanner.Text()   
		}
		// check if string matches regex */
		matched, err := regexp.MatchString(`^(?:@[a-z0-9-*~][a-z0-9-*._~]*/)?[a-z0-9-~][a-z0-9-._~]*$`, projectName)
		if err != nil {
			prettyPrint("Error: " + err.Error(), "error")
			continue
		}
		if matched {
			correct = true
		} else {
			prettyPrint("Sorry, that's not a valid project name. It must match this regex: ", "error")
			prettyPrint("^(?:@[a-z0-9-*~][a-z0-9-*._~]*/)?[a-z0-9-~][a-z0-9-._~]*$", "help")
		}
	}

	// get pretty name by replacing dashes with spaces and capitalizing first letter of each word
	projectPrettyName = strings.ReplaceAll(projectName, "-", " ")
	c := cases.Title(language.Und, cases.NoLower)
	projectPrettyName = c.String(projectPrettyName)

	correct = false
	
	prettyPrint("Take a second to describe your project.", "input")
	for !correct {
		if scanner.Scan() {
		    projectDescription = scanner.Text()
		}
		// check if string matches regex
		matched, err := regexp.MatchString(`^[A-Za-z0-9 ]+$`, projectDescription)
		if err != nil {
			prettyPrint("Error: " + err.Error(), "error")
			continue
		}
		if matched {
			correct = true
		} else {
			prettyPrint("Sorry, that's not a valid project description. It must match this regex: ", "error")
			prettyPrint("^[A-Za-z0-9 ]+$", "help")
		}
	}

	correct = false

	prettyPrint("What is your name?", "input")
	for !correct {
		if scanner.Scan() {
		    projectAuthor = scanner.Text()   
		}
		// check if string matches regex
		matched, err := regexp.MatchString(`^[A-Za-z0-9 ]+$`, projectAuthor)
		if err != nil {
			prettyPrint(err.Error(), "error")
		}
		if matched {
			correct = true
		} else {
			prettyPrint("Sorry, that's not a valid author name. It must match this regex: ", "error")
			prettyPrint(`^[A-Za-z0-9 ]+$`, "help")
		}
	}

	correct = false

	prettyPrint("What is your project ID? (com.company.app)", "input")
	for !correct {
		if scanner.Scan() {
			projectID = scanner.Text()	
		}
		// check if string matches regex
		matched, err := regexp.MatchString(`^[a-z]+(\.[a-z]+)+$`, projectID)		
		if err != nil {
			prettyPrint(err.Error(), "error")
		}
		if matched {
			correct = true
		} else {
			prettyPrint("Sorry, that's not a valid project ID. It must match this regex: ", "error")
			prettyPrint(`^[a-z]+(\.[a-z]+)+$`, "help")
		}
	}	
	
	return projectName, projectPrettyName, projectDescription, projectAuthor, projectID
}

func modifyPackage(projectName string, projectDescription string, projectAuthor string) {
	// open package.json and replace name, description, author
	prettyPrint("Updating package.json", "info")

	file, err := os.Open("package.json")
	if err != nil {
		prettyPrint("Error opening package.json", "error")
		os.Exit(1)
	}

	// save file to string
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		prettyPrint("Error reading package.json", "error")
		os.Exit(1)
	}
	
	jsonData := orderedmap.New()

	if err := json.Unmarshal(fileBytes, &jsonData); err != nil {
		prettyPrint("Error parsing package.json", "error")
		os.Exit(1)
	}
	
	jsonData.Set("name", projectName)
	jsonData.Set("description", projectDescription)
	jsonData.Set("author", projectAuthor)

	defer file.Close()

	// open file again to write
	file, err = os.OpenFile("package.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		prettyPrint("Error opening package.json", "error")
		os.Exit(1)
	}


	// convert json to string
	jsonString, err := json.Marshal(jsonData)
	if err != nil {
		prettyPrint("Error converting json to string", "error")
		os.Exit(1)
	}

	// write to file
	_, err = file.Write(jsonString)
	if err != nil {
		prettyPrint("Error writing to package.json", "error")
		os.Exit(1)
	}

	defer file.Close()
	prettyPrint("Modified package.json", "success")
}

func modifyTauri(projectName string, projectPrettyName string , projectID string, projectDescription string, projectAuthor string) {
	// open /src-tauri/tauri.conf.json and replace id and name
	prettyPrint("Updating tauri.conf.json", "info")

	tauriFile, err := os.Open("src-tauri/tauri.conf.json")
	if err != nil {
		prettyPrint("Error opening tauri.conf.json", "error")
		os.Exit(1)
	}
	
	// save file to string
	tauriBytes, err := io.ReadAll(tauriFile)
	if err != nil {
		prettyPrint("Error reading tauri.conf.json", "error")
		os.Exit(1)
	}

	//jsonData := ordered.New()

	var foo map[string]interface{}
	if err := json.Unmarshal(tauriBytes, &foo); err != nil {
		prettyPrint("Error parsing tauri.conf.json", "error")
		os.Exit(1)
	}

	foo["tauri"].(map[string]interface{})["windows"].([]interface{})[0].(map[string]interface{})["title"] = projectPrettyName
	foo["tauri"].(map[string]interface{})["identifier"] = projectID
	foo["package"].(map[string]interface{})["productName"] = projectName
	foo["tauri"].(map[string]interface{})["bundle"].(map[string]interface{})["identifier"] = projectID

	//convert value to string
	valueString, err := json.Marshal(foo)
	if err != nil {
		prettyPrint("Error converting json to string", "error")
		os.Exit(1)
	}

	// write this back to the file
	defer tauriFile.Close()

	// open file again to write
	tauriFile, err = os.OpenFile("src-tauri/tauri.conf.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		prettyPrint("Error opening tauri.conf.json", "error")
		os.Exit(1)
	}

	// write to file
	_, err = tauriFile.WriteString(string(valueString))
	if err != nil {
		prettyPrint("Error writing to tauri.conf.json", "error")
		fmt.Println(err)
		os.Exit(1)
	}
	tauriFile.Close()
	prettyPrint("Successfully modified tauri.conf.json", "success")

	// open /src-tauri/Cargo.toml and replace name
	prettyPrint("Updating Cargo.toml", "info")

	cargoFile, err := os.Open("src-tauri/Cargo.toml")
	if err != nil {
		prettyPrint("Error opening Cargo.toml", "error")
		os.Exit(1)
	}

	// save file to string
	cargoBytes, err := io.ReadAll(cargoFile)
	if err != nil {
		prettyPrint("Error reading Cargo.toml", "error")
		os.Exit(1)
	}

	cargoMap := map[string]interface{}{}
	_, err = toml.Decode(string(cargoBytes), &cargoMap)
	if err != nil {
		prettyPrint("Error parsing Cargo.toml", "error")
		os.Exit(1)
	}

	cargoMap["package"].(map[string]interface{})["name"] = projectName
	cargoMap["package"].(map[string]interface{})["description"] = projectDescription
	cargoMap["package"].(map[string]interface{})["authors"] = []string{projectAuthor}
	buf := new(bytes.Buffer)
	// convert value to string
	encoder := toml.NewEncoder(buf)
	err = encoder.Encode(cargoMap)
	if err != nil {
		prettyPrint("Error converting json to string", "error")
		os.Exit(1)
	}

	// write this back to the file
	defer cargoFile.Close()

	// open file again to write
	cargoFile, err = os.OpenFile("src-tauri/Cargo.toml", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		prettyPrint("Error opening Cargo.toml", "error")
		os.Exit(1)
	}

	// write to file
	_, err = cargoFile.WriteString(buf.String())
	if err != nil {
		prettyPrint("Error writing to Cargo.toml", "error")
		fmt.Println(err)
		os.Exit(1)
	}
	cargoFile.Close()
	prettyPrint("Successfully modified Cargo.toml", "success")
}

func customizeProject() {
	prettyPrint("Customizing project", "info")

	projectName, projectPrettyName, projectDescription, projectAuthor, projectID := getInputs()						

	modifyPackage(projectName, projectDescription, projectAuthor)

	modifyTauri(projectName, projectPrettyName, projectID, projectDescription, projectAuthor)
  
	prettyPrint("Great we updated your desktop configuration.", "success")
	

	//jsonData.Set("identifier", projectID)
	
	// read in windows from tauri.conf.json using orderedjson
	

	


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

func tryDownload()(*http.Response, bool) {
	resp, err := http.Get("https://github.com/AndreCox/next-template/archive/main.zip")
	if err != nil {
		prettyPrint("HTTP get error: " + err.Error(), "error")
	}
	if resp.StatusCode != 200 {
		prettyPrint("Error: " + resp.Status, "warning")
		os.Exit(1)
	}

	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	prettyPrint("Downloading " + strconv.Itoa(size)  + " bytes.", "📦")

	if size < 1000000 {
		prettyPrint("Size is suspiciously small.", "error")
		return resp, true
	}
	return resp, false
}

func downloadTemplate() {
	prettyPrint("Downloading template...", "info")

	retryCount := 0
	var resp *http.Response
	var retry bool

	for {
		resp, retry = tryDownload()
		if retry {
			prettyPrint("Retrying...", "info")
			retryCount++
			if retryCount > 5 {
				prettyPrint("Failed to download template.", "error")
				os.Exit(1)
			}
			continue
		}
		break
	}

		
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
		defer resp.Body.Close()
		prettyPrint("Template downloaded.", "success")
	
}

func prettyPrint(text string, level string) {

	var Reset  = "\033[0m"
	var Red    = "\033[31m"
	var Green  = "\033[32m"
	//var Yellow = "\033[33m"
	var Blue   = "\033[34m"
	var Purple = "\033[35m"
	var Cyan   = "\033[36m"
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
	case "input":
		levelIcon = "⌨️"
		text = Blue + text + Reset
	case "help":
		levelIcon = "❓"
		text = Cyan + text + Reset
	}
	



	logtime := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(levelIcon + " [" + logtime + "] " + text)
}


func update(version string) {
	latest, found, err := selfupdate.DetectLatest("AndreCox/next-gen")
	if err != nil {
		prettyPrint("An error occurred while detecting version", "error")
		return
	}
	if !found {
		prettyPrint("Latest version could not be found from github repository", "error")
		return
	}

	if latest.LessOrEqual(version) {
		prettyPrint("Current version (" + version + ") is the latest", "info")
		return 
	}

	exe, err := os.Executable()
	if err != nil {
		prettyPrint("An error occurred while detecting executable path", "error")
		return 
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, latest.AssetName, exe); err != nil {
		prettyPrint("An error occurred while updating binary", "error")
		return 
	}
	prettyPrint("Successfully updated to version " + latest.Version() , "success")
	prettyPrint("Restarting...", "info")
	overseer.Restart()
}