package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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
	case "checkout":
		cmdCheckout(args)
	case "branch":
		cmdBranch(args)
	case "push":
		cmdPush(args)
	case "test":
		cmdTest(args)
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
	fmt.Println("  gitGo checkout <branch>    Checkout a branch")
	fmt.Println("  gitGo branch               List all branches")
	fmt.Println("  gitGo status               Show working tree status")
	fmt.Println("  gitGo add [files...]       Add files to staging area")
	fmt.Println("  gitGo commit -m <message>  Commit changes")
	fmt.Println("  gitGo push                 Push changes to remote")
	fmt.Println("  gitGo log                  Show commit history")
	fmt.Println("  gitGo test                 Run Go tests")
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
	sshKey := cloneFlag.String("ssh-key", "", "Path to SSH private key (default: ~/.ssh/id_rsa)")
	cloneFlag.Parse(args[1:])

	flagArgs := cloneFlag.Args()
	if len(flagArgs) < 1 {
		fmt.Println("Usage: gitGo clone [-b branch] [--depth depth] [-f] [--ssh-key path] <url> [path]")
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
	debugLog("SSH Key: %s", *sshKey)

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

	if strings.HasPrefix(url, "https://") {
		debugLog("Skipping SSL verification by default")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		debugLog("Detected SSH URL, setting up authentication")
		keyPath := *sshKey
		if keyPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				os.Exit(1)
			}
			keyPath = filepath.Join(home, ".ssh", "id_rsa")
			debugLog("Using default SSH key path: %s", keyPath)
		}

		_, err := os.Stat(keyPath)
		if err != nil {
			fmt.Printf("Error: SSH key not found at %s\n", keyPath)
			fmt.Println("Please provide SSH key path with --ssh-key flag")
			os.Exit(1)
		}

		publicKeys, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
		if err != nil {
			debugLog("Error creating SSH public keys: %v", err)
			fmt.Printf("Error loading SSH key: %v\n", err)
			os.Exit(1)
		}
		cloneOpts.Auth = publicKeys
	}

	if *branch != "" {
		debugLog("Setting branch reference: %s", *branch)
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName(*branch)
		cloneOpts.SingleBranch = true
	} else {
		debugLog("No branch specified, trying main then master")
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName("main")
		cloneOpts.SingleBranch = true
	}

	if *depth > 0 {
		debugLog("Setting clone depth: %d", *depth)
		cloneOpts.Depth = *depth
	}

	debugLog("Calling PlainClone with branch: %s", cloneOpts.ReferenceName)
	_, err = git.PlainClone(path, false, cloneOpts)

	if err != nil && *branch == "" {
		debugLog("Clone failed with main branch, trying master")
		cloneOpts.ReferenceName = plumbing.NewBranchReferenceName("master")
		_, err = git.PlainClone(path, false, cloneOpts)
	}
	if err != nil {
		debugLog("Clone failed: %v", err)
		if strings.Contains(err.Error(), "remote repository is empty") {
			fmt.Println("Remote repository is empty, initializing local repository...")
			if path == "" {
				path = "."
			}
			_, initErr := git.PlainInit(path, false)
			if initErr != nil {
				fmt.Printf("Error initializing local repository: %v\n", initErr)
				os.Exit(1)
			}
			repo, initErr := git.PlainOpen(path)
			if initErr != nil {
				fmt.Printf("Error opening local repository: %v\n", initErr)
				os.Exit(1)
			}
			_, initErr = repo.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{url},
			})
			if initErr != nil {
				fmt.Printf("Error adding remote origin: %v\n", initErr)
				os.Exit(1)
			}
			absPath, _ := filepath.Abs(path)
			fmt.Printf("Initialized empty Git repository in %s with remote origin\n", absPath)
			return
		}
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

	pullFlag := flag.NewFlagSet("pull", flag.ExitOnError)
	force := pullFlag.Bool("f", false, "Force pull and discard local changes")
	pullFlag.Parse(args[1:])

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

	if *force {
		debugLog("Force flag set, resetting worktree")
		err = w.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
		if err != nil {
			debugLog("Error resetting worktree: %v", err)
			fmt.Printf("Error resetting worktree: %v\n", err)
			os.Exit(1)
		}
		debugLog("Worktree reset successfully")
	}

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
		fmt.Println("")
		fmt.Println("Possible solutions:")
		fmt.Println("1. Commit or stash your changes before pulling")
		fmt.Println("2. Use -f flag to discard local changes and force pull")
		fmt.Println("3. Check your remote configuration")
		fmt.Println("4. Check your network connection")
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

func cmdCheckout(args []string) {
	debugLog("Starting checkout command")

	if len(args) < 2 {
		fmt.Println("Usage: gitGo checkout <branch>")
		os.Exit(1)
	}

	branch := args[1]
	debugLog("Checkout branch: %s", branch)

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

	debugLog("Checking out branch: %s", branch)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	})

	if err != nil {
		debugLog("Error checking out: %v", err)
		fmt.Printf("Error checking out branch: %v\n", err)
		fmt.Println("")
		fmt.Println("Possible solutions:")
		fmt.Println("1. Make sure the branch exists")
		fmt.Println("2. Commit or stash your changes before checkout")
		os.Exit(1)
	}

	debugLog("Checkout completed successfully")
	fmt.Printf("Switched to branch '%s'\n", branch)
}

func cmdBranch(args []string) {
	debugLog("Starting branch command")

	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	refs, err := repo.References()
	if err != nil {
		debugLog("Error getting references: %v", err)
		fmt.Printf("Error getting references: %v\n", err)
		os.Exit(1)
	}
	debugLog("References retrieved successfully")

	headRef, err := repo.Head()
	if err != nil {
		debugLog("Error getting HEAD: %v", err)
		fmt.Printf("Error getting HEAD: %v\n", err)
		os.Exit(1)
	}
	currentBranch := headRef.Name().Short()
	debugLog("Current branch: %s", currentBranch)

	fmt.Println("Local branches:")
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branchName := ref.Name().Short()
			if branchName == currentBranch {
				fmt.Printf("* %s\n", branchName)
			} else {
				fmt.Printf("  %s\n", branchName)
			}
		}
		return nil
	})

	if err != nil {
		debugLog("Error iterating references: %v", err)
		fmt.Printf("Error iterating references: %v\n", err)
		os.Exit(1)
	}
	debugLog("Branch listing completed successfully")
}

func cmdPush(args []string) {
	debugLog("Starting push command")

	pushFlag := flag.NewFlagSet("push", flag.ExitOnError)
	sshKey := pushFlag.String("ssh-key", "", "Path to SSH private key (default: ~/.ssh/id_rsa)")
	pushFlag.Parse(args[1:])

	repo, err := git.PlainOpen(".")
	if err != nil {
		debugLog("Error opening repository: %v", err)
		fmt.Printf("Error opening repository: %v\n", err)
		os.Exit(1)
	}
	debugLog("Repository opened successfully")

	remote, err := repo.Remote("origin")
	if err != nil {
		debugLog("Error getting remote origin: %v", err)
		fmt.Printf("Error: Remote 'origin' not found\n")
		fmt.Println("Use 'gitGo remote add origin <url>' to add a remote")
		os.Exit(1)
	}
	debugLog("Remote origin found")

	remoteConfig := remote.Config()
	url := remoteConfig.URLs[0]
	debugLog("Remote URL: %s", url)

	var publicKeys ssh.AuthMethod
	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		debugLog("Detected SSH URL, setting up authentication")
		keyPath := *sshKey
		if keyPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				os.Exit(1)
			}
			keyPath = filepath.Join(home, ".ssh", "id_rsa")
			debugLog("Using default SSH key path: %s", keyPath)
		}

		_, err := os.Stat(keyPath)
		if err != nil {
			fmt.Printf("Error: SSH key not found at %s\n", keyPath)
			fmt.Println("Please provide SSH key path with --ssh-key flag")
			os.Exit(1)
		}

		publicKeys, err = ssh.NewPublicKeysFromFile("git", keyPath, "")
		if err != nil {
			debugLog("Error creating SSH public keys: %v", err)
			fmt.Printf("Error loading SSH key: %v\n", err)
			os.Exit(1)
		}
	}

	debugLog("Pushing to remote: %s", url)
	pushOpts := &git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	}
	if publicKeys != nil {
		pushOpts.Auth = publicKeys
	}

	err = repo.Push(pushOpts)

	if err == nil {
		debugLog("Push completed successfully")
		fmt.Println("Push completed successfully")
		return
	}

	if err == git.NoErrAlreadyUpToDate {
		debugLog("Already up to date")
		fmt.Println("Everything up-to-date")
		return
	}

	debugLog("Error pushing: %v", err)
	fmt.Printf("Error pushing: %v\n", err)
	fmt.Println("")
	fmt.Println("Possible solutions:")
	fmt.Println("1. Check if you have write access to the remote repository")
	fmt.Println("2. Make sure your SSH key is correctly configured")
	fmt.Println("3. For HTTPS URLs, ensure credentials are properly set up")
	fmt.Println("4. Check your network connection")
	os.Exit(1)
}

func cmdTest(args []string) {
	debugLog("Starting test command")

	testFlag := flag.NewFlagSet("test", flag.ExitOnError)
	verbose := testFlag.Bool("v", false, "Verbose output")
	race := testFlag.Bool("race", false, "Enable race detector")
	cover := testFlag.Bool("cover", false, "Enable coverage")
	testFlag.Parse(args[1:])

	debugLog("Verbose: %v", *verbose)
	debugLog("Race: %v", *race)
	debugLog("Cover: %v", *cover)

	fmt.Println("Running tests...")

	cmdArgs := []string{"test", "./..."}
	if *verbose {
		cmdArgs = append(cmdArgs, "-v")
	}
	if *race {
		cmdArgs = append(cmdArgs, "-race")
	}
	if *cover {
		cmdArgs = append(cmdArgs, "-cover")
	}

	debugLog("Executing: go %s", strings.Join(cmdArgs, " "))

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		debugLog("Test failed: %v", err)
		fmt.Printf("\nTests failed: %v\n", err)
		os.Exit(1)
	}

	debugLog("Tests completed successfully")
	fmt.Println("\nAll tests passed!")
}