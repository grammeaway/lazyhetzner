# lazyhetzner
A TUI for managing Hetzner Cloud servers with ease. Written in Golang, using Bubble Tea for the terminal user interface, and the hcloud Go client for interacting with the Hetzner Cloud API.


## Installation
### Installing with Go on your system
Clone the repository and run the following commands to install the dependencies and build the binary:

```bash
go install github.com/lazyhetzner/awsbreeze@latest
```
This will install the `awsbreeze` binary in your `$GOPATH/bin` directory. Make sure to add this directory to your `PATH` environment variable if it's not already included.

### Installing with pre-built binaries
Download the latest release matching your OS from the [releases page](https://github.com/grammeaway/awsbreeze/releases).

Unzip the downloaded file and move the `awsbreeze` binary to a directory in your `PATH`, such as `/usr/local/bin` on Linux or macOS, or `C:\Program Files\` on Windows.

### Installing the nightly build (through Go)
If you want to try the latest features and bug fixes, you can install the nightly build by running the following command:

```bash
go install github.com/grammeaway/lazyhetzner@main
```

## Verifying the installation
After installing, you can verify that `awsbreeze` is installed correctly by running the following command in your terminal:

```bash
lazyhetzner version
```


## Persisting and Managing Multiple Projects
To persist and manage multiple projects, you need to create a ```config.json``` in the ```~/.config/lazyhetzner/``` directory. This file should contain a JSON object with the following structure:

```json
{
  "projects": [
    {
      "name": "production",
      "token": "your-production-token"
    },
    {
      "name": "staging", 
      "token": "your-staging-token"
    }
  ],
  "default_project": "production",
  "default_terminal": "" // Optional: specify a default terminal emulator, e.g., "foot", "alacritty", "kitty"
}
```


