# Syncerman - Overall Project Definition

## Overview

Syncerman is a CLI application for synchronizing targets (sources and destinations) based on rclone CLI. It provides bidirectional synchronization between local and remote storage providers with flexible configuration options.

## Purpose

Syncerman simplifies the complex task of maintaining synchronized files across multiple storage providers by:
- Automating bidirectional sync using rclone bisync
- Providing declarative YAML configuration for multiple sync targets
- Supporting dry-run mode for safe validation
- Handling first-run scenarios automatically
- Verifying rclone remotes and destinations before sync

## Architecture

### Core Components

1. **CLI Framework** - Cobra-based command-line interface
2. **Configuration System** - YAML-based configuration with validation
3. **Rclone Integration Layer** - Command execution and verification
4. **Sync Engine** - Sequential sync processing with error handling
5. **Logging System** - Structured logging with multiple levels

### Package Structure

```
syncerman/
├── main.go                 # Application entry point
├── internal/
│   ├── cmd/               # CLI command definitions and handlers
│   ├── config/            # Configuration loading, parsing, and validation
│   ├── rclone/            # Rclone command execution and verification
│   ├── sync/              # Core sync logic and orchestration
│   ├── logger/            # Structured logging system
│   └── errors/            # Custom error types and utilities
├── guides/                # Project documentation
│   ├── STYLE.md          # Go code style guidelines
│   ├── PLANING.md        # Autonomous coding agent workflow
│   └── OVERALL.md        # This file - overall project definition
└── plans/                 # Development plans and tasks
```

## Configuration

### Configuration Format

Syncerman uses YAML configuration files to define sync targets with the following structure:

```yaml
<SRC provider name>:
    "<path in SRC provider>": 
        -
            to: "<DST provider>:<path in DST provider>"
            args: []           # optional array of additional rclone arguments
            resync: true       # optional flag (default: false)
        -
            to: "<DST provider 2>:<path in DST provider 2>"
            args: []
            resync: true
```

### Configuration Schema

**Provider Name**: The name of a remote storage provider configured in rclone
- Must match a remote name from `rclone listremotes`
- Can be "local" or any rclone-configured remote (e.g., "gdrive", "ydisk", "s3")

**Source Path**: The path within the source provider
- Always relative to the provider's root
- For local provider, use relative path like `./cloud/docs`

**Destination Object** (required):
- `to` (string, required): Destination in format `<provider>:<path>` or `<path>` for local
- `args` ([]string, optional): Additional rclone command-line arguments
- `resync` (bool, optional): Use --resync flag (default: false)

### Configuration Example

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

**Explanation**:
- Sync local folder `./cloud/mirror/folder` to `gdrive:folders/folder1` and `ydisk:folders/folder1`
- Sync `gdrive:folders/folder1` to `ydisk:folders/folder1`
- All syncs use standard rclone bisync options (no custom args, no resync)

### Configuration File Discovery

Syncerman searches for configuration files in the following order:
1. Explicit path specified via `--config` flag
2. `.syncerman.yml` in current directory

## CLI Reference

### Global Flags

| Flag        | Short | Description                       | Default         |
| ----------- | ----- | --------------------------------- | --------------- |
| `--config`  | `-c`  | Path to configuration file        | auto-discovered |
| `--dry-run` | `-d`  | Perform trial run without changes | false           |
| `--verbose` | `-v`  | Enable verbose output             | false           |
| `--quiet`   | `-q`  | Suppress non-error output         | false           |

### Commands

#### `sync [flags]` - Sync All Targets

Synchronize all targets defined in the configuration file.

**Usage:**
```bash
syncerman sync
syncerman sync --verbose
syncerman sync --dry-run
```

**Options:**
- Inherits all global flags

**Behavior:**
1. Validates configuration file
2. Verifies all rclone remotes are configured
3. Creates destination directories if needed
4. Sequentially runs rclone bisync for each target
5. Handles first-run errors automatically

#### `sync <provider>:<path> [flags]` - Sync Specific Target

Synchronize a specific provider and path.

**Usage:**
```bash
syncerman sync local:./cloud/docs
syncerman sync gdrive:folders/folder1 --verbose
syncerman sync ydisk:folders/folder1 --dry-run
```

**Options:**
- Inherits all global flags

**Target Format:**
- For local: `local:./path/to/folder` or `./path/to/folder`
- For remotes: `<provider>:<path>`

#### `check [flags]` - Check Configuration and Remotes

Validate the configuration file and verify all rclone remotes.

**Usage:**
```bash
syncerman check
syncerman check --config /path/to/config.yml
syncerman check --verbose
```

**Options:**
- Inherits all global flags

**Validates:**
- YAML syntax
- Configuration structure
- Provider names not empty
- Source paths not empty
- Destination format correct
- Optional field types correct
- All provider names exist in rclone configuration
- Rclone binary is accessible
- Connection to each remote is possible

### CLI Examples

**Scenario 1: First-time setup and validation**

```bash
# 1. Check your configuration and verify rclone remotes
syncerman check --verbose

# 2. Dry-run to see what would happen
syncerman sync --dry-run --verbose
```

**Scenario 2: Regular sync all targets**

```bash
# Sync all targets
syncerman sync --verbose

# Sync with quiet mode (only errors)
syncerman sync --quiet
```

**Scenario 3: Sync specific folder**

```bash
# Sync only local documents folder
syncerman sync local:./documents --verbose

# Dry-run specific folder
syncerman sync gdrive:projects/main --dry-run
```

**Scenario 4: Using custom config file**

```bash
# Use specific config file
syncerman --config /home/user/.config/syncerman/config.yml sync

# Check with custom file
syncerman --config /home/user/.config/syncerman/config.yml check
```

## Rclone Integration

### Rclone Bisync Command Template

Syncerman executes rclone bisync with the following standard command:

```bash
rclone bisync <SRC Provider:><SRC Path> <DST Provider:><DST Path> \
  --create-empty-src-dirs \
  --compare size,modtime \
  --no-slow-hash \
  -MvP \
  --drive-skip-gdocs \
  --fix-case \
  --ignore-listing-checksum \
  --fast-list \
  --transfers=10 --resilient ${@}
```

### Rclone Options Explained

| Option                      | Purpose                                          |
| --------------------------- | ------------------------------------------------ |
| `--create-empty-src-dirs`   | Sync creation and deletion of empty directories  |
| `--compare size,modtime`    | Compare files by size and modification time      |
| `--no-slow-hash`            | Skip slow checksum calculations during listing   |
| `-MvP`                      | Preserve metadata, verbose output, show progress |
| `--drive-skip-gdocs`        | Skip Google Docs files (Google Drive specific)   |
| `--fix-case`                | Force rename of case-insensitive destinations    |
| `--ignore-listing-checksum` | Don't use checksums for listings                 |
| `--fast-list`               | Use faster directory listing                     |
| `--transfers=10`            | Run 10 parallel transfers                        |
| `--resilient`               | Allow recovering from errors without full resync |
| `${@}`                      | Additional user-specified arguments              |

### Provider Handling

**Local Provider:**
- Path format: `./path/to/folder` or `local:./path/to/folder`
- No provider prefix needed for local filesystem
- Paths started with `./` are relative to current working directory
- Absolute paths are absolute

**Remote Providers:**
- Format: `<provider>:<path>`
- Provider name must match rclone remote configuration
- Examples: `gdrive:documents`, `ydisk:backup`, `s3:bucket/folder`

### Directory Creation

Syncerman automatically creates destination directories with:

```bash
rclone mkdir <remote_name>:<path/to/new/directory>
```

This is executed for each sync target before running bisync to ensure:
- Destination paths exist
- Permission errors are caught early
- Sync operations can proceed without manual intervention

### First-Run Handling

Syncerman detects and handles the first-run scenario:

**Error Pattern:**
```
/Bisync critical error: cannot find prior Path1 or Path2 listings/ AND /here are the filenames we were looking for/ AND /Do they exist/
```

example (USE THIS IN TEST):
```
2026/03/14 20:14:03 ERROR : Bisync critical error: cannot find prior Path1 or Path2 listings, likely due to critical error on prior run 
Tip: here are the filenames we were looking for. Do they exist? 
Path1: /home/jerry/.cache/rclone/bisync/tmp_tools..kinnalru@yandex.ru_tools.path1.lst
Path2: /home/jerry/.cache/rclone/bisync/tmp_tools..kinnalru@yandex.ru_tools.path2.lst
```

**Syncerman's Response:**
1. Detects the error pattern in rclone output by REGEXP
2. Automatically re-runs the sync with `--resync` flag
3. Logs the resync operation for user awareness
4. Continues with subsequent sync targets

**User Control:**
- Users can also explicitly set `resync: true` in configuration
- Force initial sync to prefer source or destination content

## Workflow

### End-to-End Workflow

1. **Initial Setup**
    ```
    a. Create configuration.yml in desired location: manual
    b. Configure rclone remotes: rclone config
    c. Validate configuration and verify remotes: syncerman check
    ```

2. **Pre-Sync Verification**
   ```
   a. Load and parse configuration file
   b. Validate configuration structure
   c. Execute rclone listremotes to verify providers
   d. Execute rclone mkdir for each destination path
   ```

3. **Sync Execution**
   ```
   a. Iterate through all defined sync targets
   b. For each target:
      - Build rclone bisync command with arguments
      - Execute command and capture output
      - Monitor for first-run errors
      - Handle errors (retry with --resync if needed)
   c. Log all operations and results
   ```

4. **Post-Sync Operations**
   ```
   a. Report sync statistics (files transferred, errors)
   b. Log any warnings or issues encountered
   c. Return appropriate exit code
   ```

### Sequential Processing

Syncerman processes sync targets sequentially, not in parallel, to:
- Ensure each sync completes successfully before starting the next
- Maintain clear error tracking and reporting
- Avoid overwhelming system resources
- Simplify troubleshooting

### Error Handling Strategy

**Configuration Errors:**
- Immediate termination with clear error message
- File location and specific validation failure reported
- Suggested fixes provided when possible

**Rclone Verification Errors:**
- Stop sync operation
- Report missing or misconfigured remotes
- Suggest running `rclone config`

**Directory Creation Errors:**
- Log specific remote and path that failed
- Provide permissions guidance
- Continue with next target (if multiple)

**Sync Errors:**
- Log error with full rclone output
- Attempt resync for first-run errors
- Continue with next target (if multiple)
- Return non-zero exit code if any errors occurred

## Security Considerations

**Configuration is READ-ONLY**
- Syncerman never modifies configuration files
- Configurations are only validated and read
- Prevents accidental configuration corruption

**Path Handling:**
- Uses proper shell escaping for all paths
- Validates paths to prevent injection attacks
- Sanitizes user input before passing to rclone

**Sensitive Data:**
- Never logs passwords, API keys, or tokens
- Redacts sensitive information from output
- Follows rclone's security practices

## Dependencies

### Required Tools

- **rclone**: Binary for all sync operations (must be in PATH)

## Exit Codes

| Code | Meaning                                       |
| ---- | --------------------------------------------- |
| 0    | Success - all operations completed            |
| 1    | General error - operation failed              |
| 2    | Configuration error - invalid configuration   |
| 3    | Rclone error - rclone command failed          |
| 4    | Validation error - remote verification failed |
| 5    | File not found - configuration file missing   |


## Future Enhancements

**Planned Features:**
- Parallel sync processing with configurable concurrency
- Email/notifications on sync completion
- Exclude/include pattern support per target

