package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var globalVerbose bool

func debugLog(format string, args ...interface{}) {
	if globalVerbose {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

func infoLog(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func main() {
	globalFlag := flag.NewFlagSet("gitgo", flag.ExitOnError)
	globalFlag.BoolVar(&globalVerbose, "v", false, "Enable verbose mode")
	globalFlag.BoolVar(&globalVerbose, "verbose", false, "Enable verbose mode")
	globalFlag.Parse(os.Args[1:])

	args := globalFlag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	debugLog("gitGo version 1.0.0")
	debugLog("Arguments: %v", os.Args)
	debugLog("Verbose mode: %v", globalVerbose)

	switch args[0] {
	case "init":
		cmdInit(args)
	case "clone":
		cmdClone(args)
	case "status":
		cmdStatus(args)
	case "add":
		cmdAdd(args)
	case "commit":
		cmdCommit(args)
	case "log":
		cmdLog(args)
	case "remote":
		cmdRemote(args)
	case "pull":
		cmdPull(args)
	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("gitGo - A zero-dependency Git command line tool")
	fmt.Println("Usage:")
	fmt.Println("  gitGo init [path]          Initialize a new git repository")
	fmt.Println("  gitGo clone <url> [path]   Clone a remote repository")
	fmt.Println("  gitGo pull                 Pull changes from remote")
	fmt.Println("  gitGo status               Show working tree status")
	fmt.Println("  gitGo add [files...]       Add files to staging area")
	fmt.Println("  gitGo commit -m <message>  Commit changes")
	fmt.Println("  gitGo log                  Show commit history")
	fmt.Println("  gitGo remote add <name> <url>  Add remote repository")
}

func cmdInit(args []string) {
	path := "."
	if len(args) > 1 {
		path = args[1]
	}

	debugLog("Initializing repository at: %s", path)
	
	_, err := git.PlainInit(path, false)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		os.Exit(1)
	}
	absPath, _ := filepath.Abs(path)
	debugLog("Repository initialized successfully at: %s", absPath)
	fmt.Printf("Initialized empty Git repository in %s\n", absPath)
}

func cmdClone(args []string) {
	debugLog("Starting clone command")
	
	cloneFlag := flag.NewFlagSet("clone", flag.ExitOnError)
	branch := cloneFlag.String("b", "", "Branch to checkout")
	depth := cloneFlag.Int("depth", 0, "Create a shallow clone with a history truncated to the specified number of commits")
	force := cloneFlag.Bool("f", false, "Force overwrite existing directory")
	cloneFlag.Parse(args[1:])

	flagArgs := cloneFlag.Args()
	if len(flagArgs) < 1 {
		fmt.Println("Usage: gitGo clone [-b branch] [--depth depth] [-f] <url> [path]")
		os.Exit(1)
	}

	url := flagArgs[0]
	path := ""
	if len(flagArgs) > 1 {
		path = flagArgs[1]
	}

	if path == "" || path == "." {
		repoName := extractRepoName(url)
		path = repoName
	}

	debugLog("Clone URL: %s", url)
	debugLog("Target path: %s", path)
	debugLog("Branch: %s", *branch)
	debugLog("Depth: %d", *depth)
	debugLog("Force: %v", *force)
	
	fmt.Printf("Cloning into %s...\n", path)
	
	info, err := os.Stat(path)
	if err == nil {
		debugLog("Path exists, checking type")
		if info.IsDir() {
			files, _ := os.ReadDir(path)
			if len(files) == 0 {
				debugLog("Directory is empty, cloning in-place")
				path = ""
			} else if *force {
				debugLog("Force flag set, removing existing directory")
				fmt.Println("Removing existing directory...")
				err = os.RemoveAll(path)
				if err != nil {
					fmt.Printf("Error removing existing directory: %v\n", err)
					fmt.Println("Make sure no files in the directory are in use")
					os.Exit(1)
				}
			} else {
				fmt.Printf("Error: directory '%s' already exists and is not empty\n", path)
				fmt.Println("Use -f flag to force overwrite")
				os.Exit(1)
			}
		} else {
			fmt.Printf("Error: '%s' is a file, not a directory\n", path)
			os.Exit(1)
		}
	} else {
		debugLog("Path does not exist, will create new directory")
	}
	
	debugLog("Creating CloneOptions")
	cloneOpts := &git.CloneOptions{
		URL:               url,
		Progress:          os.Stdout,
		SingleBranch:      false,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	
	if *branch != "" {
		debugLog("Setting branch reference: %s", *branch)
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(*branch)
		cloneOpts.SingleBranch = true
	}
	
	if *depth > 0 {
		debugLog("Setting clone depth: %d", *depth)
		cloneOpts.Depth = *depth
	}
	
	debugLog("Calling PlainClone...")
	_, err = git.PlainClone(path, false, cloneOpts)
	if err != nil {
		debugLog("Clone failed: %v", err)
		fmt.Printf("Error cloning repository: %v\n", err)
		fmt.Println("")
		fmt.Println("Possible solutions:")
		fmt.Println("1. Try using --depth 1 for shallow clone (most effective for Gitee/GitLab)")
		fmt.Println("2. Check if the URL is correct and accessible")
		fmt.Println("3. Try removing credentials from URL and use git credential helper")
		fmt.Println("4. Try using SSH protocol instead of HTTPS")
		fmt.Println("5. Check your network connection")
		fmt.Println("6. Try adding -v flag for verbose output")
		fmt.Println("")
		fmt.Println("Note: Some Git servers (like Gitee) may have protocol incompatibility with go-git library.")
		fmt.Println("If issues persist, try using the standard git command:")
		fmt.Printf("  git clone %s %s\n", url, path)
		os.Exit(1)
	}
	debugLog("Clone completed successfully")
	fmt.Println("Clone completed successfully")
}

func extractRepoName(url string) string {
	if len(url) == 0 {
		return "repo"
	}
	
	if idx := len(url) - 1; idx >= 0 && url[idx] == '/' {
		url = url[:idx]
	}
	
	if idx := len(url) - 1; idx >= 0 && url[idx] == '\\' {
		url = url[:idx]
	}
	
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' || url[i] == '\\' {
			name := url[i+1:]
			if len(name) > 4 && name[len(name)-4:] == ".git" {
				return name[:len(name)-4]
			}
			return name
		}
	}
	
	if len(url) > 4 && url[len(url)-4:] == ".git" {
		return url[:len(url)-4]
	}
	return url
}

func cmdStatus(args []string) {
	debugLog("Starting status command")
	wd, _ := os.Getwd()
	debugLog("Current directory: %s", wd)
	
	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	w, err := repo.Worktree()
	if err != nil {
		debugLog("Error getting worktree: %v", err)
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}
	debugLog("Worktree retrieved successfully")

	status, err := w.Status()
	if err != nil {
		debugLog("Error getting status: %v", err)
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}
	debugLog("Status retrieved successfully")

	if status.IsClean() {
		debugLog("Working tree is clean")
		fmt.Println("Working tree clean")
		return
	}

	debugLog("Status has changes")
	fmt.Println(status)
}

func cmdPull(args []string) {
	debugLog("Starting pull command")
	
	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	w, err := repo.Worktree()
	if err != nil {
		debugLog("Error getting worktree: %v", err)
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}
	debugLog("Worktree retrieved successfully")

	debugLog("Starting pull from origin/main")
	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName("main"),
		Progress:      os.Stdout,
	})

	if err == git.NoErrAlreadyUpToDate {
		debugLog("Already up to date")
		fmt.Println("Already up to date.")
		return
	}

	if err != nil {
		debugLog("Error pulling: %v", err)
		fmt.Printf("Error pulling repository: %v\n", err)
		fmt.Println("Try checking your remote configuration or network connection")
		os.Exit(1)
	}

	debugLog("Pull completed successfully")
	fmt.Println("Pull completed successfully")
}

func cmdAdd(args []string) {
	debugLog("Starting add command")
	
	files := args[1:]
	if len(files) == 0 {
		files = []string{"."}
	}
	debugLog("Files to add: %v", files)

	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	w, err := repo.Worktree()
	if err != nil {
		debugLog("Error getting worktree: %v", err)
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}
	debugLog("Worktree retrieved successfully")

	for _, file := range files {
		debugLog("Adding file: %s", file)
		_, err := w.Add(file)
		if err != nil {
			debugLog("Error adding file %s: %v", file, err)
			fmt.Printf("Error adding %s: %v\n", file, err)
			os.Exit(1)
		}
		debugLog("File %s added successfully", file)
	}
	debugLog("Added %d file(s)", len(files))
	fmt.Printf("Added %d file(s)\n", len(files))
}

func cmdCommit(args []string) {
	debugLog("Starting commit command")
	
	commitFlag := flag.NewFlagSet("commit", flag.ExitOnError)
	message := commitFlag.String("m", "", "Commit message")
	commitFlag.Parse(args[1:])

	if *message == "" {
		fmt.Println("Usage: gitGo commit -m <message>")
		os.Exit(1)
	}
	debugLog("Commit message: %s", *message)

	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	w, err := repo.Worktree()
	if err != nil {
		debugLog("Error getting worktree: %v", err)
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}
	debugLog("Worktree retrieved successfully")

	debugLog("Creating commit with message: %s", *message)
	commit, err := w.Commit(*message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "gitGo User",
			Email: "user@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		debugLog("Error committing: %v", err)
		fmt.Printf("Error committing: %v\n", err)
		os.Exit(1)
	}
	debugLog("Commit created: %s", commit.String())

	obj, err := repo.CommitObject(commit)
	if err != nil {
		debugLog("Error getting commit object: %v", err)
		fmt.Printf("Error getting commit object: %v\n", err)
		os.Exit(1)
	}

	debugLog("Commit hash: %s", obj.Hash.String())
	fmt.Printf("Committed %s\n", obj.Hash.String()[:7])
}

func cmdLog(args []string) {
	debugLog("Starting log command")
	
	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	ref, err := repo.Head()
	if err != nil {
		debugLog("Error getting HEAD: %v", err)
		fmt.Printf("Error getting HEAD: %v\n", err)
		os.Exit(1)
	}

	debugLog("Getting HEAD reference")
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		debugLog("Error getting log: %v", err)
		fmt.Printf("Error getting log: %v\n", err)
		os.Exit(1)
	}
	debugLog("Log iterator created successfully")

	debugLog("Iterating commit log")
	err = cIter.ForEach(func(c *object.Commit) error {
		debugLog("Commit: %s - %s", c.Hash.String()[:7], c.Message)
		fmt.Printf("%s - %s\n", c.Hash.String()[:7], c.Message)
		return nil
	})
	if err != nil {
		debugLog("Error iterating log: %v", err)
		fmt.Printf("Error iterating log: %v\n", err)
		os.Exit(1)
	}
	debugLog("Log iteration completed successfully")
}

func cmdRemote(args []string) {
	debugLog("Starting remote command")
	
	if len(args) < 2 {
		fmt.Println("Usage: gitGo remote add <name> <url>")
		os.Exit(1)
	}

	switch args[1] {
	case "add":
		if len(args) < 4 {
			fmt.Println("Usage: gitGo remote add <name> <url>")
			os.Exit(1)
		}
		name := args[2]
		url := args[3]
		debugLog("Adding remote: %s -> %s", name, url)

		repo, err := git.PlainOpen(".")
		if err != nil {
			debugLog("Error opening repository: %v", err)
			fmt.Printf("Error opening repository: %v\n", err)
			os.Exit(1)
		}
		debugLog("Repository opened successfully")

		debugLog("Creating remote: %s", name)
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: name,
			URLs: []string{url},
		})
		if err != nil {
			debugLog("Error adding remote: %v", err)
			fmt.Printf("Error adding remote: %v\n", err)
			os.Exit(1)
		}
		debugLog("Remote %s added successfully", name)
		fmt.Printf("Remote %s added with URL %s\n", name, url)
	default:
		fmt.Printf("Unknown remote subcommand: %s\n", args[1])
		os.Exit(1)
	}
}