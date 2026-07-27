package apiroutes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const UnmatchedEndpoint = "unmatched"

const (
	httpTimeout     = 15 * time.Second
	retryInterval   = 30 * time.Second
	refreshInterval = time.Hour
)

type node struct {
	children map[string]*node
	param    *node
	terminal bool
	template string
}

type Classifier struct {
	l       *zap.Logger
	specURL string
	client  *http.Client
	trie    atomic.Pointer[node]
}

func NewClassifier(apiURL string, l *zap.Logger) *Classifier {
	return &Classifier{
		l:       l,
		specURL: strings.TrimRight(apiURL, "/") + "/oapi/v3",
		client:  &http.Client{Timeout: httpTimeout},
	}
}

func (c *Classifier) Match(path string) string {
	root := c.trie.Load()
	if root == nil {
		return UnmatchedEndpoint
	}

	segs := splitPath(path)
	if template, ok := matchNode(root, segs); ok {
		return template
	}
	return UnmatchedEndpoint
}

func (c *Classifier) LoadTemplates(templates []string) {
	c.trie.Store(buildTrie(templates))
}

func (c *Classifier) Run(ctx context.Context) {
	for {
		if err := c.refresh(ctx); err != nil {
			c.l.Warn("failed to load api route spec; requests tagged as unmatched until it loads",
				zap.String("spec_url", c.specURL), zap.Error(err))
		} else {
			break
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(retryInterval):
		}
	}

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.refresh(ctx); err != nil {
				c.l.Warn("failed to refresh api route spec", zap.Error(err))
			}
		}
	}
}

func (c *Classifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.specURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, c.specURL)
	}

	var doc struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return err
	}
	if len(doc.Paths) == 0 {
		return fmt.Errorf("spec %s contained no paths", c.specURL)
	}

	templates := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		templates = append(templates, p)
	}

	c.LoadTemplates(templates)
	c.l.Info("loaded api route classifier", zap.Int("routes", len(templates)))
	return nil
}

func buildTrie(templates []string) *node {
	root := &node{children: map[string]*node{}}
	for _, tmpl := range templates {
		n := root
		for _, s := range splitPath(tmpl) {
			if isParam(s) {
				if n.param == nil {
					n.param = &node{children: map[string]*node{}}
				}
				n = n.param
				continue
			}
			child := n.children[s]
			if child == nil {
				child = &node{children: map[string]*node{}}
				n.children[s] = child
			}
			n = child
		}
		n.terminal = true
		n.template = tmpl
	}
	return root
}

func matchNode(n *node, segs []string) (string, bool) {
	if len(segs) == 0 {
		if n.terminal {
			return n.template, true
		}
		return "", false
	}

	seg, rest := segs[0], segs[1:]
	if child := n.children[seg]; child != nil {
		if template, ok := matchNode(child, rest); ok {
			return template, true
		}
	}
	if n.param != nil {
		if template, ok := matchNode(n.param, rest); ok {
			return template, true
		}
	}
	return "", false
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func isParam(s string) bool {
	return strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")
}
