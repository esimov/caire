// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package fontscan

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// fcVars captures the environment configuration that determines how fontconfig resolves configuration
// files. It can be populated from the environment by [fcVarsFromEnv], but is decoupled from the environment
// for ease of testing.
type fcVars struct {
	// xdgDataHome is the location of the user's data files, extracted from $XDG_DATA_HOME.
	xdgDataHome string
	// xdgDataHome is the location of the user's config files, extracted from $XDG_CONFIG_HOME.
	xdgConfigHome string
	// userHome is the home directory of the current user, resolved from $HOME.
	userHome string
	// configFile is the name of the configuration file, extracted from $FONTCONFIG_FILE.
	configFile string
	// paths is the list of configuration paths, extracted from $FONTCONFIG_PATH
	paths []string
	// sysroot is the root directory of the fontconfig system logically. It is prepended
	// to all other paths, and is usually empty.
	sysroot string
}

func fcVarsFromEnv() fcVars {
	home := os.Getenv("HOME")
	return fcVars{
		xdgDataHome:   getEnvWithDefault("XDG_DATA_HOME", filepath.Join(home, ".local", "share")),
		xdgConfigHome: getEnvWithDefault("XDG_CONFIG_HOME", filepath.Join(home, ".config")),
		configFile:    getEnvWithDefault("FONTCONFIG_FILE", "fonts.conf"),
		paths:         filepath.SplitList(getEnvWithDefault("$FONTCONFIG_PATH", "/etc/fonts")),
		sysroot:       os.Getenv("FONTCONFIG_SYSROOT"),
		userHome:      home,
	}
}

// resolveRoot returns the path of the root fontconfig file according to the fcVars.
func (f fcVars) resolveRoot(logger Logger) string {
	return f.resolvePath(logger, f.configFile)
}

// resolvePath applies fontconfig's heuristics for finding a path referenced within its config.
func (f fcVars) resolvePath(logger Logger, path string) string {
	hasSysroot := len(f.sysroot) > 0
	if filepath.IsAbs(path) {
		if hasSysroot && !strings.HasPrefix(path, f.sysroot) {
			path = filepath.Join(f.sysroot, path)
		}
		return path
	}
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(f.userHome, strings.TrimPrefix(path, "~"))
		if hasSysroot {
			path = filepath.Join(f.sysroot, path)
		}
		return path
	}
	for _, p := range f.paths {
		candidate := filepath.Join(p, path)
		if hasSysroot {
			candidate = filepath.Join(f.sysroot, candidate)
		}
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		return candidate
	}
	logger.Printf("fontconfig referenced path %q, but it could not be resolved to a real path", path)
	return ""
}

func getEnvWithDefault(envVar string, defaultVal string) string {
	val, ok := os.LookupEnv(envVar)
	if !ok {
		return defaultVal
	}
	return val
}

const (
	_ = iota
	fcDir
	fcInclude
)

// fcDirective is either a <dir> or a <include> element,
// as indicated by [kind]
type fcDirective struct {
	dir struct {
		Dir    string `xml:",chardata"`
		Prefix string `xml:"prefix,attr"`
	}
	include struct {
		Include       string `xml:",chardata"`
		IgnoreMissing string `xml:"ignore_missing,attr"`
		Prefix        string `xml:"prefix,attr"`
	}
	kind uint8
}

func (directive *fcDirective) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case "dir":
		directive.kind = fcDir
		return d.DecodeElement(&directive.dir, &start)
	case "include":
		directive.kind = fcInclude
		return d.DecodeElement(&directive.include, &start)
	default:
		// ignore the element
		return d.Skip()
	}
}

// parseFcFile opens and process a FontConfig config file,
// returning the font directories to scan and the (optionnal)
// supplementary config files (or directories) to include.
// The file parameter is expected to already be resolved by
// resolvePath().
func (fc fcVars) parseFcFile(logger Logger, file, currentWorkingDir string) (fontDirs, includes []string, _ error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("opening fontconfig config file: %s", err)
	}
	defer f.Close()

	var config struct {
		Fontconfig []fcDirective `xml:",any"`
	}
	err = xml.NewDecoder(f).Decode(&config)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing fontconfig config file: %s", err)
	}

	// post-process : handle "prefix" attr and use absolute path
	for _, item := range config.Fontconfig {
		switch item.kind {
		case fcDir:
			dir := item.dir.Dir
			switch item.dir.Prefix {
			case "default", "cwd":
				dir = filepath.Join(currentWorkingDir, dir)
			case "relative":
				dir = filepath.Join(filepath.Dir(file), dir)
			case "xdg":
				dir = filepath.Join(fc.xdgDataHome, dir)
			}
			fontDirs = append(fontDirs, dir)
		case fcInclude:
			include := item.include.Include
			if item.include.Prefix == "xdg" {
				include = filepath.Join(fc.xdgConfigHome, include)
			}
			include = fc.resolvePath(logger, include)
			if len(include) > 0 {
				includes = append(includes, include)
			}
		}
	}
	return
}

// parseFcDir processes all the files in [dir] matching the [09]*.conf pattern
// seen is updated with the processed fontconfig files. The dir parameter is
// expected to already be resolved by resolvePath.
func (fc fcVars) parseFcDir(logger Logger, dir, currentWorkingDir string, seen map[string]bool) (fontDirs, includes []string, _ error) {
	entries, err := readDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading fontconfig config directory: %s", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if name := entry.Name(); strings.HasSuffix(name, ".conf") {
			c := name[0]
			if '0' <= c && c <= '9' {
				file := filepath.Join(dir, name)
				seen[file] = true
				fds, incs, err := fc.parseFcFile(logger, file, currentWorkingDir)
				if err != nil {
					return nil, nil, err
				}
				fontDirs = append(fontDirs, fds...)
				includes = append(includes, incs...)
			}
		}
	}

	return
}

// parseFcConfig recursively parses the fontconfig config file at [rootConfig]
// and its includes, returning the font directories to scan
func (fc fcVars) parseFcConfig(logger Logger) ([]string, error) {
	root := fc.resolveRoot(logger)
	seen := map[string]bool{root: true}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("processing fontconfig config file: %s", err)
	}

	// includes is a queue
	dirs, includes, err := fc.parseFcFile(logger, root, cwd)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(includes); i++ {
		include := includes[i]
		if seen[include] {
			continue
		}
		seen[include] = true

		fi, err := os.Stat(include)
		if err != nil { // gracefully ignore broken includes
			logger.Printf("missing fontconfig include %s: skipping", include)
			continue
		}

		var newDirs, newIncludes []string
		if fi.IsDir() {
			newDirs, newIncludes, err = fc.parseFcDir(logger, include, cwd, seen)
		} else {
			newDirs, newIncludes, err = fc.parseFcFile(logger, include, cwd)
		}
		if err != nil {
			return nil, err
		}

		dirs = append(dirs, newDirs...)
		includes = append(includes, newIncludes...)
	}

	return dirs, nil
}
