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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		cmdInit()
	case "clone":
		cmdClone()
	case "status":
		cmdStatus()
	case "add":
		cmdAdd()
	case "commit":
		cmdCommit()
	case "log":
		cmdLog()
	case "remote":
		cmdRemote()
	case "pull":
		cmdPull()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
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

func cmdInit() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	_, err := git.PlainInit(path, false)
	if err != nil {
		fmt.Printf("Error initializing repository: %v\n", err)
		os.Exit(1)
	}
	absPath, _ := filepath.Abs(path)
	fmt.Printf("Initialized empty Git repository in %s\n", absPath)
}

func cmdClone() {
	cloneFlag := flag.NewFlagSet("clone", flag.ExitOnError)
	branch := cloneFlag.String("b", "", "Branch to checkout")
	depth := cloneFlag.Int("depth", 0, "Create a shallow clone with a history truncated to the specified number of commits")
	force := cloneFlag.Bool("f", false, "Force overwrite existing directory")
	cloneFlag.Parse(os.Args[2:])

	args := cloneFlag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: gitGo clone [-b branch] [--depth depth] [-f] <url> [path]")
		os.Exit(1)
	}

	url := args[0]
	path := ""
	if len(args) > 1 {
		path = args[1]
	}

	if path == "" || path == "." {
		repoName := extractRepoName(url)
		path = repoName
	}

	fmt.Printf("Cloning into %s...\n", path)
	
	if *force {
		_, err := os.Stat(path)
		if err == nil {
			fmt.Println("Removing existing directory...")
			err = os.RemoveAll(path)
			if err != nil {
				fmt.Printf("Error removing existing directory: %v\n", err)
				os.Exit(1)
			}
		}
	}
	
	cloneOpts := &git.CloneOptions{
		URL:               url,
		Progress:          os.Stdout,
		SingleBranch:      true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	}
	
	if *branch != "" {
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(*branch)
	}
	
	if *depth > 0 {
		cloneOpts.Depth = *depth
	}
	
	_, err := git.PlainClone(path, false, cloneOpts)
	if err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		if *depth == 0 {
			fmt.Println("Try using --depth 1 for shallow clone if repository is large")
		}
		os.Exit(1)
	}
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

func cmdStatus() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}

	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	status, err := w.Status()
	if err != nil {
		fmt.Printf("Error getting status: %v\n", err)
		os.Exit(1)
	}

	if status.IsClean() {
		fmt.Println("Working tree clean")
		return
	}

	fmt.Println(status)
}

func cmdPull() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}

	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName("main"),
		Progress:      os.Stdout,
	})

	if err == git.NoErrAlreadyUpToDate {
		fmt.Println("Already up to date.")
		return
	}

	if err != nil {
		fmt.Printf("Error pulling repository: %v\n", err)
		fmt.Println("Try checking your remote configuration or network connection")
		os.Exit(1)
	}

	fmt.Println("Pull completed successfully")
}

func cmdAdd() {
	files := os.Args[2:]
	if len(files) == 0 {
		files = []string{"."}
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}

	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		_, err := w.Add(file)
		if err != nil {
			fmt.Printf("Error adding %s: %v\n", file, err)
			os.Exit(1)
		}
	}
	fmt.Printf("Added %d file(s)\n", len(files))
}

func cmdCommit() {
	commitFlag := flag.NewFlagSet("commit", flag.ExitOnError)
	message := commitFlag.String("m", "", "Commit message")
	commitFlag.Parse(os.Args[2:])

	if *message == "" {
		fmt.Println("Usage: gitGo commit -m <message>")
		os.Exit(1)
	}

	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}

	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("Error getting worktree: %v\n", err)
		os.Exit(1)
	}

	commit, err := w.Commit(*message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "gitGo User",
			Email: "user@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Printf("Error committing: %v\n", err)
		os.Exit(1)
	}

	obj, err := repo.CommitObject(commit)
	if err != nil {
		fmt.Printf("Error getting commit object: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Committed %s\n", obj.Hash.String()[:7])
}

func cmdLog() {
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}

	ref, err := repo.Head()
	if err != nil {
		fmt.Printf("Error getting HEAD: %v\n", err)
		os.Exit(1)
	}

	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		fmt.Printf("Error getting log: %v\n", err)
		os.Exit(1)
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Printf("%s - %s\n", c.Hash.String()[:7], c.Message)
		return nil
	})
	if err != nil {
		fmt.Printf("Error iterating log: %v\n", err)
		os.Exit(1)
	}
}

func cmdRemote() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gitGo remote add <name> <url>")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "add":
		if len(os.Args) < 5 {
			fmt.Println("Usage: gitGo remote add <name> <url>")
			os.Exit(1)
		}
		name := os.Args[3]
		url := os.Args[4]

		repo, err := git.PlainOpen(".")
		if err != nil {
			fmt.Printf("Error opening repository: %v\n", err)
			os.Exit(1)
		}

		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: name,
			URLs: []string{url},
		})
		if err != nil {
			fmt.Printf("Error adding remote: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Remote %s added with URL %s\n", name, url)
	default:
		fmt.Printf("Unknown remote subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}