# lazyhetzner
A TUI for managing Hetzner Cloud servers with ease.


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
  "default_project": "production"
}
```


