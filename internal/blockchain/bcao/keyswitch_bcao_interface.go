package bcao

import "gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/keyswitch"

type IKeySwitchBCAO interface {
	CreateKeySwitchTrigger(ksTrigger *keyswitch.KeySwitchTrigger, eventID ...string) (string, error)
	CreateKeySwitchResult(ksResult *keyswitch.KeySwitchResult) (string, error)
	GetKeySwitchResult(query *keyswitch.KeySwitchResultQuery) (*keyswitch.KeySwitchResultStored, error)
	ListKeySwitchResultsByID(ksSessionID string) ([]*keyswitch.KeySwitchResultStored, error)
}
