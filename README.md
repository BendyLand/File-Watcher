# File Watcher

A basic file watcher written in Go!

## About

This tool was originally created for my build tool `blmake`, to watch for when specifically C/C++ files change as a way to implement incremental recompilation.

However, I ended up liking the project more than I thought I would, so I decided to make a standalone version that works on all file types and traverses recursively.

## Usage
```bash
./file-watcher <directory_path>|<command>
```
**Valid commands:**
 - `help`  - Shows this menu.
 - `init`  - Generate the necessary directory structure for the tool.
 - `clear` - Clears 'prev.json' in case it gets corrupted.
*(Running the tool again will repopulate it.)*

### Quick Start

```bash 
./file-watcher init
./file-watcher . # or any other directory you want to watch
# Subsequent runs will detect changes to those files
```

