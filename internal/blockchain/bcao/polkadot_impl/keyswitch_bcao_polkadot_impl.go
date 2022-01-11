package polkadot

import (
	"net/http"
	"time"

	"gitee.com/czyczk/fabric-sdk-tutorial/internal/blockchain/chaincodectx"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/errorcode"
	"gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"
)

type KeySwitchBCAOPolkadotImpl struct {
	ctx    *chaincodectx.PolkadotChaincodeCtx
	client *http.Client
}

func NewKeySwitchBCAOPolkadotImpl(ctx *chaincodectx.PolkadotChaincodeCtx) *KeySwitchBCAOPolkadotImpl {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	return &KeySwitchBCAOPolkadotImpl{
		ctx:    ctx,
		client: client,
	}
}

func (o *KeySwitchBCAOPolkadotImpl) CreateKeySwitchTrigger(ksTrigger *keyswitch.KeySwitchTrigger, eventID ...string) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *KeySwitchBCAOPolkadotImpl) CreateKeySwitchResult(ksResult *keyswitch.KeySwitchResult) (string, error) {
	// TODO
	return "", errorcode.ErrorNotImplemented
}

func (o *KeySwitchBCAOPolkadotImpl) GetKeySwitchResult(query *keyswitch.KeySwitchResultQuery) (*keyswitch.KeySwitchResultStored, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}

func (o *KeySwitchBCAOPolkadotImpl) ListKeySwitchResultsByID(ksSessionID string) ([]*keyswitch.KeySwitchResultStored, error) {
	// TODO
	return nil, errorcode.ErrorNotImplemented
}
