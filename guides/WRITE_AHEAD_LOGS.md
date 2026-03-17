# WAL (Write Ahead Log) definition and rules

- WAL persisted ONLY in file `./WAL.md`
- use strict format of `./WAL.md` file: `guides/templates/WAL_TEMPLATE.md`
- see example of WAL format: `guides/templates/WAL_EXAMPLE.md`
- DO NOT create `ROJECT COMPLETION SUMMARY` or something like in `./WAL.md`
- DO NOT create Results, Summary, or other Final Statistic - ONLY `Plan` with status and milestone names
- DO NOT analyze old plans and milestones
- you can read Markdown Frontmatters from previous plans if you REALLY need it
- order PLANS in `## Work Log` by plan number from top to bottom


## Archive or Clean

Synonims: `CLEAN WAL`, `COMPACTIFY WAL`, `ARCHIVE WAL`

Archiving (cleaning) of `WAL.md` file may be requested:
  - by user MANUALY 
  - by flow (ex. `Create New Milestone Flow`) after all plans completition

While cleaning or compactification file `./WAL.md` you need:
- remove details from completed milestones leave only names and status
- remove details from completed Plans leave only names and status
- CRITICAL format `./WAL.md` by rules defined in `WAL (Write Ahead Log) definition and rules`
