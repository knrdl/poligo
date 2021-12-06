package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func cwdExists(seg *Segment) {
	path := os.Getenv("PWD")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		parts := strings.Split(path, "/")
		for len(parts) > 0 {
			parts = parts[:len(parts)-1]
			path = strings.Join(parts, "/")
			_, err := os.Stat(path)
			if err == nil {
				break
			}
		}

		seg.AddHeadline("Your current directory is invalid.")
		seg.AddHeadline("Highest valid directory: ", path)
	}
}

func setTermTitle(seg *Segment) {
	seg.AddText("\\[\\e]0;\\u@\\h: \\w\\a\\]")
}

func shellLevel(seg *Segment) {
	lvl := os.Getenv("SHLVL")
	i, err := strconv.Atoi(lvl)
	if err == nil && i > 0 {
		seg.SetTextColor(colorShellLevel)
		seg.AddText("⏫", lvl)
	}
}

func sudoRoot(seg *Segment) {
	if _, err := exec.Command("sudo", "-n", "echo").Output(); err == nil {
		seg.SetTextColor(ColorSudoRoot)
		seg.AddText("🔑")
	}
}

func exitCode(seg *Segment, param string) {
	if param != "0" {
		seg.SetTextColor(ColorCmdFailed)
		seg.AddText("⚑", " ", param)
	}
}

func pythonVersion(seg *Segment) {
	if match, err := filepath.Glob("*.py"); err == nil && match != nil && os.Getenv("VIRTUAL_ENV") == "" {
		if out, err := exec.Command("python3", "--version").Output(); err == nil {
			if split := strings.Split(string(out), " "); len(split) >= 2 {
				seg.SetTextColor(ColorPython)
				seg.AddText("py", strings.TrimSuffix(split[1], "\n"))
			}
		}
	}
}

func goVersion(seg *Segment) {
	if match, err := filepath.Glob("*.go"); err == nil && match != nil {
		if out, err := exec.Command("go", "version").Output(); err == nil {
			if split := strings.Split(string(out), " "); len(split) >= 3 {
				seg.SetTextColor(ColorGolang)
				seg.AddText(split[2])
			}
		}
	}
}

func nodejsProject(seg *Segment) {
	type packageJSON struct {
		Version string `json:"version"`
		Name    string `json:"name"`
	}

	parts := strings.Split(os.Getenv("PWD"), "/")
	for i := 0; i < len(parts); i++ {
		path := "/" + filepath.Join(append(parts[:(len(parts)-i)], "package.json")...)
		if stat, err := os.Stat(path); err == nil && stat.Mode().IsRegular() {
			pkg := packageJSON{"!", "?"}
			if raw, err := ioutil.ReadFile(path); err == nil {
				if err := json.Unmarshal(raw, &pkg); err == nil {
					seg.SetTextColor(ColorNodeJs)
					seg.AddText("\u2B22", " ", pkg.Name, " ", pkg.Version)
				}
				break
			}
		}
	}
}

func workDirFull(seg *Segment) {
	workDirPartial(seg, "")
}

func workDirPartial(seg *Segment, maxDepthParam string) {
	maxDepth, err := strconv.Atoi(maxDepthParam)
	if err != nil {
		maxDepth = 0
	}
	const IconHome = "🏠"

	thinSep := func() {
		seg.AddText(" ")
		seg.SetTextColor(ColorDirSep)
		seg.AddText(IconSegmentSepThin)
	}

	cwd := os.Getenv("PWD")
	home := os.Getenv("HOME")
	if strings.HasPrefix(cwd, home) {
		cwd = IconHome + cwd[len(home):]
	}

	names := strings.Split(cwd, "/")
	if names[0] == "" {
		names = names[1:]
	}
	if names[0] == "" {
		names = []string{"/"}
	}

	nBefore := maxDepth - 1
	if maxDepth > 2 {
		nBefore = 2
	}

	for index, name := range names {
		if maxDepth == -1 || (index < nBefore || index > len(names)-maxDepth) {
			isHomeDir := name == IconHome
			isLastDir := index == len(names)-1
			if isHomeDir {
				seg.SetTextColor(ColorHomeDir)
			} else if isLastDir {
				seg.SetTextColor(ColorLastDir)
			} else {
				seg.SetTextColor(ColorPathDir)
			}
			if index > 0 {
				seg.AddText(" ")
			}
			seg.AddText(name)
			if name == IconHome && len(names) > 1 {
				seg.AddText(" ")
				seg.SetTextColor(SegmentColor{fg: ColorHomeDir.bg, bg: ColorPathDir.bg})
				seg.AddText(IconSegmentSep)
			} else if index < len(names)-1 {
				thinSep()
			}
		} else if maxDepth > -1 && index == nBefore {
			seg.AddText(" ", "\u2026") // shorten by ellipsis
			thinSep()
		}
	}
}

func readOnly(seg *Segment) {
	if unix.Access(os.Getenv("PWD"), unix.W_OK) != nil {
		seg.SetTextColor(ColorReadonlyDir)
		seg.AddText("🔒")
	}
}

func warnMemory(seg *Segment, maxPercentStr string) {
	maxPercent, err := strconv.Atoi(strings.TrimSuffix(maxPercentStr, "%"))
	if err != nil {
		panic(err)
	}
	reMemTotal := regexp.MustCompile("MemTotal:\\s+(\\d+)\\s+kB")
	reMemAvail := regexp.MustCompile("MemAvailable:\\s+(\\d+)\\s+kB")
	reSwpTotal := regexp.MustCompile("SwapTotal:\\s+(\\d+)\\s+kB")
	reSwpFree := regexp.MustCompile("SwapFree:\\s+(\\d+)\\s+kB")
	memTotal, memAvail, swpTotal, swpFree := -1, -1, -1, -1

	if stat, err := os.Stat("/proc/meminfo"); err == nil && stat.Mode().IsRegular() {
		file, err := os.Open("/proc/meminfo")
		if err == nil {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				line := scanner.Text()
				switch true {
				case reMemTotal.MatchString(line):
					if val, err := strconv.Atoi(reMemTotal.FindStringSubmatch(line)[1]); err == nil {
						memTotal = val
					}
				case reMemAvail.MatchString(line):
					if val, err := strconv.Atoi(reMemAvail.FindStringSubmatch(line)[1]); err == nil {
						memAvail = val
					}
				case reSwpTotal.MatchString(line):
					if val, err := strconv.Atoi(reSwpTotal.FindStringSubmatch(line)[1]); err == nil {
						swpTotal = val
					}
				case reSwpFree.MatchString(line):
					if val, err := strconv.Atoi(reSwpFree.FindStringSubmatch(line)[1]); err == nil {
						swpFree = val
					}
				}
			}
		}
	}
	ramPerc := 0
	swpPerc := 0
	if memTotal != -1 && memAvail != -1 {
		ramPerc = 100 - (memAvail * 100 / memTotal)
	}
	if swpFree != -1 && swpTotal != -1 {
		swpPerc = 100 - (swpFree * 100 / swpTotal)
	}
	if ramPerc >= maxPercent || swpPerc >= maxPercent {
		txt := "High memory usage: " + fmt.Sprintf("%d%% RAM", ramPerc)
		if swpPerc != 0 {
			txt += fmt.Sprintf(", %d%% Swap", swpPerc)
		}
		seg.AddHeadline(txt)
	}
}

func kernelVersion(seg *Segment) {
	path := os.Getenv("PWD")
	if path == "/" || strings.HasPrefix(path, "/boot") || strings.HasPrefix(path, "/usr/src") {
		buf := &unix.Utsname{}
		err := unix.Uname(buf)
		if err == nil {
			seg.SetTextColor(ColorKernel)
			seg.AddText("🐧 ", fmt.Sprintf("%s", bytes.Trim(buf.Release[:], "\x00")))
		}
	}
}

func git(seg *Segment) {
	const IconGit = "\uE0A0"

	isGitDir := func() bool {
		parts := strings.Split(os.Getenv("PWD"), "/")
		for i := 0; i < len(parts); i++ {
			if stat, err := os.Stat("/" + filepath.Join(append(parts[:(len(parts)-i)], ".git")...)); err == nil && stat.IsDir() {
				return true
			}
		}
		return false
	}

	if isGitDir() {
		out, err := exec.Command("git", "status", "--porcelain", "--branch", "--untracked-files=all").Output()
		if err == nil {
			txt := string(out)
			firstLine := strings.Split(txt, "\n")[0]
			branch := strings.TrimPrefix(firstLine, "## ")
			branch = strings.Split(branch, "...")[0]

			ahead, behind := 0, 0
			arr := strings.Split(firstLine, " [")
			if len(arr) > 1 {
				if m := regexp.MustCompile(`ahead (\d+)`).FindStringSubmatch(arr[1]); len(m) > 0 {
					ahead, _ = strconv.Atoi(m[1])
				}
				if n := regexp.MustCompile(`behind (\d+)`).FindStringSubmatch(arr[1]); len(n) > 0 {
					behind, _ = strconv.Atoi(n[1])
				}
			}

			changes := strings.Count(txt, "\n") - 1
			if changes == 0 {
				seg.SetTextColor(ColorGitClean)
				seg.AddText(IconGit, " ", branch)
			} else {
				seg.SetTextColor(ColorGitDirty)
				seg.AddText(IconGit, " ", branch, " ")
				if changes <= 20 {
					seg.AddText(fmt.Sprintf("%c", 9331+changes))
				} else {
					seg.AddText(fmt.Sprint(changes))
				}
			}

			if ahead > 0 {
				seg.AddText(" ⇑", fmt.Sprint(ahead))
			}

			if behind > 0 {
				seg.AddText(" ⇓", fmt.Sprint(behind))
			}
		}
	}
}

func warnOffline(seg *Segment) {
	if ifaces, err := net.Interfaces(); err == nil {
		for _, iface := range ifaces {
			if addrs, err := iface.Addrs(); err == nil {
				for _, addr := range addrs {
					var ip net.IP
					switch v := addr.(type) {
					case *net.IPNet:
						ip = v.IP
					case *net.IPAddr:
						ip = v.IP
					}
					if !ip.IsLoopback() && !ip.IsLinkLocalUnicast() {
						return
					}
				}
			}
		}
	}
	seg.SetTextColor(ColorOffline)
	seg.AddText("Offline")
}

func virtualEnv(seg *Segment) {
	if env := os.Getenv("VIRTUAL_ENV"); env != "" {
		if filepath.Base(env) == ".venv" {
			env = filepath.Base(filepath.Dir(env))
		}
		seg.SetTextColor(ColorVirtualEnv)
		seg.AddText(filepath.Base(env))
		if out, err := exec.Command("python", "--version").CombinedOutput(); err == nil {
			txt := string(out)
			txt = strings.TrimPrefix(txt, "Python ")
			txt = strings.TrimSuffix(txt, "\n")
			seg.AddText(" py", txt)
		}
	}
}

func dockerVersion(seg *Segment) {
	if match, err := filepath.Glob("Dockerfile"); err == nil && match != nil {
		out, _ := exec.Command("docker", "version", "--format", "{{.Client.Version}}").Output()
		txt := string(out)
		txt = strings.TrimSuffix(txt, "\n")
		txt = strings.TrimSpace(txt)
		if len(txt) > 0 {
			seg.SetTextColor(ColorDocker)
			seg.AddText("🐋", " ", txt)
		}
	}
}

func currentTime(seg *Segment) {
	seg.SetTextColor(ColorTime)
	seg.AddText("\\t")
}

func sshConnection(seg *Segment) {
	if os.Getenv("SSH_CLIENT") != "" {
		seg.SetTextColor(ColorSsh)
		seg.AddText("SSH")
	}
}

func userNameWithoutDefault(seg *Segment) {
	userNameWithDefault(seg, "")
}

func userNameWithDefault(seg *Segment, standard string) {
	if user := os.Getenv("USER"); user != standard {
		seg.SetTextColor(ColorUsername)
		seg.AddText("👤 ", user)
	}
}
