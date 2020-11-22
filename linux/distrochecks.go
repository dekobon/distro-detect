package linux

import (
	"bufio"
	"io"
	"regexp"
	"strings"
	"unicode"
)

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

	defer func() { _ = reader.Close() }()

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
			LogErrorf("unable to read in buffer for file(%s): %v", filePath, err)
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

func IsNixOS(lsbProperties ReleaseDetails, osReleaseProperties ReleaseDetails) (bool, LinuxDistro) {
	if osReleaseProperties["ID"] == "nixos" {
		return true, LinuxDistro{
			Name:       "NixOS",
			ID:         "nixos",
			Version:    osReleaseProperties["VERSION_ID"],
			LsbRelease: lsbProperties,
			OsRelease:  osReleaseProperties,
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
