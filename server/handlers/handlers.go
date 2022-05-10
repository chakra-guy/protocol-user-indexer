package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tamas-soos/wallet-explorer/store"
)

type Handler struct {
	store *store.Store
}

func New(store *store.Store) *Handler {
	return &Handler{
		store: store,
	}
}

func (h *Handler) ListProtocols(c *fiber.Ctx) error {
	pp, err := h.store.GetProtocols()
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"message": err,
			"data":    nil,
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"data": pp,
	})
}

func (h *Handler) ListProtocolsByAddress(c *fiber.Ctx) error {
	pp, err := h.store.GetProtocolsByAddress(c.Params("address"))
	if err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"message": err,
			"data":    nil,
		})
	}

	return c.Status(200).JSON(&fiber.Map{
		"data": pp,
	})
}
