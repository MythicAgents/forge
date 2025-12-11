# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.8] - 2025-12-11

### Changed

- Updated to latest Go with slightly breaking changes

## [0.0.7] - 2025-06-18

### Changed

- Merged in PR to update the download BOF functionality to leverage CustomURLs

## [0.0.6] - 2025-05-02

### Changed

- Updated forge_create to use extension.json command_name for bof creation

## [0.0.5] - 2025-05-02

### Changed

- Adjusted the forge_create logic ordering to respect local files first

## [0.0.4] - 2025-05-02

### Changed

- Updated the sliverarmory extension.json parsing to include the new `default` values 
  - https://github.com/BishopFox/sliver/issues/1910
  
## [0.0.3] - 2025-04-11

### Changed

- Removed a few sliverarmory sources that were reflective dlls and not BOFs
- Updated gomod versions

## [0.0.2] - 2025-03-12

### Changed

- Updated MythicContainer version to v1.4.19

## [0.0.1] - 2025-03-05

### Changed

- Added changelog
- Updated the newly created tasks to update the task status when execution passes to the agent's container
- Updated MythicContainer version to v1.4.18
