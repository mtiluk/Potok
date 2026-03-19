package sync

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/crypto"
)

type EventHandler struct {
	vaultName string
	vaultRoot string
	client    *client.Client
	encKey    []byte
	guard     *PullGuard
	manifest  *Manifest
	logger    *slog.Logger
}

func NewEventHandler(
	vaultName string,
	vaultRoot string,
	c *client.Client,
	encKey []byte,
	guard *PullGuard,
	manifest *Manifest,
	logger *slog.Logger,
) *EventHandler {
	return &EventHandler{
		vaultName: vaultName,
		vaultRoot: vaultRoot,
		client:    c,
		encKey:    encKey,
		guard:     guard,
		manifest:  manifest,
		logger:    logger,
	}
}

func (h *EventHandler) HandleBatch(batch []FileEvent) {
	changed := false
	for _, ev := range batch {
		if h.guard != nil && h.guard.IsLocked(ev.RelPath) {
			h.logger.Debug("skipping guarded file", "file", ev.RelPath)
			continue
		}

		switch ev.Type {
		case EventCreate, EventUpdate:
			if h.handleUpload(ev) {
				changed = true
			}
		case EventDelete, EventRename:
			if h.handleDelete(ev) {
				changed = true
			}
		}
	}

	if changed {
		if err := h.manifest.Save(); err != nil {
			h.logger.Error("save manifest failed", "err", err)
		}
	}
}

func (h *EventHandler) handleUpload(ev FileEvent) bool {
	plaintext, err := os.ReadFile(ev.AbsPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.logger.Debug("file gone before upload, skipping",
				"file", ev.RelPath,
			)
			return false
		}
		h.logger.Error("read file failed",
			"file", ev.RelPath, "err", err,
		)
		return false
	}

	ciphertext, err := crypto.EncryptBytes(h.encKey, plaintext)
	if err != nil {
		h.logger.Error("encrypt failed",
			"file", ev.RelPath, "err", err,
		)
		return false
	}

	if err := h.client.UploadFile(h.vaultName, ev.RelPath, ciphertext); err != nil {
		h.logger.Error("upload failed",
			"file", ev.RelPath, "err", err,
		)
		return false
	}

	h.logger.Info(fmt.Sprintf("[UPLOADED] %s", ev.RelPath))

	h.manifest.Set(ev.RelPath, ManifestEntry{
		Size:    int64(len(ciphertext)),
		ModTime: time.Now().UTC().Format(time.RFC3339),
	})

	return true
}

func (h *EventHandler) handleDelete(ev FileEvent) bool {
	if err := h.client.DeleteFile(h.vaultName, ev.RelPath); err != nil {
		h.logger.Error("remote delete failed",
			"file", ev.RelPath, "err", err,
		)
		return false
	}

	h.logger.Info(fmt.Sprintf("[DELETED] %s", ev.RelPath))

	h.manifest.Remove(ev.RelPath)
	return true
}
