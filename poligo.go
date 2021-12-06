package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var segmentFuncsWithoutParams = map[string]func(seg *Segment){
	"cwd-exists":     cwdExists,
	"shell-level":    shellLevel,
	"term-title":     setTermTitle,
	"sudo-root":      sudoRoot,
	"python-version": pythonVersion,
	"go-version":     goVersion,
	"nodejs-project": nodejsProject,
	"work-dir":       workDirFull,
	"read-only":      readOnly,
	"kernel-version": kernelVersion,
	"git":            git,
	"warn-offline":   warnOffline,
	"virtual-env":    virtualEnv,
	"docker-version": dockerVersion,
	"current-time":   currentTime,
	"ssh-connection": sshConnection,
	"user-name":      userNameWithoutDefault,
}

var segmentFuncsWithParam = map[string]func(seg *Segment, param string){
	"exit-code":   exitCode,
	"work-dir":    workDirPartial,
	"warn-memory": warnMemory,
	"user-name":   userNameWithDefault,
}

func main() {

	argTimeout := flag.Duration("timeout", 1*time.Second, "total execution timeout")

	flag.Usage = func() {
		w := flag.CommandLine.Output()

		fmt.Fprintf(w, "Usage of %s: \n", os.Args[0])

		flag.PrintDefaults()

		fmt.Fprintf(w, "  Segments: \n")
		fmt.Fprintf(w, "    cwd-exists: Check current working directory exists\n")
		fmt.Fprintf(w, "    warn-memory=N%%: Warn if more than N%% of memory/swap are used\n")
		fmt.Fprintf(w, "    term-title: Set terimal title\n")
		fmt.Fprintf(w, "    current-time: Display current time\n")
		fmt.Fprintf(w, "    go-version: Display installed go version, if any *.go files in current directory\n")
		fmt.Fprintf(w, "    python-version: Display installed python version, if any *.py files in current directory\n")
		fmt.Fprintf(w, "    nodejs-project: Display project title and version, if current directory contains package.json\n")
		fmt.Fprintf(w, "    docker-version: Display installed docker version, if current directory contains a Dockerfile\n")
		fmt.Fprintf(w, "    kernel-version: Display linux kernel version in /, /boot and /usr/src\n")
		fmt.Fprintf(w, "    warn-offline: Warn if no network connection available\n")
		fmt.Fprintf(w, "    shell-level: Display number of nested shells\n")
		fmt.Fprintf(w, "    virtual-env: Notify about activated python virtual environment\n")
		fmt.Fprintf(w, "    work-dir=N: Show current working directory, optional limit the output to N folders\n")
		fmt.Fprintf(w, "    sudo-root: Warn if current terminal has root permissions via sudo\n")
		fmt.Fprintf(w, "    git: Show git status, pushs, pulls, modified files and current branch\n")
		fmt.Fprintf(w, "    read-only: Warn if current directory is read only\n")
		fmt.Fprintf(w, "    ssh-connection: Warn if terminal is connected via ssh\n")
		fmt.Fprintf(w, "    user-name=DEFAULT: Show username except when it equals DEFAULT\n")
		fmt.Fprintf(w, "    exit-code=$?: Show if last command returned an error code (parameter must be $?)\n")
	}

	flag.Parse()
	argSegments := flag.Args()
	if len(argSegments) == 0 {
		flag.Usage()
	} else {
		execute(&argSegments, argTimeout)
	}

}
