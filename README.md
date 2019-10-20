# Personal Home Automation Service

## Installation

You can use PHAS on Raspberry Pi, but also on your personal computer.

Below you'll find installation instructions for all supported platforms:

- [Linux - Raspberry Pi and Desktop](#Linux)
- [Windows](#Windows)
- [macOS](#macOS)

### Linux

These instructions apply both to Raspberry Pi and Desktop distributions of Linux (tested under Debian and Ubuntu).

First, make sure the system is up to date by using the following commands:
```shell script
sudo apt-get update
sudo apt-get upgrade
```

Then, install Go 1.13.3 or newer installed from the binary distribution.
While Go 1.11 is still supported, it's better to use a newer version of
Go and benefit from all the improvements in Go Modules, compilation
speed and debugging experience.

For Raspberry Pi, use the below Go distribution. If you are on a Linux Desktop,
then skip to the Linux Desktop instructions below this section.

```shell script
cd /tmp
wget https://dl.google.com/go/go1.13.3.linux-armv6l.tar.gz
tar -zxvf go1.13.3.linux-armv6l.tar.gz
sudo mv go /usr/local/
rm go1.13.3.linux-armv6l.tar.gz
```

For Linux Desktop, you can skip this if you are on a Raspberry Pi.
```shell script
cd /tmp
wget https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz
tar -zxvf go1.13.3.linux-amd64.tar.gz
sudo mv go /usr/local/
rm go1.13.3.linux-amd64.tar.gz
```

Manually add Go's binary files, and your `$GOPATH/bin` binaries to your `$PATH`.
E.g., place `export PATH="${PATH}:/usr/local/go/bin:/home/pi/go/bin"` at the end of `~/.bash_profile` such as:
```shell script
# Include the Go SDK and GOPATH/bin in PATH
export PATH="${PATH}:/usr/local/go/bin:/home/pi/go/bin"
```

If the file does not exist, then create it.

Now, let's install PHAS:

```shell script
# clone this repository
git clone https://github.com/dlsniper/phas.git /home/pi/phas
cd /home/pi/phas

# Copy the required system libraries in the correct paths.
# You can skip these if you will always use the build commands with the correct CGO environment variables configured, see below.
sudo cp ./lib/picovoice.h /usr/local/include/
sudo cp ./lib/pv_porcupine.h /usr/local/include/

# If you are on Linux Desktop, then replace raspberrypi_arm11 with linux
sudo cp ./lib/raspberrypi_arm11/libpv_porcupine.* /usr/local/lib/

# Enable the vendoring mode for Go Modules so that no additional downloads
# are needed to compile the application
export GOFLAGS=-mod=vendor

# This is the target on which PHAS runs. The following values are useful:
# windows     - for Windows OS
# macos       - for macOS
# raspberrypi - for Raspberry PI 3 Linux OS
# linux       - for Linux OS
export PHAS_OS=raspberrypi

# Build the binary
go build -o phas
```

#### Linux Destkop

You need to install the following dependencies:

```shell script
apt-get install pkg-config portaudio19-dev
```

### Windows

For a Windows user, the compilation will be a bit more tricky.

First, you need to have [msys2](https://www.msys2.org/) installed.

Add `C:\msys64\mingw64\bin` to your Windows `PATH` environment variable

Then, open the `MSYS2 MinGW 64-bit` command prompt and run the following commands:

```shell script
# Ensure msys2 is updated. You might need to run this command a couple of times
pacman -Syu

# Install build tools dependencies
pacman -S mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config

# Install the dependency for recording audio from the microphone
pacman -S mingw-w64-x86_64-portaudio
```

#### Building on Windows

Building PHAS on Windows needs to be performed like this:

- set the following environment variables:
```shell script
export CGO_CFLAGS=-I<path to PHAS root directory>\lib
export CGO_LDFLAGS=-L<path to PHAS root directory>\lib\windows
```
Also, include `C:\msys64\mingw64\bin` in your PATH on Windows.

For PHAS to work, you need to configure the following environment variable `PHAS_OS=raspberrypi`.

- then run `go build -o phas`

### macOS

For macOS, you need to run the following commands:

```shell script
brew install portaudio pkg-config
```

During the compilation step, you also need to set the following environment variable:

```shell script
export DYLD_LIBRARY_PATH="${DYLD_LIBRARY_PATH}:<path to PHAS root directory>/lib/macos"
```

## Getting GCP credentials for the voice APIs

Before we can run the application, we need to get valid GCP credentials
to make use of the voice APIs.

If you don't yet have a Google Cloud Platform account yet, then first head over
[to this link and create one](https://console.cloud.google.com/projectcreate).

Then, read the link here about [obtaining and providing service account credentials manually](https://cloud.google.com/docs/authentication/production#obtaining_and_providing_service_account_credentials_manually).
Select ` Service User | API Keys Viewer ` from the ` Role ` and the ` json ` format.

After this, go to [APIs & Services](https://console.cloud.google.com/apis/dashboard)
and select ` +Enable APIs and Services` button. Here, you'll need to enable
the following services:
- ` Cloud Speech-to-Text API `, ` Cloud Speech-to-Text API Length Standard `
option should be enough
- ` Cloud Text-to-Speech API `, ` WaveNet ` option should be enough

Finally, place the downloaded file in a location such as ` /home/pi/phas-gcp-key.json `.
Then, add the following environment variables to your ` .bash_profile `:

```shell script
# These are needed to tell PHAS what GCP project to use and what credentials
export GCP_PROJECT_ID=<your GCP project ID>
export GOOGLE_APPLICATION_CREDENTIALS="/home/pi/phas-gcp-key.json"
```

If you are on Windows, then place the file under `C:\Users\<username>\phas-gcp-key.json`
and then set the environment variables accordingly:
```
GCP_PROJECT_ID=<your GCP project ID>
GOOGLE_APPLICATION_CREDENTIALS="/home/pi/phas-gcp-key.json"
```

## Running the application

Since we installed all dependencies and everything is up to date,
let's run the application. Execute `.\phas.exe` if you are Windows,
and `./phas` if you are on all other supported platforms.

## License 

This repository and all code from it is licensed under the [Apache 2 license](License).

The code under `lib` directory belongs to the [Picovoice](https://picovoice.ai/)
[Porcupine library](https://github.com/Picovoice/Porcupine), also available
under Apache 2 license.

The code under the `vendor` directory belongs to their original creators and
is available under their respective license. It is included here to enable the
smooth operation of the workshop.