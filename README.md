# sws-sdk-go

Official Go SDK for the **SWS** cloud platform.

## Install

```bash
go get github.com/savannaacloud/sws-sdk-go
```

## Quickstart

```go
package main

import (
    "context"
    "fmt"

    sws "github.com/savannaacloud/sws-sdk-go"
)

func main() {
    client := sws.NewClient("sws_...", sws.WithRegion("ng-lagos-1"))
    ctx := context.Background()

    // List virtual machines
    instances, err := client.Compute.ListInstances(ctx)
    if err != nil { panic(err) }
    for _, vm := range instances {
        fmt.Printf("%s — %s\n", vm.Name, vm.Status)
    }

    // Launch an instance
    vm, err := client.Compute.CreateInstance(ctx, &sws.CreateInstanceOpts{
        Name:      "web-01",
        Image:     "ubuntu-22.04",
        Plan:      "m1.medium",
        NetworkID: "net-uuid",
        KeyName:   "my-key",
    })
    if err != nil { panic(err) }
    fmt.Println("Created:", vm.ID)
}
```

## Configuration

| Option            | Env var         | Default                |
| ----------------- | --------------- | ---------------------- |
| `WithBaseURL`     | `SWS_BASE_URL`  | `https://savannaa.com` |
| `WithRegion`      | `SWS_REGION`    | `ng-lagos-1`           |
| `WithTimeout`     | —               | 30s                    |
| `WithHTTPClient`  | —               | `http.DefaultClient` (with timeout) |
| (api key)         | `SWS_API_KEY`   | _(required)_           |

## Resources

| Service              | Operations |
| -------------------- | ---------- |
| `client.Compute`     | List/Get/Create/Delete instances; Start/Stop/Reboot/Resize; ListPlans, ListImages, Keypairs CRUD |
| `client.Network`     | Networks, Subnets, SecurityGroups + Rules, PublicIPs (allocate/assign/release) |
| `client.Storage`     | Volumes (CRUD + Attach/Detach) |
| `client.Database`    | Managed databases (mysql, postgresql, …) |

## Error handling

```go
import "errors"

_, err := client.Compute.CreateInstance(ctx, opts)
switch {
case errors.Is(err, sws.ErrQuota):
    // request quota bump or smaller plan
case errors.Is(err, sws.ErrValidation):
    // fix the payload
case errors.Is(err, sws.ErrAuthentication):
    // bad/expired API key
}

// Need the status code or raw body? Type-assert:
var apiErr *sws.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Message)
}
```

## Development

```bash
go test ./...           # unit tests use httptest, no live API required
go vet ./...
```

## License

MIT
