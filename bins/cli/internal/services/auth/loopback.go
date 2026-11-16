package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"golang.org/x/oauth2"

	"github.com/nuonco/nuon/pkg/nuonauth"
)

const cliLoopbackWait = 10 * time.Minute

var errLoopbackUnavailable = errors.New("loopback listener unavailable")

func (a *Service) listenLoopback() (net.Listener, error) {
	if a.listenLoopbackFn != nil {
		return a.listenLoopbackFn()
	}
	return net.Listen("tcp", "127.0.0.1:0")
}

func (a *Service) loginWithLoopback(ctx context.Context, authBaseURL string) (*LoginResult, error) {
	ln, err := a.listenLoopback()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errLoopbackUnavailable, err)
	}

	port := ln.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, nuonauth.CLICallbackPath)
	conf := &oauth2.Config{
		ClientID: nuonauth.CLIClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   authBaseURL + "/oauth/authorize",
			TokenURL:  authBaseURL + "/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectURL,
	}
	verifier := oauth2.GenerateVerifier()
	state, err := newOAuthState()
	if err != nil {
		_ = ln.Close()
		return nil, err
	}
	authURL := conf.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	waitCtx, cancel := context.WithTimeout(ctx, cliLoopbackWait)
	defer cancel()

	exchange := func(ctx context.Context, code string) (*LoginResult, error) {
		tok, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			return nil, fmt.Errorf("failed to complete login: %w", err)
		}
		email, _ := tok.Extra("email").(string)
		if email == "" {
			email = "user"
		}
		return &LoginResult{
			AccessToken: tok.AccessToken,
			DisplayName: email,
		}, nil
	}

	type waitResult struct {
		result *LoginResult
		err    error
	}
	done := make(chan waitResult, 1)
	go func() {
		result, err := serveLoopback(waitCtx, ln, state, authBaseURL+nuonauth.CLICompletePath, exchange)
		done <- waitResult{result: result, err: err}
	}()

	fmt.Println("\nLogging in to Nuon")
	fmt.Println("Opening your browser to approve this login on this computer.")
	fmt.Printf("If the browser does not open, visit:\n\n    %s\n\n", authURL)
	_ = openBrowser(authURL)

	res := <-done
	if errors.Is(res.err, context.DeadlineExceeded) {
		return nil, fmt.Errorf("login timed out - please try again")
	}
	return res.result, res.err
}

type loopbackExchange func(ctx context.Context, code string) (*LoginResult, error)

func newOAuthState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func stateMatches(got, want string) bool {
	if len(got) != len(want) || want == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// serveLoopback accepts one callback, exchanges its code, then redirects the
// browser to the auth server's completion page so the code leaves the address bar.
func serveLoopback(ctx context.Context, ln net.Listener, state, completeURL string, exchange loopbackExchange) (*LoginResult, error) {
	type outcome struct {
		result *LoginResult
		err    error
	}
	doneCh := make(chan outcome, 1)
	var handled atomic.Bool

	redirect := func(w http.ResponseWriter, r *http.Request, result string) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		http.Redirect(w, r, completeURL+"?result="+url.QueryEscape(result), http.StatusFound)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+nuonauth.CLICallbackPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if !stateMatches(q.Get("state"), state) {
			http.Error(w, "login state did not match", http.StatusBadRequest)
			return
		}
		if !handled.CompareAndSwap(false, true) {
			http.Error(w, "login already completed", http.StatusConflict)
			return
		}

		if oauthErr := q.Get("error"); oauthErr != "" {
			result := nuonauth.CLIResultFailed
			if oauthErr == "access_denied" {
				result = nuonauth.CLIResultDenied
			}
			redirect(w, r, result)
			doneCh <- outcome{err: fmt.Errorf("authentication error: %s", oauthErr)}
			return
		}
		code := q.Get("code")
		if code == "" {
			redirect(w, r, nuonauth.CLIResultFailed)
			doneCh <- outcome{err: fmt.Errorf("missing authorization code")}
			return
		}

		result, err := exchange(ctx, code)
		if err != nil {
			redirect(w, r, nuonauth.CLIResultFailed)
			doneCh <- outcome{err: err}
			return
		}
		redirect(w, r, nuonauth.CLIResultSuccess)
		doneCh <- outcome{result: result}
	})

	srv := &http.Server{Handler: mux}
	serveErr := make(chan error, 1)
	go func() {
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-serveErr:
		return nil, err
	case out := <-doneCh:
		return out.result, out.err
	}
}
