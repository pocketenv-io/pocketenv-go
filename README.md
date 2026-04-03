# pocketenv-go

Official Go SDK for [Pocketenv](https://pocketenv.io).

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

    sb, err := client.CreateSandbox(pocketenv.CreateSandboxInput{
        Base: "openclaw",
        Name: "my-sandbox",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created sandbox:", sb.ID, sb.Status)

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

## Error handling

All API errors are returned as `*pocketenv.Error`, which exposes the HTTP status
code so callers can handle specific cases without string-parsing:

```go
_, err := sb.Variables().Get("unknown-id")
if e, ok := err.(*pocketenv.Error); ok && e.StatusCode == 404 {
    // not found
}
```

## Sandboxes

```go
// Create
sb, err := client.CreateSandbox(pocketenv.CreateSandboxInput{Base: "openclaw"})
fmt.Println(sb.ID, sb.Name, sb.Status)

// Get
sb, err = client.GetSandbox("sandbox-id")

// List — returns Page[*Sandbox] with Items and Total
page, err := client.ListSandboxes(0, 20)
fmt.Println(page.Total)
for _, sb := range page.Items {
    fmt.Println(sb.ID, sb.Status)
}

// Lightweight handle for a known ID (no API call)
sb = client.SandboxRef("sandbox-id")

// Actions
err = sb.Start()
err = sb.Stop()
err = sb.Delete()
err = sb.Refresh() // re-fetches data from API

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

Values are encrypted client-side before being sent to the API.

```go
sc := sb.Secrets()
// or: client.Secrets("sandbox-id")

secret, err := sc.Create("DB_PASSWORD", "s3cr3t")

page, err := sc.List(0, 20)
fmt.Println(page.Total)

secret, err = sc.Get("secret-id")

// Update and Delete are called on the secret itself — no ID needed
secret, err = secret.Update("DB_PASSWORD", "new-value")
err = secret.Delete()
err = secret.Refresh()
```

## Environment Variables

```go
vc := sb.Variables()
// or: client.Variables("sandbox-id")

variable, err := vc.Create("PORT", "8080")

page, err := vc.List(0, 20)
fmt.Println(page.Total)

variable, err = vc.Get("variable-id")

variable, err = variable.Update("PORT", "9090")
err = variable.Delete()
err = variable.Refresh()
```

## Files

File contents are encrypted client-side before being sent to the API.

```go
fc := sb.Files()

err = fc.Create("/app/config.json", `{"port": 8080}`)

page, err := fc.List(0, 20)
fmt.Println(page.Total)

file, err := fc.Get("file-id")

err = file.Update("/app/config.json", `{"port": 9090}`)
err = file.Delete()
err = file.Refresh()
```

## Volumes

```go
vc := sb.Volumes()

err = vc.Create("my-volume", "/data")

page, err := vc.List(0, 20)
fmt.Println(page.Total)

volume, err := vc.Get("volume-id")

err = volume.Update("my-volume", "/mnt/data")
err = volume.Delete()
err = volume.Refresh()
```

## Services

```go
sc := sb.Services()

err = sc.Create(pocketenv.AddServiceInput{
    Name:    "web",
    Command: "node server.js",
    Ports:   []int{3000},
})

services, err := sc.List()

svc := services[0]
err = svc.Update(pocketenv.UpdateServiceInput{Name: "web-v2"})
err = svc.Start()
err = svc.Stop()
err = svc.Restart()
err = svc.Delete()
```

## License

[MIT](LICENSE)
