# lazyhetzner
A TUI for managing Hetzner Cloud servers with ease. Written in Golang, using Bubble Tea for the terminal user interface, and the hcloud Go client for interacting with the Hetzner Cloud API.

Built using Golang, and the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework.

For the foreseeable future, this TUI is primarily meant for interacting with existing resources within Hetzner Cloud, rather than creating or deleting them. Creation and deletion of resources might be added in the future, but until then, consider using IaC tools like [Terraform](https://www.terraform.io/) for resource management.

The tool is heavily inspired by lovely TUI projects like [lazydocker](https://github.com/jesseduffield/lazydocker) and [lazysql](https://github.com/jorgerojas26/lazysql), both tools that I use on an almost daily basis. I hope this Hetzner TUI will be a worthy entry into the "lazy" family.

Table of Contents:
- [lazyhetzner](#lazyhetzner)
  - [Features](#features)
  - [Installation](#installation)
    - [Installing with Go on your system](#installing-with-go-on-your-system)
    - [Installing with pre-built binaries](#installing-with-pre-built-binaries)
    - [Installing the nightly build (through Go)](#installing-the-nightly-build-through-go)
  - [Verifying the installation](#verifying-the-installation)
  - [Persisting and Managing Multiple Projects](#persisting-and-managing-multiple-projects)
  - [Known Issues](#known-issues)
  - [Roadmap](#roadmap)
  - [Contributing](#contributing)

## Features
Notable features include:
- **View resource labels**: View labels for various Hetzner resources.
- **SSH into servers**: SSH into your Hetzner Cloud servers directly from the TUI, either in a new terminal window or in the current terminal.
- **SSH into servers in tmux or Zellij**: The TUI supports lunching SSH sessions in your current tmux or Zellij session, allowing you to manage your servers without leaving your current workflow.
- **Copy IP addresses**: Easily copy IP addresses of your Hetzner Cloud servers to the clipboard.

## Installation
### Installing with Go on your system
Clone the repository and run the following commands to install the dependencies and build the binary:

```bash
go install github.com/lazyhetzner@latest
```
This will install the `lazyhetzner` binary in your `$GOPATH/bin` directory. Make sure to add this directory to your `PATH` environment variable if it's not already included.

### Installing with pre-built binaries
Download the latest release matching your OS from the [releases page](https://github.com/grammeaway/awsbreeze/releases).

Unzip the downloaded file and move the `awsbreeze` binary to a directory in your `PATH`, such as `/usr/local/bin` on Linux or macOS, or `C:\Program Files\` on Windows.

### Installing the nightly build (through Go)
If you want to try the latest features and bug fixes, you can install the nightly build by running the following command:

```bash
go install github.com/grammeaway/lazyhetzner@main
```

## Verifying the installation
After installing, you can verify that `lazyhetzner` is installed correctly by running the following command in your terminal:

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

## Known Issues
Currently, the flow for persisting project configurations in the TUI itself, is quite clunky and unstable. It also doesn't offer a convenient way of setting your defualt project (it'll just tag on to whatever project you set first). It is highly recommended to manually create the `config.json` file as described above.

## Roadmap
- [ ] Improve the project configuration flow in the TUI.
- [ ] Create SSH keys in the TUI, to be used in the creation of new servers.
- [ ] Add visualizations for for the resources:
    - [ ] Firewalls
    - [ ] Floating IPs
- [ ] Add visualizations for the sub-resources:
    - [ ] Subnets
    - [ ] Server backups
    - [ ] Server snapshots
    - [ ] Server placement groups
    - [ ] Server primary IPs

## Contributing
If you would like to contribute to this project, please fork the repository and submit a pull request. Contributions are welcome!
