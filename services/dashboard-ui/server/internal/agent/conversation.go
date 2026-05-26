package agent

import (
	"sync"
	"time"
)

const maxMessages = 100

type Conversation struct {
	ID        string    `json:"id"`
	OrgID     string    `json:"org_id"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	mu        sync.Mutex
}

func (c *Conversation) AppendMessage(msg Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Messages = append(c.Messages, msg)
	if len(c.Messages) > maxMessages {
		c.Messages = c.Messages[len(c.Messages)-maxMessages:]
	}
	c.UpdatedAt = time.Now()
}

type ConversationStore struct {
	mu    sync.RWMutex
	convs map[string]*Conversation
}

func NewConversationStore() *ConversationStore {
	return &ConversationStore{
		convs: make(map[string]*Conversation),
	}
}

func (s *ConversationStore) Create(id, orgID string) *Conversation {
	now := time.Now()
	c := &Conversation{
		ID:        id,
		OrgID:     orgID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.mu.Lock()
	s.convs[id] = c
	s.mu.Unlock()
	return c
}

func (s *ConversationStore) Get(id string) *Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.convs[id]
}

func (s *ConversationStore) ListByOrg(orgID string) []*Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*Conversation
	for _, c := range s.convs {
		if c.OrgID == orgID {
			out = append(out, c)
		}
	}
	return out
}

func (s *ConversationStore) Delete(id string) {
	s.mu.Lock()
	delete(s.convs, id)
	s.mu.Unlock()
}
