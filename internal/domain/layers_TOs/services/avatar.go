package services_transfer

import "github.com/google/uuid"

type UploadAvatarInfo struct {
	UserId uuid.UUID `validate:"required,uuid"`
	Image  []byte    `validate:"required"`
}

type AvatarResult struct {
	UserId     uuid.UUID
	Avatar     string
	AvatarMini string
}
