[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)
# Distribution Detector (Distro Detect)

Distro Detect is a command line utility or golang library that detects the
distribution of an operating system such as Ubuntu, CentOS, Fedora, etc.
It performs detection by analyzing the contents of files on the system and
does not call any external programs. Additionally, it can be pointed to a
file system root so it can analyze a copy of an existing system.

## Usage

Execute the binary `disto-detect` to output the distribution of the current
running system. For example, here we have run it on an Ubuntu system.

```
$ ./distro-detect 
Distro ID: ubuntu
Distro Name: Ubuntu
Distro Version: 18.04
Distro LSB DISTRIB_RELEASE: 18.04
Distro LSB DISTRIB_CODENAME: bionic
Distro LSB DISTRIB_DESCRIPTION: Ubuntu 18.04.5 LTS
Distro LSB DISTRIB_ID: Ubuntu
Distro OS PRETTY_NAME: Ubuntu 18.04.5 LTS
Distro OS PRIVACY_POLICY_URL: https://www.ubuntu.com/legal/terms-and-policies/privacy-policy
Distro OS VERSION_CODENAME: bionic
Distro OS NAME: Ubuntu
Distro OS VERSION: 18.04.5 LTS (Bionic Beaver)
Distro OS ID: ubuntu
Distro OS ID_LIKE: debian
Distro OS UBUNTU_CODENAME: bionic
Distro OS VERSION_ID: 18.04
Distro OS HOME_URL: https://www.ubuntu.com/
Distro OS SUPPORT_URL: https://help.ubuntu.com/
Distro OS BUG_REPORT_URL: https://bugs.launchpad.net/ubuntu/
```

For more options, run the command with the `-help` flag.

### Selecting Fields

To only output a specific fields, invoke the command with the `-fields` flag.

```
$ ./distro-detect -fields id,version
  Distro ID: ubuntu
  Distro Version: 18.04
```

### Output Formats

To output only the distribution without labels, combine the `-fields` flag with
the `-format text-no-labels` flag. These options may be useful when using the
command from shell scripts. 

```
./distro-detect -fields id -format text-no-labels
ubuntu
```

To output the results of the detection to JSON, specify the `-format json` or
`-format json-one-line` flags.

```
{
  "name": "Ubuntu",
  "id": "ubuntu",
  "version": "18.04",
  "lsb_release": {
    "DISTRIB_CODENAME": "bionic",
    "DISTRIB_DESCRIPTION": "Ubuntu 18.04.5 LTS",
    "DISTRIB_ID": "Ubuntu",
    "DISTRIB_RELEASE": "18.04"
  },
  "os_release": {
    "BUG_REPORT_URL": "https://bugs.launchpad.net/ubuntu/",
    "HOME_URL": "https://www.ubuntu.com/",
    "ID": "ubuntu",
    "ID_LIKE": "debian",
    "NAME": "Ubuntu",
    "PRETTY_NAME": "Ubuntu 18.04.5 LTS",
    "PRIVACY_POLICY_URL": "https://www.ubuntu.com/legal/terms-and-policies/privacy-policy",
    "SUPPORT_URL": "https://help.ubuntu.com/",
    "UBUNTU_CODENAME": "bionic",
    "VERSION": "18.04.5 LTS (Bionic Beaver)",
    "VERSION_CODENAME": "bionic",
    "VERSION_ID": "18.04"
  }
}
```

## Author
**Elijah Zupancic**

* Twitter: [@elijah_zupancic](https://twitter.com/elijah_zupancic)