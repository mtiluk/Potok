package sync

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	gosync "sync"
	"time"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/crypto"
)

type PullGuard struct {
	mu    gosync.Mutex
	paths map[string]struct{}
}

func NewPullGuard() *PullGuard {
	return &PullGuard{paths: make(map[string]struct{})}
}

func (g *PullGuard) Lock(relPath string) {
	g.mu.Lock()
	g.paths[relPath] = struct{}{}
	g.mu.Unlock()
}

func (g *PullGuard) Unlock(relPath string) {
	g.mu.Lock()
	delete(g.paths, relPath)
	g.mu.Unlock()
}

func (g *PullGuard) IsLocked(relPath string) bool {
	g.mu.Lock()
	_, ok := g.paths[relPath]
	g.mu.Unlock()
	return ok
}

type Poller struct {
	vaultName string
	vaultRoot string
	client    *client.Client
	encKey    []byte
	guard     *PullGuard
	logger    *slog.Logger
	manifest  *Manifest
	failed    map[string]struct{}
}

func NewPoller(
	vaultName string,
	vaultRoot string,
	c *client.Client,
	encKey []byte,
	guard *PullGuard,
	logger *slog.Logger,
) (*Poller, error) {
	manifest, err := LoadManifest(vaultName)
	if err != nil {
		return nil, fmt.Errorf("load manifest: %w", err)
	}

	return &Poller{
		vaultName: vaultName,
		vaultRoot: vaultRoot,
		client:    c,
		encKey:    encKey,
		guard:     guard,
		logger:    logger,
		manifest:  manifest,
		failed:    make(map[string]struct{}),
	}, nil
}

func (p *Poller) Poll() {
	remote, err := p.client.GetManifest(p.vaultName)
	if err != nil {
		p.logger.Error("poll: get manifest failed", "err", err)
		return
	}

	for relPath := range remote {
		if isPotokInternal(relPath) {
			delete(remote, relPath)
		}
	}

	localFiles := make(map[string]struct{})
	_ = filepath.WalkDir(p.vaultRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if ignoreDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(p.vaultRoot, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		localFiles[rel] = struct{}{}
		return nil
	})

	changed := false

	for relPath, remoteInfo := range remote {
		if _, skip := p.failed[relPath]; skip {
			continue
		}

		lastKnown, tracked := p.manifest.Get(relPath)
		_, existsLocally := localFiles[relPath]

		needsPull := !existsLocally ||
			(tracked && (lastKnown.Size != remoteInfo.Size ||
				lastKnown.ModTime != remoteInfo.ModTime)) ||
			!tracked

		if !tracked && existsLocally {
			p.manifest.Set(relPath, ManifestEntry{
				Size:    remoteInfo.Size,
				ModTime: remoteInfo.ModTime,
			})
			changed = true
			continue
		}

		if needsPull {
			if p.pullFile(relPath) {
				p.manifest.Set(relPath, ManifestEntry{
					Size:    remoteInfo.Size,
					ModTime: remoteInfo.ModTime,
				})
				changed = true
			}
		}
	}

	for relPath := range p.manifest.Entries {
		if _, exists := remote[relPath]; !exists {
			absPath := filepath.Join(
				p.vaultRoot, filepath.FromSlash(relPath),
			)

			p.guard.Lock(relPath)

			if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
				p.logger.Error("poll: local delete failed",
					"file", relPath, "err", err,
				)
				p.guard.Unlock(relPath)
				continue
			}

			p.logger.Info(fmt.Sprintf("[PULL-DELETE] %s", relPath))
			p.manifest.Remove(relPath)
			changed = true

			go func(rel string) {
				time.Sleep(1 * time.Second)
				p.guard.Unlock(rel)
			}(relPath)

			dir := filepath.Dir(absPath)
			vaultAbs, _ := filepath.Abs(p.vaultRoot)
			for {
				dirAbs, _ := filepath.Abs(dir)
				if dirAbs == vaultAbs {
					break
				}
				entries, err := os.ReadDir(dir)
				if err != nil || len(entries) > 0 {
					break
				}
				os.Remove(dir)
				dir = filepath.Dir(dir)
			}
		}
	}

	if changed {
		if err := p.manifest.Save(); err != nil {
			p.logger.Error("poll: save manifest failed", "err", err)
		}
	}
}

func (p *Poller) pullFile(relPath string) bool {
	ciphertext, err := p.client.DownloadFile(p.vaultName, relPath)
	if err != nil {
		p.logger.Error("poll: download failed", "file", relPath, "err", err)
		return false
	}

	plaintext, err := crypto.DecryptBytes(p.encKey, ciphertext)
	if err != nil {
		p.logger.Warn("poll: decrypt failed, skipping on future polls",
			"file", relPath, "err", err,
		)
		p.failed[relPath] = struct{}{}
		return false
	}

	absPath := filepath.Join(p.vaultRoot, filepath.FromSlash(relPath))

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		p.logger.Error("poll: create dir failed", "file", relPath, "err", err)
		return false
	}

	p.guard.Lock(relPath)

	if err := os.WriteFile(absPath, plaintext, 0644); err != nil {
		p.logger.Error("poll: write failed", "file", relPath, "err", err)
		p.guard.Unlock(relPath)
		return false
	}

	p.logger.Info(fmt.Sprintf("[PULLED] %s", relPath))

	go func(rel string) {
		time.Sleep(1 * time.Second)
		p.guard.Unlock(rel)
	}(relPath)

	return true
}

func isPotokInternal(relPath string) bool {
	return strings.HasPrefix(relPath, ".potok/") || relPath == ".potok"
}
