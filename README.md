# pocketenv-go

Official Go SDK for the [Pocketenv](https://pocketenv.io) API.

## Installation

```bash
go get github.com/pocketenv-io/pocketenv-go
```

## Quick start

```go
package main

import (
    "fmt"
    "log"

    pocketenv "github.com/pocketenv-io/pocketenv-go"
)


func main() {
    // Token is read from POCKETENV_TOKEN env var or ~/.pocketenv/token.json
    client, err := pocketenv.New()
    if err != nil {
        log.Fatal(err)
    }

    // Create a sandbox — returns a handle with data + methods combined
    sb, err := client.CreateSandbox(pocketenv.CreateSandboxInput{
        Base: "openclaw",
        Name: "my-sandbox",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created sandbox:", sb.ID, sb.Status)

    // Call methods directly on the returned sandbox — no extra step needed
    result, err := sb.Exec("echo", "hello")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.Stdout) // "hello\n"
}
```

## Client options

| Option | Description |
|---|---|
| `WithToken(token)` | Bearer token for authentication |
| `WithBaseURL(url)` | API base URL (default: `https://api.pocketenv.io`) |

If `WithToken` is not used, the token is resolved in this order:
1. `POCKETENV_TOKEN` environment variable
2. `~/.pocketenv/token.json` (`{"token": "..."}`)

`New` returns an error if none of the above provide a token.

## Sandboxes

```go
// Create — returns *Sandbox with data fields AND methods ready to use
sb, err := client.CreateSandbox(pocketenv.CreateSandboxInput{Base: "openclaw"})
fmt.Println(sb.ID, sb.Name, sb.Status) // data fields directly accessible

// Get — same: returns *Sandbox
sb, err := client.GetSandbox("sandbox-id-or-name")

// List — returns []*Sandbox, each usable as a handle
sandboxes, total, err := client.ListSandboxes(0, 20)
for _, sb := range sandboxes {
    fmt.Println(sb.ID, sb.Status)
}

// If you only have an ID (e.g. from config), get a lightweight handle:
sb = client.Sandbox("sandbox-id")

// Actions — call directly on any *Sandbox
err = sb.Start()
err = sb.Stop()
err = sb.Delete()
err = sb.Refresh() // re-fetches data fields from API

// Execute a command
result, err := sb.Exec("npm", "install")
fmt.Println(result.Stdout, result.Stderr, result.ExitCode)

// Ports
err = sb.ExposePort(3000, "web app")
err = sb.UnexposePort(3000)
ports, err := sb.GetExposedPorts()

// SSH keys
keys, err := sb.GetSshKeys()
err = sb.PutSshKeys(publicKey, privateKey)
```

## Secrets

```go
sc := client.Sandbox("sandbox-id").Secrets()
// or: client.Secrets("sandbox-id")

secret, err := sc.Add("DB_PASSWORD", "s3cr3t")
secrets, total, err := sc.List(0, 20)
secret, err = sc.Get("secret-id")
secret, err = sc.Update("secret-id", "DB_PASSWORD", "new-value")
err = sc.Delete("secret-id")
```

## Environment Variables

```go
vc := client.Sandbox("sandbox-id").Variables()
// or: client.Variables("sandbox-id")

variable, err := vc.Add("PORT", "8080")
variables, total, err := vc.List(0, 20)
variable, err = vc.Get("variable-id")
variable, err = vc.Update("variable-id", "PORT", "9090")
err = vc.Delete("variable-id")
```

## Files

```go
fc := client.Sandbox("sandbox-id").Files()
// or: client.Files("sandbox-id")

err = fc.Add("/app/config.json", `{"port": 8080}`)
files, total, err := fc.List(0, 20)
file, err := fc.Get("file-id")
err = fc.Update("file-id", "/app/config.json", `{"port": 9090}`)
err = fc.Delete("file-id")
```

## Volumes

```go
vc := client.Sandbox("sandbox-id").Volumes()
// or: client.Volumes("sandbox-id")

err = vc.Add("my-volume", "/data")
volumes, total, err := vc.List(0, 20)
volume, err := vc.Get("volume-id")
err = vc.Update("volume-id", "my-volume", "/mnt/data")
err = vc.Delete("volume-id")
```

## Services

```go
sc := client.Sandbox("sandbox-id").Services()
// or: client.Services("sandbox-id")

err = sc.Add(pocketenv.AddServiceInput{
    Name:    "web",
    Command: "node server.js",
    Ports:   []int{3000},
})
services, err := sc.List()
err = sc.Update("service-id", pocketenv.UpdateServiceInput{Name: "web-v2"})
err = sc.Start("service-id")
err = sc.Stop("service-id")
err = sc.Restart("service-id")
err = sc.Delete("service-id")
```

## License

[MIT](LICENSE)
