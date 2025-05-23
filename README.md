
![Capivara Synct](doc/images/capivara-sync.jpg)

# Capivara Sync

[![Go](https://img.shields.io/badge/Go-1.16%2B-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Capivara Sync is a command-line tool designed to simplify file synchronization and backup operations. This project provides several commands to manage and manipulate file paths, backups, and restores efficiently.

## Commands

### 1. `backup`
The `backup` command allows users to create backups of their files. It ensures data safety by storing copies of files in a secure location.

### 2. `restore`
The `restore` command is used to restore files from a backup. It ensures that files are retrieved and placed in their original or specified locations.

### 3. `rsync`
The `Rsync` command synchronizes files between two directories. It ensures that both directories contain the same files, making it easy to keep data consistent across multiple locations.


## Sources

capivara-sync supports the following sources:
- localsource: Local file system
- sshsource: Remote file system over SSH
- webdavsource: WebDAV server


## Installation

Clone the repository and build the project using Go:

```bash
git clone https://github.com/yourusername/capivara-sync.git
cd capivara-sync
go build
```

## Usage

Run the application with the desired command:

```bash
./capivara-sync <command> [flags]
```

For example:

```bash
./capivara-sync backup --source /path/to/source --destination /path/to/destination
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
