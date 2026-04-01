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
    "context"
    "fmt"
    "log"

    pocketenv "github.com/pocketenv-io/pocketenv-go"
)

func main() {
    client := pocketenv.New(
        pocketenv.WithToken("your-api-token")
    )

    ctx := context.Background()

    // Create a sandbox — returns a handle with data + methods combined
    sb, err := client.CreateSandbox(ctx, pocketenv.CreateSandboxInput{
        Base: "ubuntu",
        Name: "my-sandbox",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Created sandbox:", sb.ID, sb.Status)

    // Call methods directly on the returned sandbox — no extra step needed
    result, err := sb.Exec(ctx, "echo hello")
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

## Sandboxes

```go
// Create — returns *Sandbox with data fields AND methods ready to use
sb, err := client.CreateSandbox(ctx, pocketenv.CreateSandboxInput{Base: "ubuntu"})
fmt.Println(sb.ID, sb.Name, sb.Status) // data fields directly accessible

// Get — same: returns *Sandbox
sb, err := client.GetSandbox(ctx, "sandbox-id-or-name")

// List — returns []*Sandbox, each usable as a handle
sandboxes, total, err := client.ListSandboxes(ctx, 0, 20)
for _, sb := range sandboxes {
    fmt.Println(sb.ID, sb.Status)
}

// If you only have an ID (e.g. from config), get a lightweight handle:
sb = client.Sandbox("sandbox-id")

// Actions — call directly on any *Sandbox
err = sb.Start(ctx)
err = sb.Stop(ctx)
err = sb.Delete(ctx)
err = sb.Refresh(ctx) // re-fetches data fields from API

// Execute a command
result, err := sb.Exec(ctx, "npm install")
fmt.Println(result.Stdout, result.Stderr, result.ExitCode)

// Ports
err = sb.ExposePort(ctx, 3000, "web app")
err = sb.UnexposePort(ctx, 3000)
ports, err := sb.GetExposedPorts(ctx)

// SSH keys
keys, err := sb.GetSshKeys(ctx)
err = sb.PutSshKeys(ctx, publicKey, privateKey)
```

## Secrets

```go
sc := client.Sandbox("sandbox-id").Secrets()
// or: client.Secrets("sandbox-id")

secret, err := sc.Add(ctx, "DB_PASSWORD", "s3cr3t")
secrets, total, err := sc.List(ctx, 0, 20)
secret, err = sc.Get(ctx, "secret-id")
secret, err = sc.Update(ctx, "secret-id", "DB_PASSWORD", "new-value")
err = sc.Delete(ctx, "secret-id")
```

## Environment Variables

```go
vc := client.Sandbox("sandbox-id").Variables()
// or: client.Variables("sandbox-id")

variable, err := vc.Add(ctx, "PORT", "8080")
variables, total, err := vc.List(ctx, 0, 20)
variable, err = vc.Get(ctx, "variable-id")
variable, err = vc.Update(ctx, "variable-id", "PORT", "9090")
err = vc.Delete(ctx, "variable-id")
```

## Files

```go
fc := client.Sandbox("sandbox-id").Files()
// or: client.Files("sandbox-id")

err = fc.Add(ctx, "/app/config.json", `{"port": 8080}`)
files, total, err := fc.List(ctx, 0, 20)
file, err := fc.Get(ctx, "file-id")
err = fc.Update(ctx, "file-id", "/app/config.json", `{"port": 9090}`)
err = fc.Delete(ctx, "file-id")
```

## Volumes

```go
vc := client.Sandbox("sandbox-id").Volumes()
// or: client.Volumes("sandbox-id")

err = vc.Add(ctx, "my-volume", "/data")
volumes, total, err := vc.List(ctx, 0, 20)
volume, err := vc.Get(ctx, "volume-id")
err = vc.Update(ctx, "volume-id", "my-volume", "/mnt/data")
err = vc.Delete(ctx, "volume-id")
```

## Services

```go
sc := client.Sandbox("sandbox-id").Services()
// or: client.Services("sandbox-id")

err = sc.Add(ctx, pocketenv.AddServiceInput{
    Name:    "web",
    Command: "node server.js",
    Ports:   []int{3000},
})
services, err := sc.List(ctx)
err = sc.Update(ctx, "service-id", pocketenv.UpdateServiceInput{Name: "web-v2"})
err = sc.Start(ctx, "service-id")
err = sc.Stop(ctx, "service-id")
err = sc.Restart(ctx, "service-id")
err = sc.Delete(ctx, "service-id")
```

## License

[MIT](LICENSE)
