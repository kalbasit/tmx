## swm tmux kill-server

Kill the server closes the tmux session for this profile and story

### Synopsis

Kill the server closes the tmux session for this profile and story

```
swm tmux kill-server [flags]
```

### Options

```
  -h, --help                help for kill-server
      --story-name string   The name of the story
      --vim-exit            if vim is found running, kill it
```

### Options inherited from parent commands

```
      --code-path string              The path to the code directory
      --debug                         Enable debugging
      --repositories-dirname string   The name of the repositories directory, a child directory of the code-path and the parent directory for all repositories (default "repositories")
      --stories-dirname string        The name of the stories directory, a child directory of the code-path and the parent directory for all stories (default "stories")
```

### SEE ALSO

* [swm tmux](swm_tmux.md)	 - Manage tmux sessions

###### Auto generated by spf13/cobra on 4-Aug-2020