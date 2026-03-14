package sync

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/michaeltukdev/Potok/internal/client"
	"github.com/michaeltukdev/Potok/internal/crypto"
)

type EventHandler struct {
	vaultName string
	vaultRoot string
	client    *client.Client
	encKey    []byte
	logger    *slog.Logger
}

func NewEventHandler(
	vaultName string,
	vaultRoot string,
	c *client.Client,
	encKey []byte,
	logger *slog.Logger,
) *EventHandler {
	return &EventHandler{
		vaultName: vaultName,
		vaultRoot: vaultRoot,
		client:    c,
		encKey:    encKey,
		logger:    logger,
	}
}

func (h *EventHandler) HandleBatch(batch []FileEvent) {
	for _, ev := range batch {
		switch ev.Type {
		case EventCreate, EventUpdate:
			h.handleUpload(ev)
		case EventDelete, EventRename:
			h.handleDelete(ev)
		}
	}
}

func (h *EventHandler) handleUpload(ev FileEvent) {
	plaintext, err := os.ReadFile(ev.AbsPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.logger.Debug("file gone before upload, skipping",
				"file", ev.RelPath,
			)
			return
		}
		h.logger.Error("read file failed",
			"file", ev.RelPath, "err", err,
		)
		return
	}

	ciphertext, err := crypto.EncryptBytes(h.encKey, plaintext)
	if err != nil {
		h.logger.Error("encrypt failed",
			"file", ev.RelPath, "err", err,
		)
		return
	}

	if err := h.client.UploadFile(h.vaultName, ev.RelPath, ciphertext); err != nil {
		h.logger.Error("upload failed",
			"file", ev.RelPath, "err", err,
		)
		return
	}

	h.logger.Info(fmt.Sprintf("[UPLOADED] %s", ev.RelPath))
}

func (h *EventHandler) handleDelete(ev FileEvent) {
	if err := h.client.DeleteFile(h.vaultName, ev.RelPath); err != nil {
		h.logger.Error("remote delete failed",
			"file", ev.RelPath, "err", err,
		)
		return
	}

	h.logger.Info(fmt.Sprintf("[DELETED] %s", ev.RelPath))
}
