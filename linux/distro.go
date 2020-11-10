package linux

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/dekobon/distro-detect/env"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"unicode"
)

// Many thanks to the people who put together this data set: https://gist.github.com/natefoo/814c5bf936922dad97ff

const moduleName = "github.com/dekobon/distro-detect"

var errorLog = log.New(os.Stderr, "error: ", 0)
var warnLog = log.New(os.Stderr, "warn: ", 0)

var FileSystemRoot = string(os.PathSeparator)
var redhatCompatibleIds = []string{"centos", "fedora", "ol", "rhel", "scientific"}
var rhelCompatibleIds = []string{"centos", "ol", "rhel", "scientific"}

var readBinaryFileFunc = func(filePaths []string) (io.ReadCloser, string, error) {
	for _, filePath := range filePaths {
		if FileSystemRoot != string(os.PathSeparator) {
			filePath = path.Clean(FileSystemRoot + string(os.PathSeparator) + filePath)
		}

		fileInfo, statErr := os.Stat(filePath)
		if statErr != nil || fileInfo.IsDir() {
			return nil, filePath, statErr
		}

		reader, readErr := os.Open(filePath)
		if readErr != nil {
			errorLog.Printf("unable to open file (%s): %v", filePath, readErr)
			return nil, filePath, readErr
		}

		return reader, filePath, nil
	}

	errMsg := fmt.Sprintf("unable to create a reader for any of the specified paths: %v", filePaths)
	return nil, "", errors.New(errMsg)
}

var readFileFunc = func(filePaths ...string) (bool, string) {
	reader, filePath, err := readBinaryFileFunc(filePaths)
	if err != nil {
		return false, ""
	}

	defer reader.Close()

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		errorLog.Printf("unable to read file (%s): %v", filePath, err)
		return false, ""
	}

	return true, string(contents)
}

// equalsSplitter is a regex to split apart key value pairs delimited with an equals sign
var equalsSplitter = regexp.MustCompile("^\\s*(\\S+)\\s*=\\s*([\\S ]+)\\s*")

// releaseSplitter is a regex to split apart the contents of /etc/*-release files in the Red Hat Format
var releaseSplitter = regexp.MustCompile("^(.+) (release|version)? (\\S+)\\s*(\\S+)?")

type ReleaseDetails = map[string]string

var DisplayKeys = map[string]string{
	"name":        "Distro Name",
	"id":          "Distro ID",
	"version":     "Distro Version",
	"lsb_release": "Distro LSB",
	"os_release":  "Distro OS",
}

type LinuxDistro struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Version string `json:"version"`
	// LsbRelease contains the contents of /etc/lsb-release.
	LsbRelease ReleaseDetails `json:"lsb_release"`
	// OsRelease contains the contents of /etc/os-release. See: https://www.freedesktop.org/software/systemd/man/os-release.html
	OsRelease ReleaseDetails `json:"os_release"`
}

func (l *LinuxDistro) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"name":        l.Name,
		"id":          l.ID,
		"version":     l.Version,
		"lsb_release": l.LsbRelease,
		"os_release":  l.OsRelease,
	}
}

func (l *LinuxDistro) WriteAllResults(labelFormat string, writer io.Writer) error {
	orderedKeys := []string{"id", "name", "version", "lsb_release", "os_release"}

	for _, key := range orderedKeys {
		err := l.WriteResult(labelFormat, key, writer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *LinuxDistro) WriteResult(labelFormat string, key string, writer io.Writer) error {
	displayKey := DisplayKeys[key]
	value := l.AsMap()[key]

	switch value.(type) {
	case string:
		label := ""
		if labelFormat != "" {
			label = fmt.Sprintf(labelFormat, displayKey)
		}
		_, err := fmt.Fprintf(writer, "%s%s%s", label, value, env.LineBreak)
		if err != nil {
			return err
		}
	case ReleaseDetails:
		ref := reflect.ValueOf(value)
		detailsMap := ref.MapRange()

		for {
			if !detailsMap.Next() {
				break
			}

			k := detailsMap.Key().String()
			v := detailsMap.Value().String()

			label := ""
			if labelFormat != "" {
				label = fmt.Sprintf(labelFormat, displayKey+" "+k)
			}

			_, err := fmt.Fprintf(writer, "%s%s%s", label, v, env.LineBreak)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *LinuxDistro) IsRedhatCompatible() bool {
	for _, id := range redhatCompatibleIds {
		if l.ID == id {
			return true
		}
	}

	if len(l.OsRelease["ID_LIKE"]) > 0 {
		for _, id := range strings.Split(l.OsRelease["ID_LIKE"], " ") {
			if id == "rhel" || id == "fedora" {
				return true
			}
		}
	}

	return false
}

func (l *LinuxDistro) IsRHELCompatible() bool {
	for _, id := range rhelCompatibleIds {
		if l.ID == id {
			return true
		}
	}

	if len(l.OsRelease["ID_LIKE"]) > 0 {
		for _, id := range strings.Split(l.OsRelease["ID_LIKE"], " ") {
			if id == "rhel" {
				return true
			}
		}
	}

	return false
}

func (l *LinuxDistro) UsesRPM() bool {
	if l.IsRedhatCompatible() {
		return true
	}

	if l.ID == "opensuse" || l.ID == "sles" {
		return true
	}

	return false
}

var DistroTests = []func(ReleaseDetails, ReleaseDetails) (bool, LinuxDistro){
	IsCentOS,
	IsRHEL,
	IsUbuntu,
	IsDebian,
	IsAmazonLinux,
	IsFedora,
	IsOpenSuSE,
	IsSLES,
	IsOracleLinux,
	IsPhoton,
	IsAlpine,
	IsArchLinux,
	IsGentoo,
	IsKali,
	IsScientificLinux,
	IsSlackware,
	IsMageia,
	IsClearLinux,
	IsMint,
	IsMXLinux,
	IsNovellOES,
	IsPuppy,
	IsRancherOS,
	IsAlt,
	IsCrux,
	IsSourceMage,
	IsAndroid,
	IsYellowDog,
	IsBusyBox, // BusyBox should come last because it uses process execution
}

func DistroTestFunctionsToFunctionNames(funcs []func(ReleaseDetails, ReleaseDetails) (bool, LinuxDistro)) []string {
	names := make([]string, len(funcs))

	for i, f := range funcs {
		fullName := getFunctionName(f)
		separator := fmt.Sprintf("%s/linux.", moduleName)
		shortName := strings.SplitAfter(fullName, separator)
		names[i] = shortName[1]
	}

	return names
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func DiscoverDistro() LinuxDistro {
	lsbProperties, _ := readReleaseFile("/etc/lsb-release")
	osReleaseProperties, _ := readReleaseFile("/etc/os-release")

	return discoverDistroFromProperties(lsbProperties, osReleaseProperties)
}

func discoverDistroFromProperties(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) LinuxDistro {
	var detectedDistro LinuxDistro
	wasDetected := false

	for _, distroTest := range DistroTests {
		wasDetected, detectedDistro = distroTest(lsbProperties, osReleaseProperties)

		if wasDetected {
			break
		}
	}

	if !wasDetected {
		detectedDistro = BestGuess(lsbProperties, osReleaseProperties)
	}

	return detectedDistro
}

func IsAlpine(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "alpine" {
		return true, LinuxDistro{
			Name:       "Alpine Linux",
			ID:         "alpine",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, content := readFileFunc("/etc/alpine-release")
	if exists {
		version := strings.TrimSpace(content)
		return true, LinuxDistro{
			Name:       "Alpine Linux",
			ID:         "alpine",
			Version:    version,
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	return false, LinuxDistro{}
}

func IsAlt(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "altlinux" {
		return true, LinuxDistro{
			Name:       "ALT Starterkit",
			ID:         "altlinux",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}
	return false, LinuxDistro{}
}

func IsAmazonLinux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] != "amzn" {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "Amazon Linux",
		ID:         "amzn",
		Version:    osReleaseProperties["VERSION_ID"],
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsAndroid(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	exists, contents := readFileFunc("/system/build.prop")
	if exists {
		version := "unknown"

		reader := strings.NewReader(contents)
		releaseInfo, err := parseOSRelease(reader)
		if err == nil {
			if releaseInfo["ro.com.google.gmsversion"] != "" {
				version = releaseInfo["ro.com.google.gmsversion"]
			} else if releaseInfo["ro.build.version.release"] != "" {
				version = releaseInfo["ro.build.version.release"]
			}
		}

		return true, LinuxDistro{
			Name:       "Android",
			ID:         "android",
			Version:    version,
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	return false, LinuxDistro{}
}

func IsArchLinux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] != "arch" {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "Arch Linux",
		ID:         "arch",
		Version:    "rolling",
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsBusyBox(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	// BusyBox isn't really a distro, but rather a collection of applications. We want to rule out the
	// chance that a distro was built using the BusyBox binaries before we indicate that the system is
	// BusyBox.
	exists, _ := readFileFunc("/etc/os-release", "/etc/lsb-release")
	if exists {
		return false, LinuxDistro{}
	}

	searchBytes := "BusyBox v"
	searchBytesSize := len(searchBytes)

	reader, filePath, openErr := readBinaryFileFunc([]string{"/bin/true"})
	if openErr != nil {
		return false, LinuxDistro{}
	}

	defer reader.Close()

	buf := make([]byte, searchBytesSize+5)
	matchedPos := 0
	foundBusyBox := false
	foundVersion := false
	var version string
	position := -1

	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			errorLog.Printf("unable to read in buffer for file(%s): %v", filePath, err)
			return false, LinuxDistro{}
		}

		for i := 0; matchedPos < searchBytesSize && i < n-1; i++ {
			position++

			if foundBusyBox {
				char := rune(buf[i])
				if unicode.IsDigit(char) || char == '.' {
					version += string(char)
				} else if len(version) < 6 {
					foundBusyBox = false
					matchedPos = 0
				} else if len(version) >= 6 {
					foundVersion = true
					break
				}
			} else if buf[i] == searchBytes[matchedPos] {
				if matchedPos+1 == searchBytesSize {
					foundBusyBox = true
				} else {
					matchedPos++
				}
			} else {
				break
			}
		}

		if foundBusyBox && foundVersion {
			break
		}
	}

	if !foundBusyBox {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "BusyBox",
		ID:         "busybox",
		Version:    "v" + version,
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsCentOS(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	// Oracle Linux tries to impersonate Red Hat, so we look to see if the oracle release file is present,
	// if so, we know that this isn't Redhat.
	imOracle, distro := IsOracleLinux(lsbProperties, osReleaseProperties)
	if imOracle {
		return imOracle, distro
	}

	exists, contents := readFileFunc("/etc/centos-release", "/etc/redhat-release")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "CentOS")
		if matched {
			return true, LinuxDistro{
				Name:       "CentOS Linux",
				ID:         "centos",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsClearLinux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "clear-linux-os" {
		return true, LinuxDistro{
			Name:       "Clear Linux OS",
			ID:         "clear-linux-os",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}
	return false, LinuxDistro{}
}

func IsCrux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	exists, contents := readFileFunc("/usr/bin/crux")
	if exists {
		version := "unknown"

		reader := strings.NewReader(contents)
		scanner := bufio.NewScanner(reader)
		rex := regexp.MustCompile("\\s*echo \"CRUX version ([0-9.]+)\"\\s*")
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			matches := rex.FindStringSubmatch(line)

			if len(matches) == 2 {
				version = matches[1]
				break
			}
		}

		return true, LinuxDistro{
			Name:       "CRUX",
			ID:         "crux",
			Version:    version,
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	return false, LinuxDistro{}
}

func IsDebian(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	// MX Linux does a good job of impersonating Debian, we test for it first to rule it out
	iamMx, distro := IsMXLinux(lsbProperties, osReleaseProperties)
	if iamMx {
		return iamMx, distro
	}

	var version string

	debianVersionExists, versionContents := readFileFunc("/etc/debian_version")
	if debianVersionExists {
		version = strings.TrimSpace(versionContents)
	} else {
		return false, LinuxDistro{}
	}

	// Check that this isn't a Debian variant like Ubuntu
	issueExists, issueContents := readFileFunc("/etc/issue")
	if issueExists {
		if !strings.HasPrefix(issueContents, "Debian") {
			return false, LinuxDistro{}
		}
	}

	// After we have checked for the files that would indicate that this is a Debian release,
	// if we don't have a non-blank Debian os release id and, this isn't a Debian distro.
	if osReleaseProperties["ID"] != "debian" && osReleaseProperties["ID"] != "" {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "Debian GNU/Linux",
		ID:         "debian",
		Version:    version,
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsFedora(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "fedora" {
		return true, LinuxDistro{
			Name:       "Fedora",
			ID:         "fedora",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	// Oracle Linux tries to impersonate Red Hat, so we look to see if the oracle release file is present,
	// if so, we know that this isn't Redhat.
	imOracle, distro := IsOracleLinux(lsbProperties, osReleaseProperties)
	if imOracle {
		return imOracle, distro
	}

	exists, contents := readFileFunc("/etc/redhat-release")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "Fedora")
		if matched {
			return true, LinuxDistro{
				Name:       "Fedora",
				ID:         "fedora",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsKali(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "kali" {
		return true, LinuxDistro{
			Name:       "Kali GNU/Linux",
			ID:         "kali",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}
	return false, LinuxDistro{}
}

func IsGentoo(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "gentoo" {
		var version string

		exists, contents := readFileFunc("/etc/gentoo-release")
		if exists {
			match, baseSystemVersion := parseRedhatReleaseContents(contents, "Gentoo")
			if match {
				version = baseSystemVersion
			} else {
				version = "unknown"
			}
		} else {
			version = "unknown"
		}

		return true, LinuxDistro{
			Name:       "Gentoo",
			ID:         "gentoo",
			Version:    version,
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}
	return false, LinuxDistro{}
}

func IsOpenSuSE(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "opensuse" {
		return true, LinuxDistro{
			Name:       "openSUSE",
			ID:         "opensuse",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, contents := readFileFunc("/etc/SuSE-release")
	if exists {
		if strings.HasPrefix(contents, "openSUSE") {
			var version string
			releaseDetails, err := parseOSRelease(strings.NewReader(contents))
			if err == nil {
				version = releaseDetails["VERSION"]
			} else {
				version = "unknown"
			}

			return true, LinuxDistro{
				Name:       "openSUSE",
				ID:         "opensuse",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsOracleLinux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "ol" && osReleaseProperties["VERSION_ID"] != "" {
		return true, LinuxDistro{
			Name:       "Oracle Linux",
			ID:         "ol",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, contents := readFileFunc("/etc/oracle-release")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "Oracle Linux")
		if matched {
			return true, LinuxDistro{
				Name:       "Oracle Linux",
				ID:         "ol",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsPhoton(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "photon" && osReleaseProperties["VERSION_ID"] != "" {
		return true, LinuxDistro{
			Name:       "VMware Photon",
			ID:         "photon",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, contents := readFileFunc("/etc/photon-release")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "VMware Photon Linux")
		if matched {
			return true, LinuxDistro{
				Name:       "VMware Photon",
				ID:         "photon",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsPuppy(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if lsbProperties["DISTRIB_ID"] != "Puppy" {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "Puppy Linux",
		ID:         "puppy",
		Version:    osReleaseProperties["VERSION_ID"],
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsMageia(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "mageia" {
		return true, LinuxDistro{
			Name:       "Mageia",
			ID:         "mageia",
			Version:    osReleaseProperties["VERSION"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}
	return false, LinuxDistro{}
}

func IsMint(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if lsbProperties["DISTRIB_ID"] != "LinuxMint" {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "Linux Mint",
		ID:         "linuxmint",
		Version:    lsbProperties["DISTRIB_RELEASE"],
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsMXLinux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if lsbProperties["DISTRIB_ID"] == "MX" {
		return true, LinuxDistro{
			Name:       "MX Linux",
			ID:         "mx",
			Version:    lsbProperties["DISTRIB_RELEASE"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, content := readFileFunc("/etc/mx-version")
	if exists {
		rex := regexp.MustCompile("(\\S+)-([0-9.]+)")
		match := rex.FindStringSubmatch(content)

		if len(match) == 3 && match[1] == "MX" {
			return true, LinuxDistro{
				Name:       "MX Linux",
				ID:         "mx",
				Version:    match[2],
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsNovellOES(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	exists, contents := readFileFunc("/etc/novell-release")
	if exists {
		if strings.HasPrefix(contents, "Novell Open Enterprise Server") {
			var version string
			releaseDetails, err := parseOSRelease(strings.NewReader(contents))
			if err == nil {
				version = releaseDetails["VERSION"]
			} else {
				version = "unknown"
			}

			return true, LinuxDistro{
				Name:       "Novell Open Enterprise Server",
				ID:         "oes",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsRancherOS(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "rancheros" {
		return true, LinuxDistro{
			Name:       "RancherOS",
			ID:         "rancheros",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	return false, LinuxDistro{}
}

func IsRHEL(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "rhel" && osReleaseProperties["VERSION_ID"] != "" {
		return true, LinuxDistro{
			Name:       "Red Hat Enterprise Linux",
			ID:         "rhel",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	// Oracle Linux tries to impersonate Red Hat, so we look to see if the oracle release file is present,
	// if so, we know that this isn't Redhat.
	imOracle, distro := IsOracleLinux(lsbProperties, osReleaseProperties)
	if imOracle {
		return imOracle, distro
	}

	exists, contents := readFileFunc("/etc/redhat-release", "/etc/redhat-version")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "Red Hat Enterprise Linux")
		if matched {
			return true, LinuxDistro{
				Name:       "Red Hat Enterprise Linux",
				ID:         "rhel",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsSLES(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "sles" {
		return true, LinuxDistro{
			Name:       "SUSE Linux",
			ID:         "sles",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, contents := readFileFunc("/etc/SuSE-release", "/etc/sles-release")
	if exists {
		if strings.HasPrefix(contents, "SUSE Linux") {
			var version string
			releaseDetails, err := parseOSRelease(strings.NewReader(contents))
			if err == nil {
				version = releaseDetails["VERSION"]
			} else {
				version = "unknown"
			}

			return true, LinuxDistro{
				Name:       "SUSE Linux",
				ID:         "sles",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsScientificLinux(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	// Oracle Linux tries to impersonate Red Hat, so we look to see if the oracle release file is present,
	// if so, we know that this isn't Redhat.
	imOracle, distro := IsOracleLinux(lsbProperties, osReleaseProperties)
	if imOracle {
		return imOracle, distro
	}

	exists, contents := readFileFunc("/etc/sl-release", "/etc/redhat-release")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "Scientific Linux")
		if matched {
			return true, LinuxDistro{
				Name:       "Scientific Linux",
				ID:         "scientific",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func IsSlackware(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "slackware" && osReleaseProperties["VERSION_ID"] != "" {
		return true, LinuxDistro{
			Name:       "Slackware",
			ID:         "slackware",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	exists, contents := readFileFunc("/etc/slackware-version")
	if exists {
		if !strings.HasPrefix(contents, "Slackware") {
			return false, LinuxDistro{}
		}

		segments := strings.SplitN(strings.TrimSpace(contents), " ", 2)
		var version string
		if len(segments) == 2 {
			version = segments[1]
		} else {
			version = "unknown"
		}

		return true, LinuxDistro{
			Name:       "Slackware",
			ID:         "slackware",
			Version:    version,
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	return false, LinuxDistro{}
}

func IsSourceMage(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	exists, contents := readFileFunc("/etc/sourcemage-release")
	if exists {
		version := "unknown"

		reader := strings.NewReader(contents)
		scanner := bufio.NewScanner(reader)
		rex := regexp.MustCompile(".*\\((.+)\\).*")
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			matches := rex.FindStringSubmatch(line)

			if len(matches) == 2 {
				version = matches[1]
				break
			}
		}

		return true, LinuxDistro{
			Name:       "Source Mage GNU/Linux",
			ID:         "sourcemage",
			Version:    version,
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
		}
	}

	return false, LinuxDistro{}
}

func IsUbuntu(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if lsbProperties["DISTRIB_ID"] != "Ubuntu" {
		return false, LinuxDistro{}
	}

	return true, LinuxDistro{
		Name:       "Ubuntu",
		ID:         "ubuntu",
		Version:    lsbProperties["DISTRIB_RELEASE"],
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func IsYellowDog(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	exists, contents := readFileFunc("/etc/yellowdog-release")
	if exists {
		matched, version := parseRedhatReleaseContents(contents, "Yellow Dog Linux")
		if matched {
			return true, LinuxDistro{
				Name:       "Yellow Dog Linux",
				ID:         "yellow-dog",
				Version:    version,
				LsbRelease: lsbProperties,
				OsRelease:  osReleaseProperties,
			}
		}
	}

	return false, LinuxDistro{}
}

func BestGuess(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) LinuxDistro {
	warnLog.Printf("distro is not part of the existing data set - attempting best guess")

	var id string
	if osReleaseProperties["ID"] != "" {
		id = osReleaseProperties["ID"]
	} else if lsbProperties["DISTRIB_ID"] != "" {
		id = strings.ToLower(lsbProperties["DISTRIB_ID"])
	} else {
		id = "unknown"
	}

	var name string
	if osReleaseProperties["NAME"] != "" {
		name = osReleaseProperties["NAME"]
	} else if osReleaseProperties["PRETTY_NAME"] != "" {
		segments := strings.SplitN(osReleaseProperties["PRETTY_NAME"], " ", 2)
		name = segments[0]
	} else if lsbProperties["DISTRIB_ID"] != "" {
		name = lsbProperties["DISTRIB_ID"]
	} else if osReleaseProperties["ID"] != "" {
		name = osReleaseProperties["ID"]
	} else {
		name = "Unknown"
	}

	var version string
	if osReleaseProperties["VERSION_ID"] != "" {
		version = osReleaseProperties["VERSION_ID"]
	} else if lsbProperties["DISTRIB_RELEASE"] != "" {
		version = lsbProperties["DISTRIB_RELEASE"]
	} else if osReleaseProperties["VERSION"] != "" {
		segments := strings.SplitN(osReleaseProperties["VERSION"], " ", 2)
		version = segments[0]
	} else {
		version = "unknown"
	}

	return LinuxDistro{
		Name:       name,
		ID:         id,
		Version:    version,
		LsbRelease: lsbProperties,
		OsRelease:  osReleaseProperties,
	}
}

func readReleaseFile(filePath string) (ReleaseDetails, error) {
	reader, filePath, openErr := readBinaryFileFunc([]string{filePath})
	if openErr != nil {
		warnLog.Printf("unable to find release file: %s", filePath)
		return ReleaseDetails{}, openErr
	}
	defer reader.Close()

	properties, parseErr := parseOSRelease(reader)
	return properties, parseErr
}

func parseOSRelease(reader io.Reader) (ReleaseDetails, error) {
	properties := ReleaseDetails{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		key, val, splitErr := splitEqualsKeyVal(line)
		if splitErr != nil {
			continue
		}

		properties[key] = val
	}

	return properties, scanner.Err()
}

func splitEqualsKeyVal(line string) (string, string, error) {
	if line == "" {
		return "", "", errors.New("can't split a blank line")
	}

	if line[0] == '#' {
		return "", "", errors.New(fmt.Sprintf("ignoring commented line: %s", line))
	}

	match := equalsSplitter.FindStringSubmatch(line)
	if len(match) == 0 {
		return "", "", errors.New(fmt.Sprintf("no splittable character for line: %s", line))
	}
	if len(match) != 3 {
		return "", "", errors.New(fmt.Sprintf("unexpected number of matches (%d) for line: %s", len(match), line))
	}

	withoutTrailingWhitespace := strings.TrimSpace(match[2])
	withoutEnclosingQuotes := strings.Trim(withoutTrailingWhitespace, "\"")

	return match[1], withoutEnclosingQuotes, nil
}

func parseRedhatReleaseContents(contents string, expectedDistro string) (bool, string) {
	matches := releaseSplitter.FindStringSubmatch(contents)

	if !strings.HasPrefix(matches[0], expectedDistro) {
		return false, ""
	}

	var version string

	if len(matches) > 3 {
		version = strings.TrimSpace(matches[3])
	} else {
		version = "unknown"
	}

	return true, version
}
