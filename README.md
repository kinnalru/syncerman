# Syncerman - CLI for bidirectional file synchronization using rclone

## Features

- Automating bidirectional sync using rclone bisync
- Providing declarative YAML configuration for multiple sync targets
- Supporting dry-run mode for safe validation
- Handling first-run scenarios automatically
- Verifying rclone remotes and destinations before sync

## Installation

### Prerequisites

- **Go 1.21+** - Required for building Syncerman from source
- **rclone CLI** - Required at runtime for all synchronization operations

### Building from Source

#### Using the Makefile (Recommended)

The simplest way to build Syncerman is to use the provided Makefile. This will compile binaries for both Linux and Windows platforms:

```bash
make build
```

The compiled binaries will be placed in the `bin/` directory:
- `bin/syncerman-linux-amd64` - Linux executable
- `bin/syncerman-windows-amd64` - Windows executable

#### Manual Build Commands

If you prefer to build manually or need to build for a specific platform only:

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/syncerman-linux-amd64
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/syncerman-windows-amd64
```

### Installing rclone

rclone is essential for Syncerman to function. Install it following the official installation guide: https://rclone.org/install/

**Quick installation on Linux:**

```bash
curl https://rclone.org/install.sh | sudo bash
```

**Quick installation on macOS:**

```bash
brew install rclone
```

**Quick installation on Windows:**

Download the installer from https://rclone.org/downloads/ and follow the installation wizard.

**Important:** After installation, ensure that `rclone` is available in your system's PATH. You can verify this by running:

```bash
rclone version
```

### Post-Installation Setup

After building and installing the required dependencies, verify your setup:

1. **Verify Syncerman binary:**

```bash
./bin/syncerman-linux-amd64 --help
# Or on Windows:
# .\bin\syncerman-windows-amd64.exe --help
```

2. **Verify rclone availability:**

```bash
rclone version
```

Both commands should display their respective version information, confirming a successful installation.

## Quick Start

This guide will help you set up and run your first synchronization with Syncerman.

### Step 1: Configure Rclone Remotes

First, set up your storage providers using rclone:

```bash
rclone config
```

Follow the interactive prompts to configure your storage providers. Common providers include:
- **Google Drive** (remote name: `gdrive`)
- **Yandex Disk** (remote name: `ydisk`)
- **Amazon S3** (remote name: `s3`)
- **Dropbox**, **OneDrive**, and many others

You can list your configured remotes at any time:

```bash
rclone listremotes
```

### Step 2: Create Configuration File

Create a `.syncerman.yml` file in your project directory. Syncerman uses YAML format to define sync targets.

**Example configuration:**

```yaml
local:
    "./cloud/mirror/folder":
        -
            to: gdrive:folders/folder1
            args: []
            resync: false
        -
            to: ydisk:folders/folder1
            args: []
            resync: false

gdrive:
    "folders/folder1":
        -
            to: ydisk:folders/folder1
            args: []
            resync: false
```

**Explanation:**
- The first section syncs local folder `./cloud/mirror/folder` to both `gdrive:folders/folder1` and `ydisk:folders/folder1`
- The second section syncs `gdrive:folders/folder1` directly to `ydisk:folders/folder1`
- `args` allows additional rclone arguments (optional)
- `resync` forces initial sync to prefer source content (default: false)

### Step 3: Validate Configuration and Remotes

Before running any sync, validate your configuration file and verify rclone remotes:

```bash
syncerman check --verbose
```

This command checks for:
- YAML syntax errors
- Configuration structure validity
- Non-empty provider names and source paths
- Correct destination format
- Proper field types
- All provider names exist in rclone configuration
- Rclone binary is accessible
- Connection to each remote is possible

**Expected output:** All checks passed or detailed error messages if issues are found.

### Step 5: Dry Run First Sync

Preview what will happen during synchronization without making any actual changes:

```bash
syncerman sync --dry-run --verbose
```

This dry run shows:
- Which files would be synced
- Direction of synchronization
- Any potential conflicts or issues

**Important:** Always review dry-run output before executing actual syncs, especially for first use.

### Step 6: Run First Sync

After confirming the dry run output looks correct, execute the actual synchronization:

```bash
syncerman sync --verbose
```

This will:
- Validate configuration and remotes (again)
- Create destination directories if needed
- Sequentially run rclone bisync for each target
- Handle first-run scenarios automatically
- Display detailed progress and results

**Expected output:** Detailed log showing sync progress, files transferred, and any warnings or errors.

## Configuration

### Configuration File Location

Syncerman searches for configuration files in the following order:
1. Explicit path specified via `--config` or `-c` flag
2. `.syncerman.yml` in current directory (default)

### Configuration Format Overview

Syncerman uses YAML configuration files to define sync targets. The configuration is structured by provider names, with each provider containing one or more source paths that define sync targets.

**Basic Structure:**
```yaml
<provider_name>:
    "<source_path>":
        -
            to: "<destination>"
            args: []
            resync: false
```

### Example 1: Basic Single-Target Sync

Sync a local folder to a Google Drive folder:

```yaml
local:
    "./documents":
        -
            to: gdrive:backup/documents
            args: []
            resync: false
```

### Example 2: Multi-Target Sync

Sync one source to multiple destinations and sync between remotes:

```yaml
local:
    "./cloud/mirror/folder":
        -
            to: gdrive:folders/folder1
            args: []
            resync: false
        -
            to: ydisk:folders/folder1
            args: []
            resync: false

gdrive:
    "folders/folder1":
        -
            to: ydisk:folders/folder1
            args: []
            resync: false
```

This configuration:
- Syncs `./cloud/mirror/folder` to both `gdrive:folders/folder1` and `ydisk:folders/folder1`
- Syncs between `gdrive:folders/folder1` and `ydisk:folders/folder1`

### Configuration Fields Explained

- **Provider Name** - The rclone remote name (e.g., `gdrive`, `ydisk`, `s3`, `local`). Must match a remote name from `rclone listremotes`.
- **Source Path** - The path within the source provider. Relative to provider's root. For `local` provider, use relative paths like `./folder`.
- **to** (required) - Destination in format `<provider>:<path>` or `<path>` for local filesystem.
- **args** (optional) - Array of additional rclone command-line arguments. Default is empty array.
- **resync** (optional) - Flag to force initial sync with `--resync` option. Default is `false`.

For detailed configuration information, see `guides/OVERALL.md:46-99`.

## CLI Commands

### Global Flags

| Flag        | Short | Description                       | Default         |
| ----------- | ----- | --------------------------------- | --------------- |
| `--config`  | `-c`  | Path to configuration file        | auto-discovered |
| `--dry-run` | `-d`  | Perform trial run without changes | false           |
| `--verbose` | `-v`  | Enable verbose output             | false           |
| `--quiet`   | `-q`  | Suppress non-error output         | false           |

### Commands

#### sync [flags]

Synchronize all targets defined in the configuration file.

**Usage:**
```bash
syncerman sync
syncerman sync --verbose
syncerman sync --dry-run
```

This command:
1. Validates configuration file
2. Verifies all rclone remotes are configured
3. Creates destination directories if needed
4. Sequentially runs rclone bisync for each target
5. Handles first-run errors automatically

#### sync <provider:path> [flags]

Synchronize a specific provider and path.

**Usage:**
```bash
syncerman sync local:./cloud/docs
syncerman sync gdrive:folders/folder1 --verbose
syncerman sync ydisk:folders/folder1 --dry-run
```

**Target Format:**
- For local: `local:./path/to/folder` or `./path/to/folder`
- For remotes: `<provider>:<path>`

#### check [flags]

Validate the configuration file and verify all rclone remotes.

**Usage:**
```bash
syncerman check
syncerman check --config /path/to/.syncerman.yml
syncerman check --verbose
```

This command validates:
- YAML syntax
- Configuration structure
- Provider names not empty
- Source paths not empty
- Destination format correct
- Optional field types correct
- All provider names exist in rclone configuration
- Rclone binary is accessible
- Connection to each remote is possible

## Usage Examples

### Scenario 1: First-time Setup and Validation

Setup Syncerman for the first time with proper validation:

```bash
# 1. Check your configuration and remotes
syncerman check --verbose

# 2. Dry-run to see what would happen
syncerman sync --dry-run --verbose

# 3. Run the actual sync
syncerman sync --verbose
```

### Scenario 2: Regular Sync All Targets

Sync all targets defined in your configuration on a regular basis:

```bash
# Sync all targets with verbose output
syncerman sync --verbose

# Sync with quiet mode (only errors)
syncerman sync --quiet

# Dry-run before actual sync (recommended for safety)
syncerman sync --dry-run --verbose
syncerman sync --verbose
```

### Scenario 3: Sync Specific Folder

Sync only a specific folder to save time or target a specific need:

```bash
# Sync only local documents folder
syncerman sync local:./documents --verbose

# Sync a specific Google Drive folder
syncerman sync gdrive:projects/main --verbose

# Dry-run specific folder first
syncerman sync gdrive:projects/main --dry-run
```

### Scenario 4: Using Custom Config File

Use a configuration file from a different location:

```bash
# Use specific config file from home directory
syncerman --config /home/user/.config/syncerman/.syncerman.yml sync

# Use custom config with verbose output
syncerman -c ./my-sync-config.yml sync --verbose

# Check with custom file
syncerman --config /home/user/.config/syncerman/.syncerman.yml check
```

### Scenario 5: Automated Backup Setup

Set up automated daily backups using cron:

```bash
# Add to crontab for daily sync at 2 AM
0 2 * * * /usr/local/bin/syncerman sync --quiet >> /var/log/syncerman.log 2>&1
```

### Scenario 6: Testing New Configuration

When modifying your configuration, test it safely:

```bash
# 1. Validate new configuration and remotes
syncerman check --verbose

# 2. Dry-run to preview changes
syncerman sync --dry-run --verbose

# 3. If dry-run looks good, run actual sync
syncerman sync --verbose
```

## Troubleshooting

### Configuration Errors

**Error: "Configuration file not found"**
- **Cause**: Syncerman cannot locate the `.syncerman.yml` file
- **Solution**: 
  - Ensure `.syncerman.yml` exists in your current directory
  - Or specify config path with `--config /path/to/.syncerman.yml`

**Error: "Invalid YAML syntax"**
- **Cause**: Malformed YAML in configuration file
- **Solution**:
  - Check YAML indentation (use spaces, not tabs)
  - Verify proper quote usage
  - Use an online YAML validator to check syntax

**Error: "Invalid configuration: provider name cannot be empty"**
- **Cause**: Empty or missing provider name
- **Solution**:
  - Ensure provider names are defined as root-level keys
  - Reference: `gdrive:`, not empty provider

**Error: "Invalid configuration: source path cannot be empty"**
- **Cause**: Empty or missing source path under provider
- **Solution**:
  - Ensure paths are quoted and non-empty
  - Example: `"./documents"` or `"folders/backup"`

### Rclone Remote Verification Failures

**Error: "Remote 'gdrive' not found in rclone configuration"**
- **Cause**: Provider name doesn't exist in rclone configuration
- **Solution**:
  - Run `rclone config` to add the remote
  - Run `rclone listremotes` to verify available remotes
  - Ensure provider name matches exactly in `.syncerman.yml`

**Error: "rclone binary not found"**
- **Cause**: rclone is not installed or not in PATH
- **Solution**:
  - Install rclone: https://rclone.org/install/
  - Verify installation: `rclone version`
  - Ensure rclone is in your system PATH

**Error: "Connection to remote failed"**
- **Cause**: Network issues, incorrect credentials, or remote server down
- **Solution**:
  - Check network connectivity
  - Verify rclone remote configuration with `rclone config show <remote>`
  - Test connection: `rclone lsd <remote>:`

### First-Run Sync Errors

**Error: "Bisync critical error: cannot find prior Path1 or Path2 listings"**
- **Cause**: First sync attempt with no prior state files
- **Solution**:
  - Syncerman automatically handles this by re-running with `--resync`
  - If manual intervention needed: Run with `resync: true` in config
  - This is normal for first sync - Syncerman will auto-resolve

**Error: "Permission denied while creating directory"**
- **Cause**: Insufficient permissions on remote storage
- **Solution**:
  - Check remote storage permissions
  - Verify account has write access
  - For local paths: ensure directory permissions allow write access

### General Sync Errors

**Sync is slow or hanging**
- **Cause**: Large file transfer, network bandwidth, or slow remote
- **Solution**:
  - Use `--verbose` flag to see detailed progress
  - Consider syncing during off-peak hours
  - Check rclone settings for optimization

**Some files are not syncing**
- **Cause**: Permission issues, file locks, or exclusions
- **Solution**:
  - Check file permissions on both source and destination
  - Close any applications using the files
  - Review rclone bisync output for specific error messages
  - Use `--verbose` flag for detailed diagnostics

### Getting Help

If you encounter issues not covered here:
- Use `--verbose` flag to get detailed output
- Enable `--dry-run` to preview changes without executing
- Check rclone documentation: https://rclone.org/docs/
- Review `guides/OVERALL.md` for detailed technical information
