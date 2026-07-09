# gitGo

A zero-dependency Git command line tool written in Go, built on top of [go-git](https://github.com/go-git/go-git).

## Features

- **Zero Dependency**: No need to install Git, everything is bundled in a single binary
- **Cross Platform**: Supports Windows, Linux (amd64, arm64, mipsle), macOS
- **Lightweight**: Static compilation, small binary size
- **Goubi Router 2 Support**: Optimized for MIPS architecture

## Installation

### Download Prebuilt Binaries

Download from the [GitHub Releases](https://github.com/YellCatt/gitGo/releases) page.

### Build from Source

```bash
# Clone the repository
git clone https://github.com/YellCatt/gitGo.git
cd gitGo

# Build for current platform
go build -o gitgo .

# Build for specific platforms
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gitgo_linux_amd64
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o gitgo_linux_arm64
GOOS=linux GOARCH=mipsle GOARM=softfloat CGO_ENABLED=0 go build -ldflags="-s -w -extldflags=-static" -tags="netgo osusergo" -o gitgo_linux_mipsle
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o gitgo_windows_amd64.exe
```

## Usage

```bash
gitGo - A zero-dependency Git command line tool

Usage:
  gitGo init [path]          Initialize a new git repository
  gitGo clone <url> [path]   Clone a remote repository
  gitGo pull                 Pull changes from remote
  gitGo status               Show working tree status
  gitGo add [files...]       Add files to staging area
  gitGo commit -m <message>  Commit changes
  gitGo log                  Show commit history
  gitGo remote add <name> <url>  Add remote repository
```

### Commands

#### init
Initialize a new Git repository:
```bash
gitgo init
gitgo init myproject
```

#### clone
Clone a remote repository:
```bash
# Basic clone
gitgo clone https://github.com/YellCatt/gitGo.git

# Clone with specific branch
gitgo clone -b main https://github.com/YellCatt/gitGo.git

# Shallow clone (faster, only latest commit)
gitgo clone --depth 1 https://github.com/YellCatt/gitGo.git

# Force overwrite existing directory
gitgo clone -f https://github.com/YellCatt/gitGo.git

# Verbose mode for debugging
gitgo -v clone https://github.com/YellCatt/gitGo.git
```

#### pull
Pull changes from remote repository:
```bash
gitgo pull
```

#### status
Show working tree status:
```bash
gitgo status
```

#### add
Add files to staging area:
```bash
# Add specific files
gitgo add file.txt

# Add all files
gitgo add .
```

#### commit
Commit changes:
```bash
gitgo commit -m "Initial commit"
```

#### log
Show commit history:
```bash
gitgo log
```

#### remote
Add remote repository:
```bash
gitgo remote add origin https://github.com/YellCatt/gitGo.git
```

## Options

| Option | Description |
|--------|-------------|
| `-v`, `--verbose` | Enable verbose mode for debugging |
| `-f`, `--force` | Force overwrite (used with clone) |
| `-b <branch>` | Specify branch (used with clone) |
| `--depth <n>` | Create shallow clone with n commits |

## Supported Platforms

| Platform | Architecture | Binary |
|----------|--------------|--------|
| Linux | amd64 | `gitgo_linux_amd64` |
| Linux | arm64 | `gitgo_linux_arm64` |
| Linux | mipsle | `gitgo_linux_mipsle` (Goubi Router 2) |
| Windows | amd64 | `gitgo_windows_amd64.exe` |

## Notes

1. **Gitee Compatibility**: Some Git servers (like Gitee) may have protocol incompatibility with go-git library. If cloning fails, try using `--depth 1` for shallow clone or use the standard git command.

2. **Credentials**: For repositories requiring authentication, you can include credentials in the URL:
   ```bash
   gitgo clone https://username:password@github.com/user/repo.git
   ```

3. **Goubi Router 2**: The MIPS binary is optimized for Goubi Router 2 with softfloat support and static compilation.

## License

MIT License

## Contributing

Feel free to submit issues and pull requests!