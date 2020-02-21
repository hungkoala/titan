package titan

import (
	"github.com/google/uuid"
)

type UUID string

func (u UUID) String() string {
	return string(u)
}

type UserInfo struct {
	ExternalUserId  UUID                   `json:"externalUserId"` // as of now , it's for role patient only
	UserId          UUID                   `json:"userId"`
	CareProviderId  UUID                   `json:"careProviderId"`
	CareProviderKey string                 `json:"careProviderKey"`
	DeviceId        string                 `json:"deviceId"` // uuid format
	Role            Role                   `json:"role"`
	Attributes      map[string]interface{} `json:"attributes"`
}

func (userInfo UserInfo) CareProviderUUID() *uuid.UUID {
	uuid, _ := uuid.Parse(userInfo.CareProviderId.String())
	return &uuid
}
func (userInfo UserInfo) UserUUID() *uuid.UUID {
	uuid, _ := uuid.Parse(userInfo.UserId.String())
	return &uuid
}

type Role string

func (r Role) String() string {
	return string(r)
}

type String struct {
	Value string
}

func (u *UserInfo) GetSubject() string {
	if u == nil {
		return ""
	}
	if val, ok := u.Attributes["sub"]; ok {
		return val.(string)
	}
	return ""
}
