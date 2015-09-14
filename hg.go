package vcs

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var hgDetectURL = regexp.MustCompile("default = (?P<foo>.+)\n")

// NewHgRepo creates a new instance of HgRepo. The remote and local directories
// need to be passed in.
func NewHgRepo(remote, local string) (*HgRepo, error) {
	ltype, err := DetectVcsFromFS(local)

	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}

	r := &HgRepo{}
	r.setRemote(remote)
	r.setLocalPath(local)
	r.Logger = Logger

	// Make sure the local Hg repo is configured the same as the remote when
	// A remote value was passed in.
	if err == nil && r.CheckLocal() == true {
		// An Hg repo was found so test that the URL there matches
		// the repo passed in here.
		oldDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		os.Chdir(local)
		defer os.Chdir(oldDir)
		out, err := exec.Command("hg", "paths").CombinedOutput()
		if err != nil {
			return nil, err
		}

		m := hgDetectURL.FindStringSubmatch(string(out))
		if m[1] != "" && m[1] != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Hg repo use that one.
		if remote == "" && m[1] != "" {
			r.setRemote(m[1])
		}
	}

	return r, nil
}

// HgRepo implements the Repo interface for the Mercurial source control.
type HgRepo struct {
	base
}

// Vcs retrieves the underlying VCS being implemented.
func (s HgRepo) Vcs() Type {
	return Hg
}

// Get is used to perform an initial clone of a repository.
func (s *HgRepo) Get() error {
	_, err := s.run("hg", "clone", "-U", s.Remote(), s.LocalPath())
	return err
}

// Update performs a Mercurial pull to an existing checkout.
func (s *HgRepo) Update() error {
	_, err := s.runFromDir("hg", "update")
	return err
}

// UpdateVersion sets the version of a package currently checked out via Hg.
func (s *HgRepo) UpdateVersion(version string) error {
	_, err := s.runFromDir("hg", "update", version)
	return err
}

// Version retrieves the current version.
func (s *HgRepo) Version() (string, error) {
	out, err := s.runFromDir("hg", "identify")
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(string(out), " ", 2)
	sha := parts[0]
	return strings.TrimSpace(sha), nil
}

// CheckLocal verifies the local location is a Git repo.
func (s *HgRepo) CheckLocal() bool {
	if _, err := os.Stat(s.LocalPath() + "/.hg"); err == nil {
		return true
	}

	return false

}
