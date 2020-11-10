package linux

import (
	"fmt"
	"github.com/dekobon/distro-detect/env"
	"io"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Uncomment me to use as a helper for creating new map values from os-release files
// func TestBarfFormatted(t *testing.T) {
// 	data := "DISTRIB_ID=Puppy\nDISTRIB_RELEASE=9\nDISTRIB_CODENAME=FossaPup64\nDISTRIB_DESCRIPTION=\"FossaPup64 9.0\""
// 	reader := strings.NewReader(data)
//
// 	properties, err := parseOSRelease(reader)
// 	if err != nil {
// 		t.Error(err)
// 	}
//
// 	for k, v := range properties {
// 		fmt.Printf("    \"%s\": \"%s\",\n", k, v)
// 	}
// }

func TestMain(m *testing.M) {
	// Randomize distro detection order so that we don't always use the same order
	var seed int64 = time.Now().UnixNano()
	fmt.Printf("DistroTest random seed: %d%s", seed, env.LineBreak)
	rand.Seed(seed)
	rand.Shuffle(len(DistroTests), func(i, j int) { DistroTests[i], DistroTests[j] = DistroTests[j], DistroTests[i] })
	distroTestNames := DistroTestFunctionsToFunctionNames(DistroTests)
	fmt.Printf("DistroTest order: %v%s", strings.Join(distroTestNames, " "), env.LineBreak)

	// Store original read file function and restore it after test run has completed
	origReadFileFunc := readFileFunc
	defer func() {
		readFileFunc = origReadFileFunc
	}()
	// Replace read file function with a function that always returns "not found". If a test needs
	// to use the function, it will be responsible for overriding it.
	readFileFunc = func(...string) (bool, string) {
		return false, ""
	}
	m.Run()

}

func TestParseEmptyOSRelease(t *testing.T) {
	data := ""
	reader := strings.NewReader(data)

	properties, err := parseOSRelease(reader)
	if err != nil {
		t.Error(err)
	}
	if len(properties) > 0 {
		t.Error("properties should be empty")
	}

	for k, v := range properties {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func TestSplitEqualsKeyValSimple(t *testing.T) {
	actual := "a_single_key=a_single_value"
	k, v, err := splitEqualsKeyVal(actual)
	if err != nil {
		t.Error(err)
	}
	if k != "a_single_key" {
		t.Errorf("k has unexpected value: [%s]", k)
	}
	if v != "a_single_value" {
		t.Errorf("v has unexpected value: [%s]", v)
	}
}

func TestSplitEqualsKeyValWithEnclosingQuotes(t *testing.T) {
	actual := "a_single_key=\"a_single_value\""
	k, v, err := splitEqualsKeyVal(actual)
	if err != nil {
		t.Error(err)
	}
	if k != "a_single_key" {
		t.Errorf("k has unexpected value: [%s]", k)
	}
	if v != "a_single_value" {
		t.Errorf("v has unexpected value: [%s]", v)
	}
}

func TestSplitEqualsKeyValWithTrailingLinebreak(t *testing.T) {
	actual := "a_single_key=\"a_single_value\"\n"
	k, v, err := splitEqualsKeyVal(actual)
	if err != nil {
		t.Error(err)
	}
	if k != "a_single_key" {
		t.Errorf("k has unexpected value: [%s]", k)
	}
	if v != "a_single_value" {
		t.Errorf("v has unexpected value: [%s]", v)
	}
}

func TestSplitEqualsKeyValWithWhitespace(t *testing.T) {
	actual := "   a_single_key\t =   a_single_value\t\r\n"
	k, v, err := splitEqualsKeyVal(actual)
	if err != nil {
		t.Error(err)
	}
	if k != "a_single_key" {
		t.Errorf("k has unexpected value: [%s]", k)
	}
	if v != "a_single_value" {
		t.Errorf("v has unexpected value: [%s]", v)
	}
}

func TestParseMissingDelimiterOSRelease(t *testing.T) {
	data := "SOMETHING-NO-SEPARATOR"
	reader := strings.NewReader(data)

	properties, err := parseOSRelease(reader)
	if err != nil {
		t.Error(err)
	}
	if len(properties) > 0 {
		t.Error("properties should be empty")
	}

	for k, v := range properties {
		fmt.Printf("%s: %s\n", k, v)
	}
}

func TestParseUbuntuOSRelease(t *testing.T) {
	data := "DISTRIB_ID=Ubuntu\nDISTRIB_RELEASE=18.04\nDISTRIB_CODENAME=bionic\nDISTRIB_DESCRIPTION=\"Ubuntu 18.04.5 LTS\"\n"
	reader := strings.NewReader(data)

	properties, err := parseOSRelease(reader)
	if err != nil {
		t.Error(err)
	}

	expected := map[string]string{
		"DISTRIB_ID":          "Ubuntu",
		"DISTRIB_RELEASE":     "18.04",
		"DISTRIB_CODENAME":    "bionic",
		"DISTRIB_DESCRIPTION": "Ubuntu 18.04.5 LTS",
	}

	if !reflect.DeepEqual(properties, expected) {
		t.Errorf("unexpected values parsed from os release data:\nExpected:\n%s\nActual:\n%s",
			expected, properties)
	}
}

func TestOracleLinuxOSRelease(t *testing.T) {
	data := "NAME=\"Oracle Linux Server\" \nVERSION=\"6.10\" \nID=\"ol\" \nVERSION_ID=\"6.10\" \nPRETTY_NAME=\"Oracle Linux Server 6.10\"\nANSI_COLOR=\"0;31\" \nCPE_NAME=\"cpe:/o:oracle:linux:6:10:server\"\nHOME_URL=\"https://linux.oracle.com/\" \nBUG_REPORT_URL=\"https://bugzilla.oracle.com/\" \n\nORACLE_BUGZILLA_PRODUCT=\"Oracle Linux 6\" \nORACLE_BUGZILLA_PRODUCT_VERSION=6.10 \nORACLE_SUPPORT_PRODUCT=\"Oracle Linux\" \nORACLE_SUPPORT_PRODUCT_VERSION=6.10\n"
	reader := strings.NewReader(data)

	properties, err := parseOSRelease(reader)
	if err != nil {
		t.Error(err)
	}

	expected := map[string]string{
		"BUG_REPORT_URL":                  "https://bugzilla.oracle.com/",
		"ORACLE_BUGZILLA_PRODUCT_VERSION": "6.10",
		"ORACLE_SUPPORT_PRODUCT":          "Oracle Linux",
		"ORACLE_SUPPORT_PRODUCT_VERSION":  "6.10",
		"NAME":                            "Oracle Linux Server",
		"VERSION":                         "6.10",
		"CPE_NAME":                        "cpe:/o:oracle:linux:6:10:server",
		"HOME_URL":                        "https://linux.oracle.com/",
		"ORACLE_BUGZILLA_PRODUCT":         "Oracle Linux 6",
		"ID":                              "ol",
		"VERSION_ID":                      "6.10",
		"PRETTY_NAME":                     "Oracle Linux Server 6.10",
		"ANSI_COLOR":                      "0;31",
	}

	if !reflect.DeepEqual(properties, expected) {
		t.Errorf("unexpected values parsed from os release data:\nExpected:\n%s\nActual:\n%s",
			expected, properties)
	}
}

func TestParseRedhatReleaseContentsRHEL(t *testing.T) {
	contents := "Red Hat Enterprise Linux Server release 7.6 (Maipo)\n"
	expected := "7.6"
	matched, actual := parseRedhatReleaseContents(contents, "Red Hat")
	if !matched {
		t.Error("distro name didn't match")
	}
	if expected != actual {
		t.Errorf("parsed version (%s) didn't match expectation (%s)", actual, expected)
	}
}

func TestParseRedhatReleaseContentsOldGentoo(t *testing.T) {
	contents := "Gentoo Base System version 1.6.14\n"
	expected := "1.6.14"
	matched, actual := parseRedhatReleaseContents(contents, "Gentoo")
	if !matched {
		t.Error("distro name didn't match")
	}
	if expected != actual {
		t.Errorf("parsed version (%s) didn't match expectation (%s)", actual, expected)
	}
}

func TestDiscoverAlpineOld(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/alpine-release"}) {
			return true, "3.12.1"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "alpine", "Alpine Linux", "3.12.1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverAlpine3(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/alpine-release"}) {
			return true, "3.12.1"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"VERSION_ID":     "3.12.1",
		"PRETTY_NAME":    "Alpine Linux v3.12",
		"HOME_URL":       "https://alpinelinux.org/",
		"BUG_REPORT_URL": "https://bugs.alpinelinux.org/",
		"NAME":           "Alpine Linux",
		"ID":             "alpine",
	}

	distroIsDetectedBasedOnProperties(t, "alpine", "Alpine Linux", "3.12.1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverAlt(t *testing.T) {
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"ANSI_COLOR":     "1;33",
		"CPE_NAME":       "cpe:/o:alt:starterkit:p9",
		"HOME_URL":       "http://en.altlinux.org/starterkits",
		"NAME":           "starter kit",
		"VERSION":        "p9 (Hypericum)",
		"ID":             "altlinux",
		"VERSION_ID":     "p9",
		"PRETTY_NAME":    "ALT Starterkit (Hypericum)",
		"BUG_REPORT_URL": "https://bugs.altlinux.org/",
	}

	distroIsDetectedBasedOnProperties(t, "altlinux", "ALT Starterkit", "p9", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverAmazonLinux(t *testing.T) {
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"NAME":        "Amazon Linux",
		"VERSION":     "2",
		"VERSION_ID":  "2",
		"PRETTY_NAME": "Amazon Linux 2",
		"ANSI_COLOR":  "0;33",
		"CPE_NAME":    "cpe:2.3:o:amazon:amazon_linux:2",
		"HOME_URL":    "https://amazonlinux.com/",
		"ID":          "amzn",
		"ID_LIKE":     "centos rhel fedora",
	}

	distroIsDetectedBasedOnProperties(t, "amzn", "Amazon Linux", "2", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverAndroid(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/system/build.prop"}) {
			return true, "\n# begin build properties\n# autogenerated by buildinfo.sh\nro.build.id=PI\nro.build.display.id=android_x86_64-userdebug 9 PI eng.lh.20200325.112926 test-keys\nro.build.version.incremental=eng.lh.20200325.112926\nro.build.version.sdk=28\nro.build.version.preview_sdk=0\nro.build.version.codename=REL\nro.build.version.all_codenames=REL\nro.build.version.release=9\nro.build.version.security_patch=2018-08-05\nro.build.version.base_os=\nro.build.version.min_supported_target_sdk=17\nro.build.date=Wed Mar 25 11:28:56 CST 2020\nro.build.date.utc=1585106936\nro.build.type=userdebug\nro.build.user=lh\nro.build.host=server2\nro.build.tags=test-keys\nro.build.flavor=android_x86_64-userdebug\nro.product.brand=Android-x86\nro.product.name=android_x86_64\nro.product.device=x86_64\n# ro.product.cpu.abi and ro.product.cpu.abi2 are obsolete,\n# use ro.product.cpu.abilist instead.\nro.product.cpu.abi=x86_64\nro.product.cpu.abilist=x86_64,x86,armeabi-v7a,armeabi\nro.product.cpu.abilist32=x86,armeabi-v7a,armeabi\nro.product.cpu.abilist64=x86_64\nro.product.locale=en-US\nro.wifi.channels=\n# ro.build.product is obsolete; use ro.product.device\nro.build.product=x86_64\n# Do not try to parse description, fingerprint, or thumbprint\nro.build.description=android_x86_64-userdebug 9 PI eng.lh.20200325.112926 test-keys\nro.build.fingerprint=Android-x86/android_x86_64/x86_64:9/PI/lh03251128:userdebug/test-keys\nro.build.characteristics=tablet\n# end build properties\n\n#\n# ADDITIONAL_BUILD_PROPERTIES\n#\nro.com.android.dateformat=MM-dd-yyyy\nro.ril.hsxpa=1\nro.ril.gprsclass=10\nkeyguard.no_require_sim=true\nro.com.android.dataroaming=true\nmedia.sf.hwaccel=1\nmedia.sf.omx-plugin=libffmpeg_omx.so\nmedia.sf.extractor-plugin=libffmpeg_extractor.so\nro.opengles.version=196608\nro.hardware.vulkan.level=1\nro.hardware.vulkan.version=4194307\ndalvik.vm.heapstartsize=16m\ndalvik.vm.heapgrowthlimit=192m\ndalvik.vm.heapsize=512m\ndalvik.vm.heaptargetutilization=0.75\ndalvik.vm.heapminfree=512k\ndalvik.vm.heapmaxfree=8m\nro.com.google.gmsversion=9.0_r1\nro.com.google.clientidbase=android-asus\nro.com.google.clientidbase.ms=android-asus\nro.com.google.clientidbase.am=android-asus\nro.com.google.clientidbase.gmm=android-asus\nro.com.google.clientidbase.yt=android-asus\nro.setupwizard.mode=ENABLED\nro.dalvik.vm.isa.arm=x86\nro.enable.native.bridge.exec=1\nro.dalvik.vm.isa.arm64=x86_64\nro.enable.native.bridge.exec64=1\nro.carrier=unknown\nro.config.notification_sound=OnTheHunt.ogg\nro.config.alarm_alert=Alarm_Classic.ogg\nro.dalvik.vm.native.bridge=0\nro.bionic.ld.warning=1\nro.art.hiddenapi.warning=1\nro.treble.enabled=false\npersist.sys.dalvik.vm.lib.2=libart.so\ndalvik.vm.isa.x86_64.variant=x86_64\ndalvik.vm.isa.x86_64.features=default\ndalvik.vm.isa.x86.variant=x86_64\ndalvik.vm.isa.x86.features=default\ndalvik.vm.lockprof.threshold=500\nnet.bt.name=Android\ndalvik.vm.stack-trace-dir=/data/anr\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "android", "Android", "9.0_r1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverArchLinux(t *testing.T) {
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"DOCUMENTATION_URL": "https://wiki.archlinux.org/",
		"BUG_REPORT_URL":    "https://bugs.archlinux.org/",
		"NAME":              "Arch Linux",
		"PRETTY_NAME":       "Arch Linux",
		"ID":                "arch",
		"BUILD_ID":          "rolling",
		"ANSI_COLOR":        "38;2;23;147;209",
		"HOME_URL":          "https://www.archlinux.org/",
		"SUPPORT_URL":       "https://bbs.archlinux.org/",
		"LOGO":              "archlinux",
	}

	distroIsDetectedBasedOnProperties(t, "arch", "Arch Linux", "rolling", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverBusyBox(t *testing.T) {
	originalReadBinaryFileFunc := readBinaryFileFunc
	readBinaryFileFunc = func(filePaths []string) (io.ReadCloser, string, error) {
		if reflect.DeepEqual(filePaths, []string{"/bin/true"}) {
			reader, err := os.Open("test-binary-busybox-amd64-true")
			return reader, "/bin/true", err
		} else {
			return nil, "", nil
		}
	}
	t.Cleanup(func() {
		readBinaryFileFunc = originalReadBinaryFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "busybox", "BusyBox", "v1.32.0", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverCentOS5(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/centos-release", "/etc/redhat-release"}) {
			return true, "CentOS release 5.11 (Final)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "centos", "CentOS Linux", "5.11", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverCentOS6(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/centos-release", "/etc/redhat-release"}) {
			return true, "CentOS release 6.10 (Final)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "centos", "CentOS Linux", "6.10", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverCentOS7(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/centos-release", "/etc/redhat-release"}) {
			return true, "CentOS Linux release 7.8.2003 (Core)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"NAME":                            "CentOS Linux",
		"ID":                              "centos",
		"ID_LIKE":                         "rhel fedora",
		"VERSION_ID":                      "7",
		"REDHAT_SUPPORT_PRODUCT":          "centos",
		"BUG_REPORT_URL":                  "https://bugs.centos.org/",
		"VERSION":                         "7 (Core)",
		"PRETTY_NAME":                     "CentOS Linux 7 (Core)",
		"CPE_NAME":                        "cpe:/o:centos:centos:7",
		"CENTOS_MANTISBT_PROJECT":         "CentOS-7",
		"REDHAT_SUPPORT_PRODUCT_VERSION":  "7",
		"ANSI_COLOR":                      "0;31",
		"HOME_URL":                        "https://www.centos.org/",
		"CENTOS_MANTISBT_PROJECT_VERSION": "7",
	}

	distroIsDetectedBasedOnProperties(t, "centos", "CentOS Linux", "7.8.2003", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverCentOS8(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/centos-release", "/etc/redhat-release"}) {
			return true, "CentOS Linux release 8.2.2004 (Core)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"VERSION_ID":                      "8",
		"CPE_NAME":                        "cpe:/o:centos:centos:8",
		"REDHAT_SUPPORT_PRODUCT":          "centos",
		"VERSION":                         "8 (Core)",
		"ID_LIKE":                         "rhel fedora",
		"CENTOS_MANTISBT_PROJECT":         "CentOS-8",
		"ID":                              "centos",
		"BUG_REPORT_URL":                  "https://bugs.centos.org/",
		"PLATFORM_ID":                     "platform:el8",
		"ANSI_COLOR":                      "0;31",
		"HOME_URL":                        "https://www.centos.org/",
		"CENTOS_MANTISBT_PROJECT_VERSION": "8",
		"REDHAT_SUPPORT_PRODUCT_VERSION":  "8",
		"NAME":                            "CentOS Linux",
		"PRETTY_NAME":                     "CentOS Linux 8 (Core)",
	}

	distroIsDetectedBasedOnProperties(t, "centos", "CentOS Linux", "8.2.2004", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverClearLinux(t *testing.T) {
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"HOME_URL":           "https://clearlinux.org",
		"BUG_REPORT_URL":     "mailto:dev@lists.clearlinux.org",
		"NAME":               "Clear Linux OS",
		"VERSION":            "1",
		"ID_LIKE":            "clear-linux-os",
		"VERSION_ID":         "33910",
		"PRIVACY_POLICY_URL": "http://www.intel.com/privacy",
		"BUILD_ID":           "33910",
		"ID":                 "clear-linux-os",
		"PRETTY_NAME":        "Clear Linux OS",
		"ANSI_COLOR":         "1;35",
		"SUPPORT_URL":        "https://clearlinux.org",
	}

	distroIsDetectedBasedOnProperties(t, "clear-linux-os", "Clear Linux OS", "33910", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverCrux3(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/usr/bin/crux"}) {
			return true, "#!/bin/sh\n\necho \"CRUX version 3.0\"\n\n# End of file\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "crux", "CRUX", "3.0", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverDebian6(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}

		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "6.0.10\n"
		} else if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 6.0 \\n \\l\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "debian", "Debian GNU/Linux", "6.0.10", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverDebian7(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}

		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "7.11\n"
		} else if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 7 \\n \\l\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"VERSION":        "7 (wheezy)",
		"ID":             "debian",
		"ANSI_COLOR":     "1;31",
		"VERSION_ID":     "7",
		"NAME":           "Debian GNU/Linux",
		"HOME_URL":       "http://www.debian.org/",
		"SUPPORT_URL":    "http://www.debian.org/support/",
		"BUG_REPORT_URL": "http://bugs.debian.org/",
		"PRETTY_NAME":    "Debian GNU/Linux 7 (wheezy)",
	}

	distroIsDetectedBasedOnProperties(t, "debian", "Debian GNU/Linux", "7.11", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverDebian8(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}

		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "8.11\n"
		} else if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 8 \\n \\l\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"VERSION_ID":     "8",
		"VERSION":        "8 (jessie)",
		"ID":             "debian",
		"HOME_URL":       "http://www.debian.org/",
		"SUPPORT_URL":    "http://www.debian.org/support",
		"BUG_REPORT_URL": "https://bugs.debian.org/",
		"PRETTY_NAME":    "Debian GNU/Linux 8 (jessie)",
		"NAME":           "Debian GNU/Linux",
	}

	distroIsDetectedBasedOnProperties(t, "debian", "Debian GNU/Linux", "8.11", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverDebian9(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}

		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "9.13\n"
		} else if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 9 \\n \\l\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"ID":               "debian",
		"SUPPORT_URL":      "https://www.debian.org/support",
		"NAME":             "Debian GNU/Linux",
		"VERSION":          "9 (stretch)",
		"VERSION_CODENAME": "stretch",
		"HOME_URL":         "https://www.debian.org/",
		"BUG_REPORT_URL":   "https://bugs.debian.org/",
		"PRETTY_NAME":      "Debian GNU/Linux 9 (stretch)",
		"VERSION_ID":       "9",
	}

	distroIsDetectedBasedOnProperties(t, "debian", "Debian GNU/Linux", "9.13", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverDebian10(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}

		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "10.6\n"
		} else if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 10 \\n \\l\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"VERSION_ID":       "10",
		"VERSION":          "10 (buster)",
		"VERSION_CODENAME": "buster",
		"SUPPORT_URL":      "https://www.debian.org/support",
		"BUG_REPORT_URL":   "https://bugs.debian.org/",
		"PRETTY_NAME":      "Debian GNU/Linux 10 (buster)",
		"NAME":             "Debian GNU/Linux",
		"ID":               "debian",
		"HOME_URL":         "https://www.debian.org/",
	}

	distroIsDetectedBasedOnProperties(t, "debian", "Debian GNU/Linux", "10.6", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverFedora20(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/redhat-release"}) {
			return true, "Fedora release 20 (Heisenbug)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"BUG_REPORT_URL":                  "https://bugzilla.redhat.com/",
		"REDHAT_BUGZILLA_PRODUCT_VERSION": "20",
		"VERSION_ID":                      "20",
		"CPE_NAME":                        "cpe:/o:fedoraproject:fedora:20",
		"ID":                              "fedora",
		"PRETTY_NAME":                     "Fedora 20 (Heisenbug)",
		"ANSI_COLOR":                      "0;34",
		"HOME_URL":                        "https://fedoraproject.org/",
		"REDHAT_BUGZILLA_PRODUCT":         "Fedora",
		"REDHAT_SUPPORT_PRODUCT":          "Fedora",
		"NAME":                            "Fedora",
		"VERSION":                         "20 (Heisenbug)",
		"REDHAT_SUPPORT_PRODUCT_VERSION":  "20",
	}

	distroIsDetectedBasedOnProperties(t, "fedora", "Fedora", "20", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverGentoo1(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/gentoo-release"}) {
			return true, "Gentoo Base System version 1.6.14\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"HOME_URL":       "https://www.gentoo.org/",
		"SUPPORT_URL":    "https://www.gentoo.org/support/",
		"BUG_REPORT_URL": "https://bugs.gentoo.org/",
		"NAME":           "Gentoo",
		"ID":             "gentoo",
		"PRETTY_NAME":    "Gentoo/Linux",
		"ANSI_COLOR":     "1;32",
	}

	distroIsDetectedBasedOnProperties(t, "gentoo", "Gentoo", "1.6.14", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverGentoo2(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/gentoo-release"}) {
			return true, "Gentoo Base System release 2.6\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"HOME_URL":       "https://www.gentoo.org/",
		"SUPPORT_URL":    "https://www.gentoo.org/support/",
		"BUG_REPORT_URL": "https://bugs.gentoo.org/",
		"NAME":           "Gentoo",
		"ID":             "gentoo",
		"PRETTY_NAME":    "Gentoo/Linux",
		"ANSI_COLOR":     "1;32",
	}

	distroIsDetectedBasedOnProperties(t, "gentoo", "Gentoo", "2.6", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverKali(t *testing.T) {
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"BUG_REPORT_URL":   "https://bugs.kali.org/",
		"PRETTY_NAME":      "Kali GNU/Linux Rolling",
		"ID":               "kali",
		"VERSION":          "2020.3",
		"VERSION_CODENAME": "kali-rolling",
		"ID_LIKE":          "debian",
		"ANSI_COLOR":       "1;31",
		"NAME":             "Kali GNU/Linux",
		"VERSION_ID":       "2020.3",
		"HOME_URL":         "https://www.kali.org/",
		"SUPPORT_URL":      "https://forums.kali.org/",
	}

	distroIsDetectedBasedOnProperties(t, "kali", "Kali GNU/Linux", "2020.3", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverRHEL6(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/redhat-release", "/etc/redhat-version"}) {
			return true, "Red Hat Enterprise Linux Server release 6.5 (Santiago)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "rhel", "Red Hat Enterprise Linux", "6.5", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverRHEL7(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/redhat-release", "/etc/redhat-version"}) {
			return true, "Red Hat Enterprise Linux Server release 7.6 (Maipo)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"CPE_NAME":                        "cpe:/o:redhat:enterprise_linux:7.6:GA:server",
		"REDHAT_SUPPORT_PRODUCT":          "Red Hat Enterprise Linux",
		"REDHAT_SUPPORT_PRODUCT_VERSION":  "7.6",
		"ANSI_COLOR":                      "0;31",
		"VARIANT":                         "Server",
		"ID":                              "rhel",
		"VERSION":                         "7.6 (Maipo)",
		"ID_LIKE":                         "fedora",
		"HOME_URL":                        "https://www.redhat.com/",
		"REDHAT_BUGZILLA_PRODUCT":         "Red Hat Enterprise Linux 7",
		"REDHAT_BUGZILLA_PRODUCT_VERSION": "7.6",
		"NAME":                            "Red Hat Enterprise Linux Server",
		"VERSION_ID":                      "7.6",
		"PRETTY_NAME":                     "Red Hat Enterprise Linux",
		"BUG_REPORT_URL":                  "https://bugzilla.redhat.com/",
		"VARIANT_ID":                      "server",
	}

	distroIsDetectedBasedOnProperties(t, "rhel", "Red Hat Enterprise Linux", "7.6", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverMageia(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "Mageia",
		"DISTRIB_RELEASE":     "3",
		"DISTRIB_CODENAME":    "thornicroft",
		"DISTRIB_DESCRIPTION": "Mageia 3",
	}
	osReleaseProperties := map[string]string{
		"ID":             "mageia",
		"PRETTY_NAME":    "Mageia 3",
		"HOME_URL":       "http://www.mageia.org/",
		"SUPPORT_URL":    "http://www.mageia.org/support/",
		"BUG_REPORT_URL": "https://bugs.mageia.org/",
		"NAME":           "Mageia",
		"VERSION":        "3",
		"ID_LIKE":        "mandriva fedora",
		"ANSI_COLOR":     "1;36",
	}

	distroIsDetectedBasedOnProperties(t, "mageia", "Mageia", "3", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverMint(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "LinuxMint",
		"DISTRIB_RELEASE":     "20",
		"DISTRIB_CODENAME":    "ulyana",
		"DISTRIB_DESCRIPTION": "Linux Mint 20 Ulyana",
	}
	osReleaseProperties := map[string]string{
		"HOME_URL":           "https://www.linuxmint.com/",
		"SUPPORT_URL":        "https://forums.ubuntu.com/",
		"BUG_REPORT_URL":     "http://linuxmint-troubleshooting-guide.readthedocs.io/en/latest/",
		"PRIVACY_POLICY_URL": "https://www.linuxmint.com/",
		"NAME":               "Linux Mint",
		"VERSION_ID":         "20",
		"ID_LIKE":            "ubuntu",
		"PRETTY_NAME":        "Linux Mint 20",
		"VERSION_CODENAME":   "ulyana",
		"UBUNTU_CODENAME":    "focal",
		"VERSION":            "20 (Ulyana)",
		"ID":                 "linuxmint",
	}

	distroIsDetectedBasedOnProperties(t, "linuxmint", "Linux Mint", "20", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverMXLinuxOld(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}
		mxVersionPaths := []string{"/etc/mx-version"}

		if reflect.DeepEqual(filePaths, mxVersionPaths) {
			return true, "MX-19.2_ahs_x64 patito feo May 31, 2020\n"
		}
		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "10.6\n"
		}
		if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 10 \\n \\l\n"
		}

		return false, ""
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "mx", "MX Linux", "19.2", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverMXLinux(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		debianVersionPaths := []string{"/etc/debian_version"}
		issuePaths := []string{"/etc/issue"}
		mxVersionPaths := []string{"/etc/mx-version"}

		if reflect.DeepEqual(filePaths, mxVersionPaths) {
			return true, "MX-19.2_ahs_x64 patito feo May 31, 2020\n"
		}
		if reflect.DeepEqual(filePaths, debianVersionPaths) {
			return true, "10.6\n"
		}
		if reflect.DeepEqual(filePaths, issuePaths) {
			return true, "Debian GNU/Linux 10 \\n \\l\n"
		}

		return false, ""
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{
		"PRETTY_NAME":         "MX 19.2 patito feo",
		"DISTRIB_ID":          "MX",
		"DISTRIB_RELEASE":     "19.2",
		"DISTRIB_CODENAME":    "patito feo",
		"DISTRIB_DESCRIPTION": "MX 19.2 patito feo",
	}
	osReleaseProperties := map[string]string{
		"VERSION_ID":       "10",
		"VERSION":          "10 (buster)",
		"VERSION_CODENAME": "buster",
		"SUPPORT_URL":      "https://www.debian.org/support",
		"BUG_REPORT_URL":   "https://bugs.debian.org/",
		"PRETTY_NAME":      "Debian GNU/Linux 10 (buster)",
		"NAME":             "Debian GNU/Linux",
		"ID":               "debian",
		"HOME_URL":         "https://www.debian.org/",
	}

	distroIsDetectedBasedOnProperties(t, "mx", "MX Linux", "19.2", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverNovellOES(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/novell-release"}) {
			return true, "Novell Open Enterprise Server 2.0.1 (i586)\nVERSION = 2.0.1\nPATCHLEVEL = 1\nBUILD\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "oes", "Novell Open Enterprise Server", "2.0.1", lsbProperties,
		osReleaseProperties)
}

// TestOpenSuSEOld tests versions of Open SuSE that don't have a /etc/os-release file.
func TestDiscoverOpenSuSEOld(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/SuSE-release"}) {
			return true, "openSUSE 42.1 (x86_64)\nVERSION = 42.1\nCODENAME = Malachite\n# /etc/SuSE-release is deprecated and will be removed in the future, use /etc/os-release instead\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "opensuse", "openSUSE", "42.1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverOpenSuSE42(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/SuSE-release"}) {
			return true, "openSUSE 42.1 (x86_64)\nVERSION = 42.1\nCODENAME = Malachite\n# /etc/SuSE-release is deprecated and will be removed in the future, use /etc/os-release instead\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"ID_LIKE":        "suse",
		"NAME":           "openSUSE Leap",
		"VERSION":        "42.1",
		"VERSION_ID":     "42.1",
		"ANSI_COLOR":     "0;32",
		"HOME_URL":       "https://opensuse.org/",
		"PRETTY_NAME":    "openSUSE Leap 42.1 (x86_64)",
		"ID":             "opensuse",
		"CPE_NAME":       "cpe:/o:opensuse:opensuse:42.1",
		"BUG_REPORT_URL": "https://bugs.opensuse.org",
	}

	distroIsDetectedBasedOnProperties(t, "opensuse", "openSUSE", "42.1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverOracleLinux6(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/redhat-release"}) {
			// Yes - of course, Oracle Linux impersonates Red Hat if you try to read /etc/redhat-release
			return true, "Red Hat Enterprise Linux Server release 6.10 (Santiago)\n"
		}
		if reflect.DeepEqual(filePaths, []string{"/etc/oracle-release"}) {
			// Yes - of course, Oracle Linux impersonates Red Hat if you try to read /etc/redhat-release
			return true, "Oracle Linux Server release 6.10\n"
		}

		return false, ""
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"BUG_REPORT_URL":                  "https://bugzilla.oracle.com/",
		"ORACLE_BUGZILLA_PRODUCT_VERSION": "6.10",
		"ORACLE_SUPPORT_PRODUCT":          "Oracle Linux",
		"ORACLE_SUPPORT_PRODUCT_VERSION":  "6.10",
		"NAME":                            "Oracle Linux Server",
		"VERSION":                         "6.10",
		"CPE_NAME":                        "cpe:/o:oracle:linux:6:10:server",
		"HOME_URL":                        "https://linux.oracle.com/",
		"ORACLE_BUGZILLA_PRODUCT":         "Oracle Linux 6",
		"ID":                              "ol",
		"VERSION_ID":                      "6.10",
		"PRETTY_NAME":                     "Oracle Linux Server 6.10",
		"ANSI_COLOR":                      "0;31",
	}

	distroIsDetectedBasedOnProperties(t, "ol", "Oracle Linux", "6.10", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverOracleLinux7(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/redhat-release"}) {
			// Yes - of course, Oracle Linux impersonates Red Hat if you try to read /etc/redhat-release
			return true, "Red Hat Enterprise Linux Server release 7.9 (Maipo)\n"
		}
		if reflect.DeepEqual(filePaths, []string{"/etc/oracle-release"}) {
			// Yes - of course, Oracle Linux impersonates Red Hat if you try to read /etc/redhat-release
			return true, "Oracle Linux Server release 7.9\n"
		}

		return false, ""
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"NAME":                            "Oracle Linux Server",
		"VARIANT":                         "Server",
		"HOME_URL":                        "https://linux.oracle.com/",
		"ORACLE_SUPPORT_PRODUCT":          "Oracle Linux",
		"ORACLE_SUPPORT_PRODUCT_VERSION":  "7.9",
		"VARIANT_ID":                      "server",
		"PRETTY_NAME":                     "Oracle Linux Server 7.9",
		"ID":                              "ol",
		"ID_LIKE":                         "fedora",
		"VERSION_ID":                      "7.9",
		"ANSI_COLOR":                      "0;31",
		"CPE_NAME":                        "cpe:/o:oracle:linux:7:9:server",
		"ORACLE_BUGZILLA_PRODUCT_VERSION": "7.9",
		"VERSION":                         "7.9",
		"BUG_REPORT_URL":                  "https://bugzilla.oracle.com/",
		"ORACLE_BUGZILLA_PRODUCT":         "Oracle Linux 7",
	}

	distroIsDetectedBasedOnProperties(t, "ol", "Oracle Linux", "7.9", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverOracleLinux8(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/redhat-release"}) {
			// Yes - of course, Oracle Linux impersonates Red Hat if you try to read /etc/redhat-release
			return true, "Oracle Linux Server release 7.9\n"
		}
		if reflect.DeepEqual(filePaths, []string{"/etc/oracle-release"}) {
			return true, "Oracle Linux Server release 8.2\n"
		}

		return false, ""
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"ANSI_COLOR":                      "0;31",
		"CPE_NAME":                        "cpe:/o:oracle:linux:8:2:server",
		"BUG_REPORT_URL":                  "https://bugzilla.oracle.com/",
		"PLATFORM_ID":                     "platform:el8",
		"PRETTY_NAME":                     "Oracle Linux Server 8.2",
		"ORACLE_BUGZILLA_PRODUCT":         "Oracle Linux 8",
		"ORACLE_SUPPORT_PRODUCT":          "Oracle Linux",
		"ID":                              "ol",
		"ORACLE_SUPPORT_PRODUCT_VERSION":  "8.2",
		"VERSION_ID":                      "8.2",
		"VERSION":                         "8.2",
		"ID_LIKE":                         "fedora",
		"VARIANT":                         "Server",
		"VARIANT_ID":                      "server",
		"HOME_URL":                        "https://linux.oracle.com/",
		"ORACLE_BUGZILLA_PRODUCT_VERSION": "8.2",
		"NAME":                            "Oracle Linux Server",
	}

	distroIsDetectedBasedOnProperties(t, "ol", "Oracle Linux", "8.2", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverPhoton(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_RELEASE":     "1.0",
		"DISTRIB_CODENAME":    "Photon",
		"DISTRIB_DESCRIPTION": "VMware Photon 1.0",
		"DISTRIB_ID":          "VMware Photon",
	}
	osReleaseProperties := map[string]string{
		"ANSI_COLOR":     "1;34",
		"HOME_URL":       "https://vmware.github.io/photon/",
		"BUG_REPORT_URL": "https://github.com/vmware/photon/issues",
		"NAME":           "VMware Photon",
		"VERSION":        "1.0",
		"ID":             "photon",
		"VERSION_ID":     "1.0",
		"PRETTY_NAME":    "VMware Photon/Linux",
	}

	distroIsDetectedBasedOnProperties(t, "photon", "VMware Photon", "1.0", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverPuppy(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_DESCRIPTION": "FossaPup64 9.0",
		"DISTRIB_ID":          "Puppy",
		"DISTRIB_RELEASE":     "9",
		"DISTRIB_CODENAME":    "FossaPup64",
	}
	osReleaseProperties := map[string]string{
		"VERSION":        "9.5",
		"ID":             "puppy_fossapup64",
		"VERSION_ID":     "9.5",
		"PRETTY_NAME":    "fossapup64 9.5",
		"ANSI_COLOR":     "0;34",
		"CPE_NAME":       "cpe:/o:puppy:puppy_linux:9.5",
		"NAME":           "Puppy",
		"HOME_URL":       "http://puppylinux.com/",
		"SUPPORT_URL":    "http://www.murga-linux.com/puppy/index.php",
		"BUG_REPORT_URL": "https://github.com/puppylinux-woof-CE/woof-CE",
	}

	distroIsDetectedBasedOnProperties(t, "puppy", "Puppy Linux", "9.5", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverRancherOS(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "RancherOS",
		"DISTRIB_RELEASE":     "v1.5.6",
		"DISTRIB_DESCRIPTION": "RancherOS v1.5.6",
	}
	osReleaseProperties := map[string]string{
		"ID":             "rancheros",
		"ID_LIKE":        "",
		"VERSION_ID":     "v1.5.6",
		"PRETTY_NAME":    "RancherOS v1.5.6",
		"HOME_URL":       "http://rancher.com/rancher-os/",
		"SUPPORT_URL":    "https://forums.rancher.com/c/rancher-os",
		"BUG_REPORT_URL": "https://github.com/rancher/os/issues",
		"BUILD_ID":       "",
	}

	distroIsDetectedBasedOnProperties(t, "rancheros", "RancherOS", "v1.5.6", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverScientificLinux6(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/sl-release", "/etc/redhat-release"}) {
			return true, "Scientific Linux release 6.10 (Carbon)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "scientific", "Scientific Linux", "6.10", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverScientificLinux7(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/sl-release", "/etc/redhat-release"}) {
			return true, "Scientific Linux release 7.9 (Nitrogen)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"NAME":                            "Scientific Linux",
		"ID":                              "scientific",
		"REDHAT_SUPPORT_PRODUCT_VERSION":  "7.9",
		"ANSI_COLOR":                      "0;31",
		"CPE_NAME":                        "cpe:/o:scientificlinux:scientificlinux:7.9:GA",
		"BUG_REPORT_URL":                  "mailto:scientific-linux-devel@listserv.fnal.gov",
		"REDHAT_BUGZILLA_PRODUCT":         "Scientific Linux 7",
		"REDHAT_SUPPORT_PRODUCT":          "Scientific Linux",
		"VERSION":                         "7.9 (Nitrogen)",
		"VERSION_ID":                      "7.9",
		"PRETTY_NAME":                     "Scientific Linux 7.9 (Nitrogen)",
		"ID_LIKE":                         "rhel centos fedora",
		"HOME_URL":                        "http://www.scientificlinux.org//",
		"REDHAT_BUGZILLA_PRODUCT_VERSION": "7.9",
	}

	distroIsDetectedBasedOnProperties(t, "scientific", "Scientific Linux", "7.9", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverSLESOld(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/SuSE-release", "/etc/sles-release"}) {
			return true, "SUSE Linux Enterprise Server 12 (x86_64)\nVERSION = 12\nPATCHLEVEL = 1\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "sles", "SUSE Linux", "12", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverSLES12(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/SuSE-release"}) {
			return true, "SUSE Linux Enterprise Server 12 (x86_64)\nVERSION = 12\nPATCHLEVEL = 1\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"NAME":        "SLES",
		"VERSION":     "12",
		"VERSION_ID":  "12",
		"PRETTY_NAME": "SUSE Linux Enterprise Server 12",
		"ID":          "sles",
		"ANSI_COLOR":  "0;32",
		"CPE_NAME":    "cpe:/o:suse:sles:12",
	}

	distroIsDetectedBasedOnProperties(t, "sles", "SUSE Linux", "12", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverSlackwareOld(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/slackware-version"}) {
			return true, "Slackware 14.1"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "slackware", "Slackware", "14.1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverSlackware14(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/slackware-version"}) {
			return true, "Slackware 14.1"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{
		"HOME_URL":       "http://slackware.com/",
		"SUPPORT_URL":    "http://www.linuxquestions.org/questions/slackware-14/",
		"BUG_REPORT_URL": "http://www.linuxquestions.org/questions/slackware-14/",
		"VERSION":        "14.1",
		"ID":             "slackware",
		"VERSION_ID":     "14.1",
		"ANSI_COLOR":     "0;34",
		"CPE_NAME":       "cpe:/o:slackware:slackware_linux:14.1",
		"NAME":           "Slackware",
		"PRETTY_NAME":    "Slackware 14.1",
	}

	distroIsDetectedBasedOnProperties(t, "slackware", "Slackware", "14.1", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverSourceMage(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/sourcemage-release"}) {
			return true, "Source Mage GNU/Linux x86_64-pc-linux-gnu\nInstalled from tarball using chroot image (Grimoire 0.62-stable) generated on Thu Dec  1 01:34:47 UTC 2016\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})
	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "sourcemage", "Source Mage GNU/Linux", "Grimoire 0.62-stable", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverUbuntu510(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "Ubuntu",
		"DISTRIB_RELEASE":     "5.10",
		"DISTRIB_CODENAME":    "breezy",
		"DISTRIB_DESCRIPTION": "Ubuntu (The Breezy Badger Release)",
	}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "ubuntu", "Ubuntu", "5.10", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverUbuntu1204(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "Ubuntu",
		"DISTRIB_RELEASE":     "12.04",
		"DISTRIB_CODENAME":    "precise",
		"DISTRIB_DESCRIPTION": "Ubuntu 12.04 LTS",
	}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "ubuntu", "Ubuntu", "12.04", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverUbuntu1404(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "Ubuntu",
		"DISTRIB_RELEASE":     "14.04",
		"DISTRIB_CODENAME":    "trusty",
		"DISTRIB_DESCRIPTION": "Ubuntu 14.04 LTS",
	}
	osReleaseProperties := map[string]string{
		"NAME":           "Ubuntu",
		"VERSION":        "14.04, Trusty Tahr",
		"ID":             "ubuntu",
		"ID_LIKE":        "debian",
		"PRETTY_NAME":    "Ubuntu 14.04 LTS",
		"VERSION_ID":     "14.04",
		"HOME_URL":       "http://www.ubuntu.com/",
		"SUPPORT_URL":    "http://help.ubuntu.com/",
		"BUG_REPORT_URL": "http://bugs.launchpad.net/ubuntu/",
	}

	distroIsDetectedBasedOnProperties(t, "ubuntu", "Ubuntu", "14.04", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverUbuntu1804(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "Ubuntu",
		"DISTRIB_RELEASE":     "18.04",
		"DISTRIB_CODENAME":    "bionic",
		"DISTRIB_DESCRIPTION": "Ubuntu 18.04.05 LTS",
	}
	osReleaseProperties := map[string]string{
		"NAME":               "Ubuntu",
		"VERSION":            "18.04.5 LTS (Bionic Beaver)",
		"ID":                 "ubuntu",
		"ID_LIKE":            "debian",
		"PRETTY_NAME":        "Ubuntu 18.04.5 LTS",
		"VERSION_ID":         "18.04",
		"HOME_URL":           "http://www.ubuntu.com/",
		"SUPPORT_URL":        "http://help.ubuntu.com/",
		"BUG_REPORT_URL":     "http://bugs.launchpad.net/ubuntu/",
		"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
		"VERSION_CODENAME":   "bionic",
		"UBUNTU_CODENAME":    "bionic",
	}

	distroIsDetectedBasedOnProperties(t, "ubuntu", "Ubuntu", "18.04", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverUbuntu2004(t *testing.T) {
	lsbProperties := map[string]string{
		"DISTRIB_ID":          "Ubuntu",
		"DISTRIB_RELEASE":     "20.04",
		"DISTRIB_CODENAME":    "focal",
		"DISTRIB_DESCRIPTION": "Ubuntu 20.04.1 LTS",
	}
	osReleaseProperties := map[string]string{
		"NAME":               "Ubuntu",
		"VERSION":            "20.04.1 LTS (Focal Fossa)",
		"ID":                 "ubuntu",
		"ID_LIKE":            "debian",
		"PRETTY_NAME":        "Ubuntu 20.04.1 LTS",
		"VERSION_ID":         "20.04",
		"HOME_URL":           "http://www.ubuntu.com/",
		"SUPPORT_URL":        "http://help.ubuntu.com/",
		"BUG_REPORT_URL":     "http://bugs.launchpad.net/ubuntu/",
		"PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
		"VERSION_CODENAME":   "focal",
		"UBUNTU_CODENAME":    "focal",
	}

	distroIsDetectedBasedOnProperties(t, "ubuntu", "Ubuntu", "20.04", lsbProperties,
		osReleaseProperties)
}

func TestDiscoverYellowDog(t *testing.T) {
	originalReadFileFunc := readFileFunc
	readFileFunc = func(filePaths ...string) (bool, string) {
		if reflect.DeepEqual(filePaths, []string{"/etc/yellowdog-release"}) {
			return true, "Yellow Dog Linux release 4.0 (Orion)\n"
		} else {
			return false, ""
		}
	}
	t.Cleanup(func() {
		readFileFunc = originalReadFileFunc
	})

	lsbProperties := map[string]string{}
	osReleaseProperties := map[string]string{}

	distroIsDetectedBasedOnProperties(t, "yellow-dog", "Yellow Dog Linux", "4.0", lsbProperties,
		osReleaseProperties)
}

func distroIsDetectedBasedOnProperties(t *testing.T, id string, name string, version string, lsbProperties map[string]string,
	osReleaseProperties map[string]string) {
	distro := discoverDistroFromProperties(lsbProperties, osReleaseProperties)
	if distro.ID != id {
		t.Errorf("Linux distro id was not detected correctly. Expected (%s) was (%s).", id, distro.ID)
	}
	if distro.Name != name {
		t.Errorf("Linux distro name was not detected correctly. Expected (%s) was (%s).", name, distro.Name)
	}
	if distro.Version != version {
		t.Errorf("Linux distro version was not detected correctly. Expected (%s) was (%s).", version, distro.Version)
	}
	if !reflect.DeepEqual(lsbProperties, distro.LsbRelease) {
		t.Error("lsb properties weren't copied properly into distro struct")
	}
	if !reflect.DeepEqual(osReleaseProperties, distro.OsRelease) {
		t.Error("OS release properties weren't copied properly into distro struct")
	}
}
