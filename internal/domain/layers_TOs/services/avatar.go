package services_transfer

import "github.com/google/uuid"

type UploadAvatarInfo struct {
	UserId uuid.UUID
	Image  []byte
}

type AvatarResult struct {
	UserId     uuid.UUID
	Avatar     string
	AvatarMini string
}
