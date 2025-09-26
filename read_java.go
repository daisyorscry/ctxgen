package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func readJava(root string) *JavaInfo {
	pom := filepath.Join(root, "pom.xml")
	if exists(pom) {
		b, _ := os.ReadFile(pom)
		ji := &JavaInfo{BuildTool: "maven", Deps: map[string]string{}}
		reGA := regexp.MustCompile(`(?s)<groupId>([^<]+)</groupId>.*?<artifactId>([^<]+)</artifactId>`)
		if m := reGA.FindSubmatch(b); len(m) >= 3 {
			ji.GroupID = strings.TrimSpace(string(m[1]))
			ji.Artifact = strings.TrimSpace(string(m[2]))
		}
		reDep := regexp.MustCompile(`(?s)<dependency>.*?<groupId>([^<]+)</groupId>.*?<artifactId>([^<]+)</artifactId>.*?(?:<version>([^<]+)</version>)?.*?</dependency>`)
		for _, mm := range reDep.FindAllSubmatch(b, -1) {
			key := strings.TrimSpace(string(mm[1])) + ":" + strings.TrimSpace(string(mm[2]))
			val := ""
			if len(mm) >= 4 {
				val = strings.TrimSpace(string(mm[3]))
			}
			ji.Deps[key] = val
		}
		return ji
	}
	gradle := firstExist(root, []string{"build.gradle.kts", "build.gradle"})
	if gradle != "" {
		b, _ := os.ReadFile(filepath.Join(root, gradle))
		ji := &JavaInfo{BuildTool: "gradle", Deps: map[string]string{}}
		rePlugin := regexp.MustCompile(`id\("([^"]+)"\)`)
		for _, mm := range rePlugin.FindAllSubmatch(b, -1) {
			ji.Plugins = append(ji.Plugins, string(mm[1]))
		}
		reDep := regexp.MustCompile(`["']([^:"']+):([^:"']+):([^"']+)["']`)
		for _, mm := range reDep.FindAllSubmatch(b, -1) {
			key := string(mm[1]) + ":" + string(mm[2])
			ji.Deps[key] = string(mm[3])
		}
		return ji
	}
	return nil
}
