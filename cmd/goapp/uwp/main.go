package main

import (
	"fmt"
	"path/filepath"

	"github.com/murlokswarm/app/internal/file"
)

func main() {
	windowsFiles("x64")
}

func windowsFiles(arch string) {
	synchronize([]sync{
		{
			src: file.RepoPath("drivers", "win", "uwp", arch, "Release", "goapp.dll"),
			dst: filepath.Join("uwp", arch, "goapp.dll"),
		},
		{
			src: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "AppX", "clrcompression.dll"),
			dst: filepath.Join("uwp", arch, "clrcompression.dll"),
		},
		{
			src: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "AppX", "uwp.dll"),
			dst: filepath.Join("uwp", arch, "uwp.dll"),
		},
		{
			src: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "AppX", "uwp.exe"),
			dst: filepath.Join("uwp", arch, "uwp.exe"),
		},
		{
			src: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "App.xbf"),
			dst: filepath.Join("uwp", arch, "App.xbf"),
		},
		{
			src: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "WindowPage.xbf"),
			dst: filepath.Join("uwp", arch, "WindowPage.xbf"),
		},
		{
			src: filepath.Join("uwp", arch, "goapp.dll"),
			dst: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Debug", "AppX", "goapp.dll"),
		},
		{
			src: filepath.Join("uwp", arch, "goapp.dll"),
			dst: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "AppX", "goapp.dll"),
		},
		{
			src: file.RepoPath("examples", "demo", "demo.app", "demo.exe"),
			dst: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Debug", "AppX", "demo.exe"),
		},
		{
			src: file.RepoPath("examples", "demo", "demo.app", "demo.exe"),
			dst: file.RepoPath("drivers", "win", "uwp", "uwp", "bin", arch, "Release", "AppX", "demo.exe"),
		},
	})
}

type sync struct {
	src string
	dst string
}

func synchronize(syncs []sync) {
	for _, s := range syncs {
		if err := file.Copy(s.dst, s.src); err != nil {
			fmt.Println("copy", s.src, "failed:", err)
		}
	}
}
