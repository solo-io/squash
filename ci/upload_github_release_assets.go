package main

import (
	"github.com/solo-io/go-utils/githubutils"
	"github.com/solo-io/go-utils/logger"
	"github.com/solo-io/go-utils/pkgmgmtutils"
)

func main() {
	const buildDir = "_output"
	const repoOwner = "solo-io"
	const repoName = "squash"

	assets := make([]githubutils.ReleaseAssetSpec, 3)
	assets[0] = githubutils.ReleaseAssetSpec{
		Name:       "squashctl-darwin",
		ParentPath: buildDir,
		UploadSHA:  true,
	}
	assets[1] = githubutils.ReleaseAssetSpec{
		Name:       "squashctl-linux",
		ParentPath: buildDir,
		UploadSHA:  true,
	}
	assets[2] = githubutils.ReleaseAssetSpec{
		Name:       "squashctl-windows.exe",
		ParentPath: buildDir,
		UploadSHA:  true,
	}

	spec := githubutils.UploadReleaseAssetSpec{
		Owner:             repoOwner,
		Repo:              repoName,
		Assets:            assets,
		SkipAlreadyExists: true,
	}
	githubutils.UploadReleaseAssetCli(&spec)

	fOpts := []pkgmgmtutils.FormulaOptions{
		{
			Name:           "homebrew-tap/squashctl",
			FormulaName:    "squashctl",
			Path:           "Formula/squashctl.rb",
			RepoOwner:      repoOwner,      // Make change in this repo owner
			RepoName:       "homebrew-tap", //   expects this repo is forked from PRRepoOwner if PRRepoOwner != RepoOwner
			PRRepoOwner:    repoOwner,      // Make PR to this repo owner
			PRRepoName:     "homebrew-tap", //   and this repo
			PRBranch:       "master",       //   and merge into this branch
			PRDescription:  "",
			PRCommitName:   "Solo-io Bot",
			PRCommitEmail:  "bot@solo.io",
			VersionRegex:   `version\s*"([0-9.]+)"`,
			DarwinShaRegex: `url\s*".*-darwin.*\W*sha256\s*"(.*)"`,
			LinuxShaRegex:  `url\s*".*-linux.*\W*sha256\s*"(.*)"`,
		},
		{
			Name:            "fish-food/squashctl",
			FormulaName:     "squashctl",
			Path:            "Food/squashctl.lua",
			RepoOwner:       repoOwner,
			RepoName:        "fish-food",
			PRRepoOwner:     "fishworks",
			PRRepoName:      "fish-food",
			PRBranch:        "master",
			PRDescription:   "",
			PRCommitName:    "Solo-io Bot",
			PRCommitEmail:   "bot@solo.io",
			VersionRegex:    `version\s*=\s*"([0-9.]+)"`,
			DarwinShaRegex:  `os\s*=\s*"darwin",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
			LinuxShaRegex:   `os\s*=\s*"linux",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
			WindowsShaRegex: `os\s*=\s*"windows",\W*.*\W*.*\W*.*\W*sha256\s*=\s*"(.*)",`,
		},
	}

	// Update package manager install formulas
	status, err := pkgmgmtutils.UpdateFormulas(repoOwner, repoName, buildDir,
		`squashctl-(darwin|linux|windows).*\.sha256`, fOpts)
	if err != nil {
		logger.Fatalf("Error trying to update package manager formulas. Error was: %s", err.Error())
	}
	for _, s := range status {
		if !s.Updated {
			if s.Err != nil {
				logger.Fatalf("Error while trying to update formula %s. Error was: %s", s.Name, s.Err.Error())
			} else {
				logger.Fatalf("Error while trying to update formula %s. Error was nil", s.Name) // Shouldn't happen; really bad if it does
			}
		}
		if s.Err != nil {
			if s.Err == pkgmgmtutils.ErrAlreadyUpdated {
				logger.Warnf("Formula %s was updated externally, so no updates applied during this release", s.Name)
			} else {
				logger.Fatalf("Error updating Formula %s. Error was: %s", s.Name, s.Err.Error())
			}
		}
	}
}
