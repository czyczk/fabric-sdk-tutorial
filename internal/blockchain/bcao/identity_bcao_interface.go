package bcao

import "gitee.com/czyczk/fabric-sdk-tutorial/pkg/models/identity"

type IIdentityBCAO interface {
	GetDepartmentIdentity() (*identity.DepartmentIdentityStored, error)
}
