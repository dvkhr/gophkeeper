// Factory создаёт клиент с проверкой мастер-пароля и восстановлением сессии
package client

import "github.com/dvkhr/gophkeeper/client/session"

type Factory struct {
	sessionMgr    *session.Manager
	authenticator *Authenticator
	serverAddress string
}

func NewFactory(sessionMgr *session.Manager, authenticator *Authenticator, serverAddress string) *Factory {
	return &Factory{
		sessionMgr:    sessionMgr,
		authenticator: authenticator,
		serverAddress: serverAddress,
	}
}

// NewAuthenticatedClient создаёт клиент с проверкой пароля и восстановлением токенов
func (f *Factory) NewAuthenticatedClient() (*Client, error) {
	if ok, _ := f.sessionMgr.IsAuthenticated(); !ok {
		return nil, ErrUnauthorized
	}

	key, err := f.authenticator.Authenticate()
	if err != nil {
		return nil, err
	}

	client, err := NewClient(f.serverAddress, key)
	if err != nil {
		return nil, err
	}

	sess, _ := f.sessionMgr.Load()
	if sess.AccessToken != "" && sess.RefreshToken != "" {
		_ = client.SetToken(sess.AccessToken, sess.RefreshToken)
	}

	return client, nil
}
