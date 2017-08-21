package code

import (
	"errors"
	"regexp"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
)

var (
	// AppFS represents the filesystem of the app. It is exported to be used as a
	// test helper.
	AppFS afero.Fs

	// ErrCodePathEmpty is returned if Code.Path is empty or invalid
	ErrCodePathEmpty = errors.New("code path is empty or does not exist")

	// ErrProfileNoFound is returned if the profile was not found
	ErrProfileNoFound = errors.New("profile not found")

	// ErrProjectNotFound is returned if the session name did not yield a project
	// we know about
	ErrProjectNotFound = errors.New("project not found")

	// ErrInvalidURL is returned by AddProject if the URL given is not valid
	ErrInvalidURL = errors.New("invalid URL given")

	// ErrProjectAlreadyExists is returned if the project already exists
	ErrProjectAlreadyExists = errors.New("project already exists")

	// ErrCoderNotScanned is returned if San() was never called
	ErrCoderNotScanned = errors.New("code was not scanned")
)

func init() {
	// initialize AppFs to use the OS filesystem
	AppFS = afero.NewOsFs()
}

// code implements the coder interface
type code struct {
	// path is the base path of this profile
	path string

	// excludePattern is a list of patterns to ignore
	excludePattern *regexp.Regexp

	mu       sync.RWMutex
	profiles map[string]*profile
}

// New returns a new empty Code, caller must call Load to load from cache or
// scan the code directory
func New(p string, ignore *regexp.Regexp) Coder {
	return &code{
		path:           p,
		excludePattern: ignore,
		profiles:       make(map[string]*profile),
	}
}

// Path returns the absolute path of this coder
func (c *code) Path() string { return c.path }

// Profile returns the profile given it's name or an error if no profile with
// this name was found
func (c *code) Profile(name string) (Profile, error) { return c.getProfile(name) }

// Scan loads the code from the cache (if it exists), otherwise it will
// initiate a full scan and save it in cache.
func (c *code) Scan() error {
	// validate the Code, we cannot load an invalid Code
	if err := c.validate(); err != nil {
		return err
	}
	c.scan()

	return nil
}

// getProfile return the profile identified by name
func (c *code) getProfile(name string) (*profile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// make sure we scanned already
	if len(c.profiles) == 0 {
		return nil, ErrCoderNotScanned
	}
	// get the profile
	p, ok := c.profiles[name]
	if !ok {
		return nil, ErrProfileNoFound
	}
	return p, nil
}

// addProfile adds the profile to the list of profiles
func (c *code) addProfile(name string) *profile {
	// if the profile already exists, return it
	if p, err := c.getProfile(name); err == nil {
		return p
	}
	// otherwise add the profile to the map
	p := newProfile(c, name)
	c.mu.Lock()
	c.profiles[name] = p
	c.mu.Unlock()

	return p
}

// scan scans the entire profile to build the workspaces
func (c *code) scan() {
	// initialize the variables
	var wg sync.WaitGroup
	// read the profile and scan all profiles
	entries, err := afero.ReadDir(AppFS, c.path)
	if err != nil {
		log.Error().Str("path", c.path).Msgf("error reading the directory: %s", err)
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			// does this folder match the exclude pattern?
			if c.excludePattern != nil && c.excludePattern.MatchString(entry.Name()) {
				continue
			}
			// create the profile
			log.Debug().Msgf("found profile: %s", entry.Name())
			wg.Add(1)
			go func(name string) {
				p := c.addProfile(name)
				p.scan()
				wg.Done()
			}(entry.Name())
		}
	}
	wg.Wait()
}

func (c *code) validate() error {
	if c.path == "" {
		return ErrCodePathEmpty
	}
	if _, err := AppFS.Stat(c.path); err != nil {
		return ErrCodePathEmpty
	}

	return nil
}
