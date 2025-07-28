// Authenticator запрос и проверка пароля
package client

import (
	"bytes"
	"fmt"

	"github.com/dvkhr/gophkeeper/client/internal/utils"
	"github.com/dvkhr/gophkeeper/client/session"
	"github.com/dvkhr/gophkeeper/pkg/crypto"
)

type Authenticator struct {
	sessionMgr *session.Manager
}

func NewAuthenticator(sessionMgr *session.Manager) *Authenticator {
	return &Authenticator{sessionMgr: sessionMgr}
}

// Authenticate запрашивает мастер-пароль и проверяет его
func (a *Authenticator) Authenticate() ([]byte, error) {
	sess, err := a.sessionMgr.Load()
	if err != nil {
		return nil, err
	}

	if len(sess.Salt) == 0 {
		return nil, ErrNoSalt
	}

	if len(sess.MasterKeyHash) == 0 {
		return nil, ErrNoMasterKeyHash
	}

	password, err := utils.ReadMasterPassword("Master-пароль: ")
	if err != nil {
		return nil, err
	}
	defer utils.ZeroBytes(password)

	key := crypto.DeriveKey(string(password), sess.Salt)
	hash := crypto.SHA256(key)

	if !bytes.Equal(hash, sess.MasterKeyHash) {
		utils.ZeroBytes(password)
		return nil, ErrInvalidPassword
	}

	return key, nil
}

var (
	ErrNoSalt          = fmt.Errorf("соль не найдена")
	ErrNoMasterKeyHash = fmt.Errorf("master_key_hash не найден")
	ErrInvalidPassword = fmt.Errorf("неверный мастер-пароль — данные не могут быть зашифрованы")
	ErrUnauthorized    = fmt.Errorf("вы не авторизованы")
)
