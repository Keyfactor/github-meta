# instructions for large language agents

Background: This repository contains useful github-related content for the Keyfactor organization. It is public for a mixture of convenience and transparency.

## Rules

1. Follow good practices: Reusable, documented, etc. 
2. Do not reinvent the wheel; Where there is an appropriate library to do the work, use it.
3. Assume you are likely wrong and ask for clarification on any aspect which is not already explicitly described to you.

## Coding Standards
### Python

All Python scripts should be built around using `uv`. The most basic script is

```python
#!/usr/bin/env uv run 
#/// script
# requirements: []
#///

def main():
    pass

if __name__ == "__main__":
    main()
```

### HTML

When writing HTML, prefer using CSS over inline styles. Do not play div-golf. Do not nest divs inside divs without reason. 

### Go

When writing Go, 

* Follow the Go style guide from Google with the specific exception that parameter names should be multiple characters long, loop variables should be reasonably descriptive, and comments should be shorter.
* Look for libraries that solve a problem before writing things from first principles
* When using tools, use `go get -tool <package>` rather than `go get -u <package>` and then call that tool using `go tool <toolname>`.

### Make

Avoid using Make. Update where make is currently in use, but do not create new Makefiles unless explictly directed to. Use Just instead.

### Just

Include at  the beginning the following recipe:

```justfile
set windows-shell := ['pwsh', '-c']

# list recipies.
_default: 
    @just --list
```

Wherever you write shell scripts:

* if it is only UNIX applicable, write only a standard shell script. use `set -uexo pipefail` to strengthen scripts.
* if it is applicable to both UNIX and Windows, either write both Shell and Powershell scripts (if it is shorter than 10 lines or so) or write explicitly in Powershell (if it is longer or simpler to use Powershell)

### Shell

* Always use `set -uexo pipefail`
* Avoid bashishms. Use as POSIX compliant shell as you can. Where you must, you can use Debian's tricks to avoid bashisms (e.g. `if [ "x$FOO" -eq "x"]` to determine if an environment variable is set) 
* be EXTREMELY cautious. Do **NOT** build complicated pipe trains. 

### Powershell

Expect Powershell 7. Be sure to include any library calls you need. Prefer to use builtins first. Don't assume all Powershell is Windows. 