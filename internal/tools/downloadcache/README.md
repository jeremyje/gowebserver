# downloadcache

A cache for the build's downloads, run on each CI machine. `make clean` deletes `build/`, so every CI run downloads the toolchain again: linters, Terraform, Trivy, ffmpeg, Playwright's browsers and Go modules, about 900 MB. With this cache, the first run downloads them once and later runs read them from local disk.

It serves any public `http` or `https` URL under a path that names it. `https://github.com/a/b` is served at `http://localhost:3142/https/github.com/a/b`. The cache follows each server's own caching headers:

- A file is served from disk for as long as its server says it's fresh. Versioned release files and Go module zips stay fresh for hours to a year.
- After that, the cache sends a conditional request. If the file hasn't changed, only a small reply comes back.
- When the server can't be reached or fails, the kept copy is served anyway.

Only `200` responses are kept, and nothing marked `no-store`. Credentials are never sent upstream, so only public files are cached. Files used least recently are removed once the cache passes `-max-gb`, which defaults to 50.

## How the build uses it

Set `DOWNLOAD_MIRROR` to the cache's address in the machine's environment. The Makefiles send downloads through it, and fall back to the original server when it fails (see `Makefile_core.mk`). When `DOWNLOAD_MIRROR` is unset, nothing changes.

| Download | How it reaches the cache | When the cache is down |
| --- | --- | --- |
| Release files (`$(call download,...)`) | URL rewritten for curl | Falls back to a direct download |
| Go modules (`go install`, `make deps`) | `GOPROXY=<cache>/https/proxy.golang.org\|https://proxy.golang.org,direct` | Go falls back by itself |
| ffmpeg (`internal/tools/installffmpeg`) | An `http.Client` transport | Falls back to a direct download |
| npm packages | `--registry` | Skipped: the cache is checked first (`/healthz`) |
| Playwright's browsers | `PLAYWRIGHT_*_DOWNLOAD_HOST` | Skipped: the cache is checked first (`/healthz`) |

It doesn't cover `actions/setup-go` (a self-hosted runner keeps Go in its tool cache), `apt-get` in the Linux job (`apt-cacher-ng` does that job for apt), or base images for Docker builds (use Docker's `registry-mirrors`).

## Running it on a CI machine

Install it with Go:

```sh
go install github.com/cloudfra/gowebserver/internal/tools/downloadcache@latest
```

By default it listens on `127.0.0.1:3142` and keeps files in the user cache directory: `~/.cache/downloadcache` on Linux, `%LocalAppData%\downloadcache` on Windows. Any other listening address lets other machines fetch any URL through it, so pass `-hosts` with the hosts to allow.

Then give the runner the address. A self-hosted runner reads a `.env` file in its install directory when it starts, so no workflow has to change:

```text
DOWNLOAD_MIRROR=http://localhost:3142
```

Restart the runner service after editing the file.

### Linux (systemd)

```ini
# /etc/systemd/system/downloadcache.service
[Unit]
Description=Build download cache
After=network-online.target

[Service]
User=runner
ExecStart=/home/runner/go/bin/downloadcache -dir /var/cache/downloadcache -max-gb 50
Restart=always

[Install]
WantedBy=multi-user.target
```

```sh
sudo mkdir -p /var/cache/downloadcache && sudo chown runner /var/cache/downloadcache
sudo systemctl enable --now downloadcache
```

### Windows (scheduled task)

`downloadcache` is a plain console program, not a Windows service, so start it with a scheduled task at boot, as the runner's user:

```powershell
$action  = New-ScheduledTaskAction -Execute "$env:USERPROFILE\go\bin\downloadcache.exe" -Argument '-dir C:\downloadcache -max-gb 50'
$trigger = New-ScheduledTaskTrigger -AtStartup
$settings = New-ScheduledTaskSettingsSet -RestartCount 999 -RestartInterval (New-TimeSpan -Minutes 1) -ExecutionTimeLimit 0
Register-ScheduledTask -TaskName downloadcache -Action $action -Trigger $trigger -Settings $settings -User 'NT AUTHORITY\NETWORK SERVICE'
Start-ScheduledTask -TaskName downloadcache
```

Use the account the runner runs as. The Windows runner in this repository's CI logs runs as `NETWORK SERVICE`, and that account needs write access to `-dir`.

## Checking it

```sh
curl -sI http://localhost:3142/healthz
curl -s -o /dev/null -D - http://localhost:3142/https/github.com/hadolint/hadolint/releases/download/v2.15.1/hadolint-linux-x86_64 | grep X-Cache
```

`X-Cache` is `MISS` the first time and `HIT` after that. It's `REVALIDATED` when the server confirmed the file hasn't changed, and `STALE` when the server couldn't be asked.
