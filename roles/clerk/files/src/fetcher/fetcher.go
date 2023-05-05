package fetcher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type Details struct {
	Manager string
	Name    string
	Version string
}

const get2Cols = " | awk -F' ' '{print $1, $2}'"
const kernelCmd = "uname -r"

// const brewPath = "/opt/homebrew/Cellar"
// const golang = "/opt/golang/"
const pipCmd = "pip freeze | sed -r 's/==/ /g'"
const pip3Cmd = "pip3 freeze | sed -r 's/==/ /g'"
const yumCmd = "yum list installed 2>/dev/null | grep arch"
const aptCmd = "dpkg-query -W"
const rpmCmd = "rpm -qa --queryformat '%{NAME} %{VERSION}\\n'"
const zypperCmd = "zypper search --installed-only -s | awk 'NR>5 {printf $3\" \"$7\"\\n\"}'"

var managers = map[string]string{
	"pip":    pipCmd,
	"pip3":   pip3Cmd,
	"apt":    aptCmd,
	"rpm":    rpmCmd,
	"yum":    yumCmd,
	"zypper": zypperCmd}

func GetPackages(args []string) []Details {
	var allPackages []Details
	for i, manager := range args[1:] {
		if strings.Contains(manager, folderDelimiter) {
			unmanagedPackages(manager, &allPackages, args[1:], i+1)
			continue
		}
		if _, ok := managers[manager]; ok {
			managedPackages(manager, &allPackages)
		}
	}
	allPackages = deduplicate(allPackages)
	return allPackages
}

func fileReader(path string, allPackages *[]Details, managerIndex string) {
	var packObj Details
	parentFolder := getParentFolder(path, managerIndex)
	catFile := fmt.Sprintf("cat %s | head -n 1", path)
	stdout, err := exec.Command("bash", "-c", catFile).Output()
	if err != nil {
		log.Fatal(err)
		return
	}

	packObj.Name = parentFolder
	packObj.Version = removeNewLine(string(stdout), "")
	packObj.Manager = ""
	*allPackages = append(*allPackages, packObj)
}

func folderPackages(path string, allPackages *[]Details, managerIndex string) {
	parentFolder := getParentFolder(path, managerIndex)
	listCmd := fmt.Sprintf("find %s -maxdepth 2 | grep -v '%s$' | grep -E '[0-9\\.\\_\\-]{3,}$'", path, folderDelimiter)
	stdout, err := exec.Command("bash", "-c", listCmd).Output()
	if err != nil {
		log.Fatal(err)
		return
	}

	packagesList := strings.ReplaceAll(string(stdout), path+"/", "")
	packages := extractColumn(packagesList, parentFolder)
	*allPackages = append(*allPackages, packages...)
}

func managedPackages(managerName string, allPackages *[]Details) {
	manager, cmd := managerDetails(managerName)
	if len(manager) > 4 {
		stdout, err := exec.Command("bash", "-c", cmd+get2Cols).Output()
		if err != nil {
			log.Fatal(err)
		}

		packagesList := strings.ReplaceAll(string(stdout), "==", " ")
		packages := extractColumn(packagesList, manager)
		*allPackages = append(*allPackages, packages...)
	}
}

func unmanagedPackages(manager string, allPackages *[]Details, args []string, index int) {
	var managerIndex string
	fileInfo, err := os.Stat(manager)
	if err != nil {
		log.Fatal(err)
	}
	if len(args) > index {
		managerIndex = args[index]
	} else {
		managerIndex = "none"
	}

	if fileInfo.IsDir() {
		folderPackages(manager, &*allPackages, managerIndex)
	} else {
		fileReader(manager, &*allPackages, managerIndex)
	}
}

func getParentFolder(path string, managerIndex string) string {
	cleaned := removeTrailingDelimiter(path, folderDelimiter)
	folders := strings.Split(cleaned, folderDelimiter)
	if value, err := strconv.Atoi(managerIndex); err == nil {
		return folders[value]
	}
	return folders[len(folders)-1]
}

const darwinDetails = "system_profiler SPSoftwareDataType | grep \": \""
const linuxDetails = "cat /etc/os-release | grep ="

func OsDetails(ami string) []Details {
	cmd := linuxDetails
	if runtime.GOOS == "darwin" {
		cmd = darwinDetails
	}

	stdout, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	osKernel := fmt.Sprintf("%s\nKernel=%s\nAMI=%s", string(stdout), kernelDetails(), ami)
	Details := extractColumn(osKernel, "os")
	return Details
}

func kernelDetails() string {
	stdout, err := exec.Command("bash", "-c", kernelCmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	trimmed := strings.ReplaceAll(string(stdout), "\n", "")
	return trimmed
}

const folderDelimiter = "/"

func managerDetails(name string) (string, string) {
	var versionCmd = fmt.Sprintf("%s --version | head -n 1", name)
	stdout, err := exec.Command("bash", "-c", versionCmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	trimmed := removeNewLine(string(stdout), "")
	pathRegex := regexp.MustCompile(`/[^\s]+ `)
	trimmed = pathRegex.ReplaceAllString(trimmed, "")
	return trimmed, managers[name]
}

func initList() (Details, []Details) {
	var packObj Details
	var packages = make([]Details, 0)
	return packObj, packages
}

var pkgRegex = regexp.MustCompile(`^(.*)[/@ ](.*)`)
var osRegex = regexp.MustCompile(`^(.*)[:=](.*)`)

func extractColumn(data string, manager string) []Details {
	regexStr := pkgRegex
	if manager == "os" {
		regexStr = osRegex
	}
	arr := strings.Split(data, "\n")
	packObj, packages := initList()
	for _, pkg := range arr {
		values := regexStr.FindStringSubmatch(pkg)
		if len(values) < 3 {
			continue
		}
		if manager != "os" {
			packObj.Manager = manager
		}
		packObj.Name = strings.TrimSpace(values[1])
		packObj.Version = strings.TrimSpace(values[2])
		fmt.Println((packObj))
		packages = append(packages, packObj)
	}
	return packages
}

func removeNewLine(str string, new string) string {
	newLine := regexp.MustCompile(`:*\n`)
	return newLine.ReplaceAllString(str, new)
}

func removeTrailingDelimiter(str string, delimiter string) string {
	newLine := regexp.MustCompile(delimiter + `$`)
	return newLine.ReplaceAllString(str, "")
}

func deduplicate(sample []Details) []Details {
	var unique []Details
sampleLoop:
	for _, v := range sample {
		for i, u := range unique {
			if v.Manager == u.Manager && v.Name == u.Name && v.Version == u.Version {
				unique[i] = v
				continue sampleLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}
