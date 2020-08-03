package code

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/kalbasit/swm/ifaces"
	"github.com/kalbasit/swm/testhelper"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// discard logs
	log.Logger = zerolog.New(ioutil.Discard)
}

func TestScan(t *testing.T) {
	// create a temporary directory
	dir, err := ioutil.TempDir("", "swm-test-*")
	require.NoError(t, err)

	// delete it once we are done here
	defer func() { os.RemoveAll(dir) }()

	// create the filesystem we want to scan
	require.NoError(t, testhelper.CreateProjects(dir))

	// define the assertion function
	assertFn := func(c ifaces.Code, story_name string) {
		// assert the repositories
		for _, importPath := range []string{"github.com/owner1/repo1", "github.com/owner2/repo2", "github.com/owner3/repo3"} {
			prj, err := c.GetProjectByRelativePath(importPath)
			require.NoError(t, err)

			assert.Equal(t, importPath, prj.String())
			assert.Equal(t, path.Join(dir, "repositories", importPath), prj.RepositoryPath())
			if story_name != "" {
				assert.NoError(t, prj.Ensure())

				sn, err := prj.StoryPath()
				if assert.NoError(t, err) {
					assert.Equal(t, path.Join(dir, "stories", story_name, importPath), sn)
				}
			}
		}
	}

	// create a code without a story
	c := New(dir, regexp.MustCompile("^.snapshots$"))
	require.NoError(t, c.Scan())
	assertFn(c, "")

	// create a new code with a story
	sc := New(dir, regexp.MustCompile("^.snapshots$"))
	sc.SetStoryName(t.Name())
	require.NoError(t, sc.Scan())
	assertFn(sc, t.Name())
}

func TestPath(t *testing.T) {
	c := &code{path: "/code"}
	assert.Equal(t, "/code", c.Path())
}

func TestGetProject(t *testing.T) {
	// create a temporary directory
	dir, err := ioutil.TempDir("", "swm-test-*")
	require.NoError(t, err)

	// delete it once we are done here
	defer func() { os.RemoveAll(dir) }()

	// create the filesystem we want to scan
	require.NoError(t, testhelper.CreateProjects(dir))

	testCases := []struct {
		story_name string
	}{
		{
			story_name: "",
		},
		{
			story_name: t.Name(),
		},
	}

	for _, testCase := range testCases {
		// create a code
		c := New(dir, regexp.MustCompile("^.snapshots$"))
		c.SetStoryName(testCase.story_name)
		require.NoError(t, c.Scan())

		// get the project and assert things
		for _, importPath := range []string{"github.com/owner1/repo1", "github.com/owner2/repo2", "github.com/owner3/repo3"} {
			prj, err := c.GetProjectByRelativePath(importPath)
			require.NoError(t, err)
			assert.Equal(t, path.Join(dir, "repositories", importPath), prj.RepositoryPath())

			if testCase.story_name != "" {
				sp, err := prj.StoryPath()
				if assert.NoError(t, err) {
					assert.Equal(t, path.Join(dir, "stories", testCase.story_name, importPath), sp)
				}
			}
		}
	}
}

func TestProjects(t *testing.T) {
	// create a temporary directory
	dir, err := ioutil.TempDir("", "swm-test-*")
	require.NoError(t, err)

	// delete it once we are done here
	defer func() { os.RemoveAll(dir) }()

	// create the filesystem we want to scan
	require.NoError(t, testhelper.CreateProjects(dir))

	// create a code
	c := New(dir, regexp.MustCompile("^.snapshots$"))
	require.NoError(t, c.Scan())

	// get all the projects, and collect their import paths, then compare those to the expected ones
	expectedImportPaths := []string{"github.com/owner1/repo1", "github.com/owner2/repo2", "github.com/owner3/repo3"}
	sort.Strings(expectedImportPaths)

	prjs := c.Projects()
	if assert.Len(t, prjs, len(expectedImportPaths)) {
		var gotImportPaths []string
		for _, prj := range prjs {
			gotImportPaths = append(gotImportPaths, prj.String())
		}

		sort.Strings(gotImportPaths)

		assert.EqualValues(t, expectedImportPaths, gotImportPaths)
	}
}

func TestClone(t *testing.T) {
	// create a temporary directory
	dir, err := ioutil.TempDir("", "swm-test-*")
	require.NoError(t, err)

	// delete it once we are done here
	defer func() { os.RemoveAll(dir) }()

	// create the filesystem we want to scan
	require.NoError(t, testhelper.CreateProjects(dir))

	// create a code
	c := New(dir, regexp.MustCompile("^.snapshots$"))
	require.NoError(t, c.Scan())

	// clone the repo4 from the ignored location, but first validate it does not exist in the scanned projects
	importPath := strings.TrimPrefix(path.Join(dir, ".snapshots", "github.com/owner4/repo4"), string(os.PathSeparator))
	_, err = c.GetProjectByRelativePath(importPath)
	require.True(t, errors.Is(err, ErrProjectNotFound))

	err = c.Clone(fmt.Sprintf("file://%s", path.Join(dir, ".snapshots", "github.com/owner4/repo4")))
	if assert.NoError(t, err) {
		prj, err := c.GetProjectByRelativePath(importPath)
		if assert.NoError(t, err) {
			assert.Equal(t, importPath, prj.String())
			assert.Equal(t, path.Join(c.RepositoriesDir(), importPath), prj.RepositoryPath())
		}
	}
}

func TestGetProjectByAbsolutePath(t *testing.T) {
	// create a temporary directory
	dir, err := ioutil.TempDir("", "swm-test-*")
	require.NoError(t, err)

	// delete it once we are done here
	defer func() { os.RemoveAll(dir) }()

	// create the filesystem we want to scan
	require.NoError(t, testhelper.CreateProjects(dir))

	// create a code
	c := New(dir, regexp.MustCompile("^.snapshots$"))
	require.NoError(t, c.Scan())

	tests := map[string]string{
		dir + "/repositories/github.com/owner1/repo1": "github.com/owner1/repo1",

		dir + "/repositories/github.com/owner2/repo2": "github.com/owner2/repo2",
	}

	for p, ip := range tests {
		prj, err := c.GetProjectByAbsolutePath(p)
		if assert.NoError(t, err) {
			assert.Equal(t, ip, prj.String())
		}
	}

	_, err = c.GetProjectByAbsolutePath("/code/not-existing/base")
	assert.Error(t, err)
	_, err = c.GetProjectByAbsolutePath(dir + "/repositories/github.com/user/repo")
	assert.Error(t, err)
}

func TestStoryName(t *testing.T) {
	t.Run("no story name", func(t *testing.T) {
		c := &code{}
		assert.Empty(t, c.StoryName())
	})

	t.Run("story name is set", func(t *testing.T) {
		c := &code{story_name: "foobar"}
		assert.Equal(t, "foobar", c.StoryName())
	})
}

func TestStoryBranchName(t *testing.T) {
	t.Run("no story branch name or a story name", func(t *testing.T) {
		c := &code{}
		assert.Empty(t, c.StoryBranchName())
	})

	t.Run("no story branch name but a story name", func(t *testing.T) {
		c := &code{story_name: "foobar"}
		assert.Equal(t, "foobar", c.StoryBranchName())
	})

	t.Run("story branch name, an no story name", func(t *testing.T) {
		c := &code{story_branch_name: "foobar"}
		assert.Equal(t, "foobar", c.StoryBranchName())
	})

	t.Run("story branch name and a story name", func(t *testing.T) {
		c := &code{story_branch_name: "foobar", story_name: "nope"}
		assert.Equal(t, "foobar", c.StoryBranchName())
	})
}
