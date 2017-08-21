package code

import (
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/kalbasit/swm/testhelper"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// discard logs
	log.Logger = zerolog.New(ioutil.Discard)
}

func TestCodeScan(t *testing.T) {
	// swap the filesystem
	oldAppFS := AppFS
	AppFS = afero.NewMemMapFs()
	defer func() { AppFS = oldAppFS }()
	// create the filesystem we want to scan
	testhelper.CreateProjects(t, AppFS)
	// create a code
	c := New("/code", regexp.MustCompile("^.snapshots$"))
	// define the assertion function
	assertFn := func() {
		// create the expected structs
		p := newProfile(c.(*code), t.Name())
		baseStory := p.addStory(baseStoryName)
		baseStory.addProject("github.com/kalbasit/swm")
		baseStory.addProject("github.com/kalbasit/dotfiles")
		baseStory.addProject("github.com/kalbasit/workflow")
		story123 := p.addStory("STORY-123")
		story123.addProject("github.com/kalbasit/swm")
		story123.addProject("github.com/kalbasit/dotfiles")

		// get the profile
		profile, err := c.(*code).getProfile(t.Name())
		require.NoError(t, err)

		// assert the base story
		assert.Equal(t, baseStory.name, profile.getStory("base").name)
		assert.Equal(t, baseStory.profile.name, profile.getStory("base").profile.name)
		assert.Equal(t, baseStory.getProjects()["github.com/kalbasit/swm"].importPath, profile.getStory("base").getProjects()["github.com/kalbasit/swm"].importPath)
		assert.Equal(t, baseStory.getProjects()["github.com/kalbasit/dotfiles"].importPath, profile.getStory("base").getProjects()["github.com/kalbasit/dotfiles"].importPath)
		assert.Equal(t, baseStory.getProjects()["github.com/kalbasit/workflow"].importPath, profile.getStory("base").getProjects()["github.com/kalbasit/workflow"].importPath)

		// assert the STORY-123 story
		assert.Equal(t, story123.name, profile.getStory("STORY-123").name)
		assert.Equal(t, story123.profile.name, profile.getStory("STORY-123").profile.name)
		assert.Equal(t, story123.getProjects()["github.com/kalbasit/dotfiles"].importPath, profile.getStory("STORY-123").getProjects()["github.com/kalbasit/dotfiles"].importPath)
		assert.Equal(t, story123.getProjects()["github.com/kalbasit/swm"].importPath, profile.getStory("STORY-123").getProjects()["github.com/kalbasit/swm"].importPath)
	}
	// scan now
	require.NoError(t, c.Scan())
	assertFn()
}

func TestCodeProfile(t *testing.T) {
	// swap the filesystem
	oldAppFS := AppFS
	AppFS = afero.NewMemMapFs()
	defer func() { AppFS = oldAppFS }()
	// create the filesystem we want to scan
	testhelper.CreateProjects(t, AppFS)
	// create a code
	c := New("/code", regexp.MustCompile("^.snapshots$"))
	// define the assertion function
	assertFn := func(pTest *profile) {
		// create the expected structs
		p := newProfile(c.(*code), t.Name())
		baseStory := p.addStory(baseStoryName)
		baseStory.addProject("github.com/kalbasit/swm")
		baseStory.addProject("github.com/kalbasit/dotfiles")
		baseStory.addProject("github.com/kalbasit/workflow")
		story123 := p.addStory("STORY-123")
		story123.addProject("github.com/kalbasit/swm")
		story123.addProject("github.com/kalbasit/dotfiles")

		// assert the base story
		assert.Equal(t, baseStory.name, pTest.getStory("base").name)
		assert.Equal(t, baseStory.profile.name, pTest.getStory("base").profile.name)
		assert.Equal(t, baseStory.getProjects()["github.com/kalbasit/swm"].importPath, pTest.getStory("base").getProjects()["github.com/kalbasit/swm"].importPath)
		assert.Equal(t, baseStory.getProjects()["github.com/kalbasit/dotfiles"].importPath, pTest.getStory("base").getProjects()["github.com/kalbasit/dotfiles"].importPath)
		assert.Equal(t, baseStory.getProjects()["github.com/kalbasit/workflow"].importPath, pTest.getStory("base").getProjects()["github.com/kalbasit/workflow"].importPath)

		// assert the STORY-123 story
		assert.Equal(t, story123.name, pTest.getStory("STORY-123").name)
		assert.Equal(t, story123.profile.name, pTest.getStory("STORY-123").profile.name)
		assert.Equal(t, story123.getProjects()["github.com/kalbasit/dotfiles"].importPath, pTest.getStory("STORY-123").getProjects()["github.com/kalbasit/dotfiles"].importPath)
		assert.Equal(t, story123.getProjects()["github.com/kalbasit/swm"].importPath, pTest.getStory("STORY-123").getProjects()["github.com/kalbasit/swm"].importPath)
	}
	// assert it throws an error before scanning
	_, err := c.Profile(t.Name())
	assert.EqualError(t, err, ErrCoderNotScanned.Error())
	// scan now
	require.NoError(t, c.Scan())
	// get the profile
	p, err := c.Profile(t.Name())
	require.NoError(t, c.Scan())
	assertFn(p.(*profile))
}

func TestPath(t *testing.T) {
	c := &code{path: "/code"}
	assert.Equal(t, "/code", c.Path())
}
