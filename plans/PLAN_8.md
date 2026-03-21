---
title: "PLAN_8: Configuration Structure Refactoring"
status: "completed"
---

# PLAN_8: Configuration Structure Refactoring

## Overview

The objective of this plan is to refactor the configuration structure to a more job-centric approach. The new structure introduces `jobs` which have identifiers, optional names, enable/disable flags, and priorities. Each job contains a list of `tasks` representing a source (`from`) and multiple destinations (`to`). This structure explicitly separates synchronization tasks, allows disabling parts of the configuration without deletion, and prioritizes execution regardless of YAML map order, while still preserving sequential execution within tasks and destinations.

### New Configuration Structure
```yaml
jobs:
  personal_photos: # <job_id> just name fo logging
    name: "Personal Photo Archive" # optional default: <job_id> for human
    enabled: true  # optional default: true
    priority: 10   # optional default: 10
    tasks:
      - from: "local:/home/user/photos"
        to:
          - path: "gd:photos/personal"
            args: []      # optional default: []
            resync: false # optional default: false
          - path: "yd:photos"
            args: []
        enabled: true # optional default: true
  work_documents:
    name: "Work Documents"
    priority: 20
    tasks:
      - from: "local:./work" # may be RELATIVE to workdir/execdir path
        to:
          - path: "gd:docs/work"
            args: ["--exclude", "*.tmp"]
      - from: "gd:docs/shared"
        to:
          - path: "yd:work/shared"
            args: []
            resync: true
```

### Advantages of the new structure:
1. **Explicit task separation** - каждая синхронизация имеет идентификатор и имя.
2. **Включение/отключение** - `enabled: false` без удаления конфигурации.
3. **Приоритеты** - `priority` упорядочивает выполнение независимо от порядка YAML.
4. **Несколько источников в задаче** - `tasks` массив поддерживает любую топологию.
5. **Простое обращение к задачам** - `syncerman sync personal_photos`.
6. **Sequential Execution with Order Preservation** - порядок сохраняется внутри каждой задачи через порядок в массивах `tasks` и `to`.

## Milestones

### Milestone 1: Update Configuration Models and Parsing - completed

**Goal**: Redesign the configuration structs in `internal/config` and implement robust parsing and validation for the new YAML structure.

**Key Requirements**:
  - Create new structs: `Config`, `Job`, `Task`, `Destination`.
  - Implement YAML unmarshaling supporting default values (e.g., `enabled: true`, `priority: 10` for jobs, defaulting `name` to `job_id`).
  - Implement configuration validation (checking required fields, valid remote formats).
  - Implement sorting of jobs based on the `priority` field to guarantee execution order.

**Context**:
  - Impacts `internal/config` package.
  - References `specs/PACKAGE_SYNC.md` regarding order preservation.
  - Needs to replace the current provider-to-path nested map parsing.

### Milestone 2: Refactor Sync Engine and Target Expansion - completed

**Goal**: Update the `internal/sync` engine to generate executable targets from the new configuration structure while strictly maintaining sequential execution invariants.

**Key Requirements**:
  - Refactor `ExpandTargets` to iterate over sorted jobs.
  - Preserve array order for `tasks` and `to` destinations inside each job.
  - Filter out disabled jobs and tasks (`enabled: false`).
  - Pass job metadata (`job_id`, `name`) into targets for improved structured logging.

**Context**:
  - Impacts `internal/sync` package.
  - Must respect `Sequential Execution with Order Preservation` for linear synchronization chains.

### Milestone 3: CLI Commands Integration and Filtering - completed

**Goal**: Update the CLI to accept job identifiers for targeted synchronization and adapt existing commands to the new structures.

**Key Requirements**:
  - Modify `cmd/sync.go` to support job targeting: `syncerman sync <job_id>`.
  - Ensure backward compatibility or clear error messages for old target formats (`provider:path`).
  - Update `cmd/check.go` to use the new validation logic and report clear, path-aware errors for the new structure.

**Context**:
  - Impacts `internal/cmd` package.
  - Makes user interaction more intuitive as requested: `syncerman sync personal_photos`.

### Milestone 4: Update Documentation and Examples - completed - completed

**Goal**: Align all project documentation and configuration examples with the new structure.

**Key Requirements**:
  - Update `syncerman.yaml.example` with the new structure.
  - Rewrite `README.md` and `guides/OVERALL.md` configuration sections.
  - Update any existing unit and integration tests to use the new configuration format.

**Context**:
  - Ensures the repository reflects the latest changes and keeps the user guide accurate.