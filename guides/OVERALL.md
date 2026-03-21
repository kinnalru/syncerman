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
jobs:
  <job_id>:
    name: "Optional readable name"
    enabled: true        # optional flag (default: true)
    priority: 10         # optional sort order (default: 10)
    tasks:
      - from: "<SRC provider>:<path in SRC provider>"
        to:
          - path: "<DST provider>:<path in DST provider>"
            args: []           # optional array of additional rclone arguments
            resync: true       # optional flag (default: false)
          - path: "<DST provider 2>:<path in DST provider 2>"
            args: []
            resync: true
        enabled: true    # optional flag (default: true)
```

### Configuration Schema

**jobs**: Root map containing sync jobs.

**Job Object**:
- `<job_id>` (string, required): Key defining the job identifier. Used in CLI.
- `name` (string, optional): Human-readable name. Defaults to job_id.
- `enabled` (bool, optional): Skip this job if false. Defaults to true.
- `priority` (int, optional): Execution priority. Lower runs first. Defaults to 10.
- `tasks` (array, required): List of sync tasks.

**Task Object**:
- `from` (string, required): Source path in format `<provider>:<path>`.
  - Can be "local:" or any rclone-configured remote (e.g., "gdrive:", "ydisk:", "s3:").
  - Always relative to the provider's root. For local, use relative path like `local:./cloud/docs` or absolute `local:/home/user/docs`.
- `to` (array, required): List of destinations.
- `enabled` (bool, optional): Skip this task if false. Defaults to true.

**Destination Object**:
- `path` (string, required): Destination in format `<provider>:<path>`.
- `args` ([]string, optional): Additional rclone command-line arguments.
- `resync` (bool, optional): Use --resync flag (default: false).

### Configuration Example

```yaml
jobs:
  cloud_mirror:
    name: "Cloud Mirroring"
    tasks:
      - from: "local:./cloud/mirror/folder"
        to:
          - path: "gdrive:folders/folder1"
          - path: "ydisk:folders/folder1"
  inter_cloud:
    tasks:
      - from: "gdrive:folders/folder1"
        to:
          - path: "ydisk:folders/folder1"
```

**Explanation**:
- Defines two jobs: `cloud_mirror` and `inter_cloud`. Both run because `enabled` is implicitly true. Both have priority 10, so they run in alphabetical or defined order based on sorting rules.
- `cloud_mirror` task syncs local folder `./cloud/mirror/folder` to `gdrive:folders/folder1` and `ydisk:folders/folder1`.
- `inter_cloud` task syncs `gdrive:folders/folder1` to `ydisk:folders/folder1`.
- All syncs use standard rclone bisync options (no custom args, no resync).

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
4. Sequentially runs rclone bisync for each target in YAML configuration order
5. Handles first-run errors automatically

#### `sync <job_id> [flags]` - Sync Specific Target

Synchronize a specific job defined in the configuration.

**Usage:**
```bash
syncerman sync cloud_mirror
syncerman sync document_backup --verbose
syncerman sync inter_cloud --dry-run
```

**Options:**
- Inherits all global flags

**Target Format:**
- Uses the `<job_id>` from the configuration file's `jobs` map.

#### `check [flags]` - Check Configuration and Remotes

Validate the configuration file and verify all rclone remotes.

**Usage:**
```bash
syncerman check
syncerman check --config /path/to/.syncerman.yml
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

**Scenario 3: Sync specific job**

```bash
# Sync only document backup job
syncerman sync document_backup --verbose

# Dry-run specific job
syncerman sync cloud_mirror --dry-run
```

**Scenario 4: Using custom config file**

```bash
# Use specific config file
syncerman --config /home/user/.config/syncerman/.syncerman.yml sync

# Check with custom file
syncerman --config /home/user/.config/syncerman/.syncerman.yml check
```

## Rclone Integration

### Rclone Bisync Command Template

Syncerman executes rclone bisync with the following standard command:

```bash
rclone bisync <SRC Provider>:<SRC Path> <DST Provider>:<DST Path> \
  --create-empty-src-dirs \
  --compare size,modtime \
  --no-slow-hash \
  -Mv \
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
| `-Mv`                       | Preserve metadata, verbose output                |
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

Example (USE THIS IN TEST):
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
    a. Create .syncerman.yml in current directory: manual
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

### Configuration Order Preservation

Syncerman preserves the execution order of sync targets through job priorities and array sequences, which is critical for linear synchronization chains.

**Implementation Details:**
- Jobs are sorted and executed ascending by their `priority` value (default 10).
- Within a job, `tasks` are executed in the exact sequence they appear in the YAML array.
- Within a task, multiple `to` destinations are executed in their exact YAML array sequence.
- Order is maintained throughout the entire sync pipeline (configuration loading → target expansion → execution)

**Why Order Matters:**
Linear synchronization chains require precise execution order to propagate files correctly through multiple destinations:
```
Configuration:
jobs:
  chain_sync:
    tasks:
      - from: 'local:/data'              # Target 1: local → gd
        to:
          - path: 'gd:syncerman/data/'
      - from: 'gd:syncerman/data/'       # Target 2: gd → yd
        to:
          - path: 'yd:syncerman/data/'
      - from: 'yd:syncerman/data/'       # Target 3: yd → local2
        to:
          - path: 'local:/backup/yd_backup/'

Execution order (preserved by task array):
  1. local:/data → gd:syncerman/data/   (Initial sync)
  2. gd:syncerman/data/ → yd:syncerman/data/   (Files from gd)
  3. yd:syncerman/data/ → local:/backup/yd_backup/   (Files from yd)
```

Without order preservation, targets would execute randomly, causing:
- Empty destinations in the chain
- Files not propagating through the entire chain
- State file corruption errors on subsequent syncs

**Order-Guaranteed Methods:**
- `ExpandTargets()` returns targets sorted by job priority, then by task order, then by destination order
- `RunAll()` executes targets in the exact order returned by ExpandTargets()

All configuration examples in this documentation assume order preservation for correct behavior.

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

