# lazyhetzner
A TUI for managing Hetzner Cloud servers with ease. Written in Golang, using Bubble Tea for the terminal user interface, and the hcloud Go client for interacting with the Hetzner Cloud API.


## Build from Source
To build lazyhetzner from source, you need to have Go installed on your system. Follow these steps:
1. Clone the repository:
   ```bash
   git clone foobar/lazyhetzner.git
    cd lazyhetzner
    ```

2. Build the project:
    ```bash
    go build -o lazyhetzner
    ```
3. Make the binary executable:
    ```bash
    chmod +x lazyhetzner
    ```
4. Move the binary to a directory in your PATH:
    ```bash
    sudo mv lazyhetzner /usr/local/bin/
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


