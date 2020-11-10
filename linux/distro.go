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
)

// Many thanks to the people who put together this data set: https://gist.github.com/natefoo/814c5bf936922dad97ff

const moduleName = "github.com/dekobon/distro-detect"

var errorLog = log.New(os.Stderr, "error: ", 0)
var warnLog = log.New(os.Stderr, "warn: ", 0)

var FileSystemRoot = string(os.PathSeparator)
var redhatCompatibleIds = []string{"centos", "fedora", "ol", "rhel", "scientific"}
var rhelCompatibleIds = []string{"centos", "ol", "rhel", "scientific"}

var LogErrorf = func(format string, args ...interface{}) {
	if len(args) > 0 {
		errorLog.Printf(format, args)
	} else {
		warnLog.Println(format)
	}
}

var LogWarnf = func(format string, args ...interface{}) {
	if len(args) > 0 {
		warnLog.Printf(format, args)
	} else {
		warnLog.Println(format)
	}
}

var readBinaryFileFunc = func(filePaths []string) (io.ReadCloser, string, error) {
	for _, filePath := range filePaths {
		if FileSystemRoot != string(os.PathSeparator) {
			filePath = path.Clean(FileSystemRoot + string(os.PathSeparator) + filePath)
		}

		fileInfo, statErr := os.Stat(filePath)
		if statErr != nil || fileInfo.IsDir() {
			continue
		}

		reader, readErr := os.Open(filePath)
		if readErr != nil {
			LogErrorf("unable to open file (%s): %v", filePath, readErr)
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

	defer func() { _ = reader.Close() }()

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		LogErrorf("unable to read file (%s): %v", filePath, err)
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

func BestGuess(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) LinuxDistro {
	LogWarnf("distro is not part of the existing data set - attempting best guess")

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
	reader, pathRead, openErr := readBinaryFileFunc([]string{filePath})
	if openErr != nil {
		if pathRead != "" {
			warnLog.Printf("unable to read release file at the path: %s", pathRead)
		}

		return ReleaseDetails{}, openErr
	}
	defer func() { _ = reader.Close() }()

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
