package gitops

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

var errStop = errors.New("stop iteration")

func IsGitRepo(dest string) bool {
	_, err := git.PlainOpen(dest)
	return err == nil
}

func DetectDefaultBranch(url string) (string, error) {
	sto := memory.NewStorage()
	repo, err := git.Init(sto, nil)
	if err != nil {
		return "", fmt.Errorf("init memory repo: %w", err)
	}

	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		return "", fmt.Errorf("creating remote: %w", err)
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("listing remote refs: %w", err)
	}

	for _, ref := range refs {
		if ref.Name() == plumbing.HEAD && ref.Type() == plumbing.SymbolicReference {
			target := string(ref.Target())
			return strings.TrimPrefix(target, "refs/heads/"), nil
		}
	}

	return "", nil
}

func CloneRepo(url, dest, branch string, sparsePaths []string) error {
	refName := plumbing.ReferenceName("refs/heads/" + branch)

	if len(sparsePaths) > 0 {
		repo, err := git.PlainClone(dest, false, &git.CloneOptions{
			URL:           url,
			NoCheckout:    true,
			SingleBranch:  true,
			ReferenceName: refName,
		})
		if err != nil {
			return fmt.Errorf("clone (no-checkout): %w", err)
		}

		wt, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("worktree: %w", err)
		}

		return wt.Checkout(&git.CheckoutOptions{
			Branch:                    refName,
			SparseCheckoutDirectories: sparsePaths,
		})
	}

	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:           url,
		SingleBranch:  true,
		ReferenceName: refName,
	})
	if err != nil {
		return fmt.Errorf("clone: %w", err)
	}
	return nil
}

func PullRepo(dest, branch string, sparsePaths []string) error {
	repo, err := git.PlainOpen(dest)
	if err != nil {
		return fmt.Errorf("opening repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	if err := wt.Reset(&git.ResetOptions{Mode: git.HardReset}); err != nil {
		return fmt.Errorf("resetting worktree: %w", err)
	}

	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/heads/" + branch + ":refs/remotes/origin/" + branch),
		},
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("fetch: %w", err)
	}

	remoteRef, err := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+branch), true)
	if err != nil {
		return fmt.Errorf("resolving remote ref: %w", err)
	}

	if err := wt.Reset(&git.ResetOptions{
		Mode:   git.HardReset,
		Commit: remoteRef.Hash(),
	}); err != nil {
		return fmt.Errorf("resetting to remote: %w", err)
	}

	if len(sparsePaths) > 0 {
		return wt.Checkout(&git.CheckoutOptions{
			Hash:                      remoteRef.Hash(),
			SparseCheckoutDirectories: sparsePaths,
			Force:                     true,
		})
	}

	return nil
}

func GetHEAD(dest string) (plumbing.Hash, error) {
	repo, err := git.PlainOpen(dest)
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("opening repo: %w", err)
	}
	ref, err := repo.Head()
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("resolving HEAD: %w", err)
	}
	return ref.Hash(), nil
}

func CountIncoming(dest, branch string) (int, error) {
	repo, err := git.PlainOpen(dest)
	if err != nil {
		return 0, fmt.Errorf("opening repo: %w", err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return 0, fmt.Errorf("resolving HEAD: %w", err)
	}
	headHash := headRef.Hash()

	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			config.RefSpec("refs/heads/" + branch + ":refs/remotes/origin/" + branch),
		},
	})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return 0, nil
	}

	remoteRef, err := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/"+branch), true)
	if err != nil {
		return 0, nil
	}

	iter, err := repo.Log(&git.LogOptions{
		From: remoteRef.Hash(),
	})
	if err != nil {
		return 0, nil
	}
	defer iter.Close()

	count := 0
	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash == headHash {
			return errStop
		}
		count++
		return nil
	})
	if err != nil && !errors.Is(err, errStop) {
		return 0, nil
	}

	return count, nil
}

func CountChanges(dest string, prevHead plumbing.Hash) (int, error) {
	repo, err := git.PlainOpen(dest)
	if err != nil {
		return 0, fmt.Errorf("opening repo: %w", err)
	}

	headRef, err := repo.Head()
	if err != nil {
		return 0, fmt.Errorf("resolving HEAD: %w", err)
	}

	if prevHead == plumbing.ZeroHash || headRef.Hash() == plumbing.ZeroHash {
		return 0, nil
	}

	fromCommit, err := repo.CommitObject(prevHead)
	if err != nil {
		return 0, nil
	}

	toCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return 0, nil
	}

	patch, err := fromCommit.Patch(toCommit)
	if err != nil {
		return 0, nil
	}

	return len(patch.Stats()), nil
}

func AbsTarget(projectDir, targetDir string) string {
	return filepath.Join(projectDir, targetDir)
}
