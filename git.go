package main

import (
	"os/exec"
	"strings"
)

func readGit(root string) *GitInfo {
	branch := runGit(root, "rev-parse", "--abbrev-ref", "HEAD")
	changedRaw := runGit(root, "status", "--porcelain")
	var changed []string
	for _, l := range strings.Split(changedRaw, "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		parts := strings.SplitN(l, " ", 2)
		p := parts[len(parts)-1]
		p = strings.TrimSpace(p)
		if p != "" {
			changed = append(changed, p)
		}
	}
	gi := &GitInfo{Branch: strings.TrimSpace(branch), Changed: changed}
	if gi.Branch == "" && len(changed) == 0 {
		return nil
	}
	return gi
}

func runGit(root string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	b, _ := cmd.CombinedOutput()
	return string(b)
}
