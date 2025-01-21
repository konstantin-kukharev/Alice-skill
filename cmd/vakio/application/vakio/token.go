package vakio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/konstantin-kukharev/Alice-skill/internal"
	"github.com/konstantin-kukharev/Alice-skill/internal/logger"
	"go.uber.org/zap"
)

type (
	Vakio struct {
		cli      *http.Client
		login    string // login
		password string // user password
		cid      string // client id
		secret   string // client secret
		token    Token
		mx       *sync.RWMutex
		log      *logger.Logger
	}

	Token struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expires      int    `json:"expires_in"`
		Type         string `json:"token_type"`
		Scope        any    `json:"scope,omitempty"`
		refreshTimer *time.Timer
	}

	OAuthRequest struct {
		CID       string `json:"client_id"`
		Secret    string `json:"client_secret"`
		GrantType string `json:"grant_type"`
	}

	AuthRequest struct {
		OAuthRequest
		Username string `json:"username"`
		Password string `json:"password"`
	}

	RefreshRequest struct {
		OAuthRequest
		RefreshToken string `json:"refresh_token"`
	}
)

func NewTokenApp(log *logger.Logger, login, cid, secret, password string) *Vakio {
	rt := http.DefaultTransport
	cli := &http.Client{
		Transport: rt,
		Timeout:   internal.HttpClientTimeout,
	}

	return &Vakio{
		log:      log,
		mx:       &sync.RWMutex{},
		cli:      cli,
		login:    login,
		password: password,
		cid:      cid,
		secret:   secret}
}

func (v *Vakio) GetToken() string {
	v.mx.RLock()
	defer v.mx.RUnlock()
	return v.token.AccessToken
}

func (v *Vakio) getRefreshTimer() time.Timer {
	return *v.token.refreshTimer
}

func (v *Vakio) setToken(_ context.Context, t Token) {
	v.mx.Lock()
	v.token = t
	v.token.refreshTimer = time.NewTimer(
		time.Duration(t.Expires)*time.Second -
			internal.HttpClientTimeout)
	v.mx.Unlock()
}

func (v *Vakio) auth(ctx context.Context) error {
	authURL := internal.VakioBaseUrl + "/oauth/token"
	b, err := json.Marshal(AuthRequest{
		OAuthRequest: OAuthRequest{
			CID:       v.cid,
			Secret:    v.secret,
			GrantType: "password",
		},
		Username: v.login,
		Password: v.password,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")
	resp, err := v.cli.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server answer with status %s", resp.Status)
	}

	br, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	token := new(Token)
	err = json.Unmarshal(br, token)
	if err != nil {
		return err
	}

	v.setToken(ctx, *token)

	return nil
}

func (v *Vakio) refresh(ctx context.Context) error {
	authURL := internal.VakioBaseUrl + "/oauth/token"
	b, err := json.Marshal(RefreshRequest{
		OAuthRequest: OAuthRequest{
			CID:       v.cid,
			Secret:    v.secret,
			GrantType: "refresh_token",
		},
		RefreshToken: v.token.RefreshToken,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")
	resp, err := v.cli.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server answer with status %s", resp.Status)
	}

	br, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	token := new(Token)
	err = json.Unmarshal(br, token)
	if err != nil {
		return err
	}

	v.setToken(ctx, *token)

	return nil
}

func (v *Vakio) Run(ctx context.Context) error {
	v.log.InfoCtx(ctx, "run vakio token manager")
	err := v.auth(ctx)
	if err != nil {
		v.log.ErrorCtx(ctx, "auth error", zap.Error(err))
		return err
	}

	for {
		select {
		case <-v.getRefreshTimer().C:
			v.log.InfoCtx(ctx, "refresh token event")
			err := v.refresh(ctx)
			if err != nil {
				v.log.ErrorCtx(ctx, "refresh token error", zap.Error(err))
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
