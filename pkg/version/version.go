package version

import (
	"fmt"
	"runtime"
)

var (
	version    = "0.0.1"
	buildTime  = "1970-01-01T00:00:00Z"
	gitCommit  = "notSet"
	binaryName = "cluster0-operators"
)

func PrintVersion() string {
	version := fmt.Sprintf("%s v%s+%s", binaryName, version, gitCommit)
	return version
}

func GetGitCommit() string {
	return gitCommit
}

func GetBuildTime() string {
	return buildTime
}

func GetGoVersion() string {
	return runtime.Version()
}

func GetGoPlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

func GetGoCompiler() string {
	return runtime.Compiler
}
