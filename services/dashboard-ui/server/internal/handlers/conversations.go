package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/agent"
)

type ConversationsHandler struct {
	cfg   *internal.Config
	l     *zap.Logger
	agent *agent.Agent
	store *agent.ConversationStore
}

func NewConversationsHandler(cfg *internal.Config, l *zap.Logger, agentInstance *agent.Agent, store *agent.ConversationStore) *ConversationsHandler {
	return &ConversationsHandler{
		cfg:   cfg,
		l:     l,
		agent: agentInstance,
		store: store,
	}
}

func (h *ConversationsHandler) RegisterRoutes(e *gin.Engine) error {
	if !h.cfg.AgentEnabled {
		return nil
	}

	g := e.Group("/api/orgs/:orgId/conversations")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:conversationId", h.Get)
	g.POST("/:conversationId/messages", h.SendMessage)
	g.DELETE("/:conversationId", h.Delete)
	return nil
}

func (h *ConversationsHandler) Create(c *gin.Context) {
	orgID := c.Param("orgId")
	id := uuid.New().String()
	conv := h.store.Create(id, orgID)
	c.JSON(http.StatusCreated, conv)
}

func (h *ConversationsHandler) List(c *gin.Context) {
	orgID := c.Param("orgId")
	convs := h.store.ListByOrg(orgID)
	if convs == nil {
		convs = []*agent.Conversation{}
	}
	c.JSON(http.StatusOK, convs)
}

func (h *ConversationsHandler) Get(c *gin.Context) {
	conv := h.store.Get(c.Param("conversationId"))
	if conv == nil || conv.OrgID != c.Param("orgId") {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	c.JSON(http.StatusOK, conv)
}

type sendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *ConversationsHandler) SendMessage(c *gin.Context) {
	orgID := c.Param("orgId")
	convID := c.Param("conversationId")

	conv := h.store.Get(convID)
	if conv == nil || conv.OrgID != orgID {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	token, err := c.Cookie(authCookie)
	if err != nil || token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conv.AppendMessage(agent.Message{
		Role:    "user",
		Content: req.Content,
	})

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ctx := c.Request.Context()
	sseOut := make(chan agent.SSEEvent, 64)

	go h.agent.Run(ctx, agent.RunOpts{
		Conversation: conv,
		Token:        token,
		OrgID:        orgID,
	}, sseOut)

	for ev := range sseOut {
		select {
		case <-ctx.Done():
			return
		default:
		}
		fmt.Fprint(c.Writer, agent.FormatSSE(ev.Event, ev.Data))
		c.Writer.Flush()
	}
}

func (h *ConversationsHandler) Delete(c *gin.Context) {
	convID := c.Param("conversationId")
	conv := h.store.Get(convID)
	if conv == nil || conv.OrgID != c.Param("orgId") {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	h.store.Delete(convID)
	c.Status(http.StatusNoContent)
}
