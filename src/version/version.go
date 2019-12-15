package version

import "fmt"

var (
	Version string

	//will be overwritten automatically by the build system
	GitCommit string
	GoVersion string
	BuildTime string
)

func FullVersion() string {
	return fmt.Sprintf("Version: %6s \nGit commit: %6s \nGo version: %6s \nBuild time: %6s \n",
		Version, GitCommit, GoVersion, BuildTime)
}
func AppVersion() string {
	return fmt.Sprintf(`
		##       ##         ##        ##   ##       ##       ## ##     ## ## ##      ## ## 
		##       ##       ## ##     ## ##  ##       ##    ##      ##      ##      ##      ##
		##       ##      ##  ##    ##  ##  ##       ##   ##               ##     ##
		##       ##     ##   ##   ##   ##  ##       ##    ## ## ##        ##     ##
		##       ##    ##    ##  ##    ##  ##       ##            ##      ##     ## 
		##       ##   ##     ## ##     ##  ##       ##  ##        ##      ##      ##      ##
		## ## ## ##  ##      ####      ##  ## ## ## ##   ## ## ##      ## ## ##    ## ## ##
		
                       %s`+"  by cnsilvan（https://github.com/cnsilvan/UnblockNeteaseMusic） \n", Version)
}
