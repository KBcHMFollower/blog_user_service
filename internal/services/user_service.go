package services

import (
	"context"
	"fmt"
	ssov1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/auth"
	usersv1 "github.com/KBcHMFollower/blog_user_service/api/protos/gen/users"
	"github.com/KBcHMFollower/blog_user_service/internal/lib/tokens"

	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	s3client "github.com/KBcHMFollower/blog_user_service/internal/s3"
	"log/slog"
	"time"

	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	log         *slog.Logger
	tokenTtl    time.Duration
	tokenSecret string
	userRep     repository.UserStore
	imgStore    s3client.ImageStore
}

func New(log *slog.Logger, tokenTtl time.Duration, tokenSecret string, userRep repository.UserStore, imgStore s3client.ImageStore) *UserService {
	return &UserService{
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		userRep:     userRep,
		imgStore:    imgStore,
	}
}

func (a *UserService) RegisterUser(ctx context.Context, req *ssov1.RegisterDTO) (string, error) {

	op := "authService/registerUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", req.Email),
	)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		log.Error("can`t generate hashPass: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	userId, err := a.userRep.CreateUser(ctx, &repository.CreateUserDto{
		Email:    req.GetEmail(),
		FName:    req.GetFname(),
		LName:    req.GetLname(),
		HashPass: hashPass,
	})
	if err != nil {
		log.Error("can`t create user in db: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	token, err := tokens_helper.CreateNewJwt(userId, req.GetEmail(), a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	return token, nil
}

func (a *UserService) LoginUser(ctx context.Context, email string, password string) (string, error) {

	op := "authService/loginUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := a.userRep.GetUserByEmail(ctx, email)
	if err != nil {
		log.Error("can`t get user from db: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Error("passwords not eq: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t generate jwt: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	return token, nil
}

func (a *UserService) CheckAuth(ctx context.Context, token string) (string, error) {

	op := "authService/checkAuth"

	log := a.log.With(
		slog.String("op", op),
	)

	parsedToken, err := tokens_helper.Parse(token, a.tokenSecret)
	if err != nil {
		log.Error("can`t parse token: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}
	if !parsedToken.Valid {
		log.Error("token is invalid: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		log.Error("can`t  parse jwt claims: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	return newToken, nil
}

func (a *UserService) GetUserById(ctx context.Context, dto *usersv1.GetUserDTO) (*usersv1.GetUserRDO, error) {
	op := "UserService/GetUserById"

	log := a.log.With(
		slog.String("op", op),
	)

	userUUID, err := uuid.Parse(dto.GetId())
	if err != nil {
		log.Error("can`t parse user uuid: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	user, err := a.userRep.GetUserById(ctx, userUUID)
	if err != nil {
		log.Error("can`t get user from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.GetUserRDO{
		User: user.ConvertToProto(),
	}, nil
}

func (a *UserService) GetSubscribers(ctx context.Context, dto *usersv1.GetSubscribersDTO) (*usersv1.GetSubscribersRDO, error) {
	op := "UserService/GetSubscribers"
	log := a.log.With(
		slog.String("op", op))

	bloggerUUID, err := uuid.Parse(dto.GetBloggerId())
	if err != nil {
		log.Error("can`t parse blogger uuid: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	users, totalCount, err := a.userRep.GetUserSubscribers(ctx, bloggerUUID, uint64(dto.GetPage()), uint64(dto.GetSize()))
	if err != nil {
		log.Error("can`t get user subscribers from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.GetSubscribersRDO{
		Subscribers: models.UsersArrayToProto(users),
		TotalCount:  int32(totalCount),
	}, nil
}

func (a *UserService) GetSubscriptions(ctx context.Context, dto *usersv1.GetSubscriptionsDTO) (*usersv1.GetSubscriptionsRDO, error) {
	op := "UserService/GetSubscriptions"
	log := a.log.With(
		slog.String("op", op))

	subscriberUUID, err := uuid.Parse(dto.GetSubscriberId())
	if err != nil {
		log.Error("can`t parse subscriber uuid: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	users, totalCount, err := a.userRep.GetUserSubscriptions(ctx, subscriberUUID, uint64(dto.GetPage()), uint64(dto.GetSize()))
	if err != nil {
		log.Error("can`t get user bloggers from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.GetSubscriptionsRDO{
		Subscriptions: models.UsersArrayToProto(users),
		TotalCount:    int32(totalCount),
	}, nil
}

func (a *UserService) UpdateUser(ctx context.Context, dto *usersv1.UpdateUserDTO) (*usersv1.UpdateUserRDO, error) {
	op := "UserService/UpdateUser"
	log := a.log.With(
		slog.String("op", op))

	userUUID, err := uuid.Parse(dto.GetId())
	if err != nil {
		log.Error("can`t parse user uuid: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	var updateItems = make([]*repository.UpdateInfo, 0)
	for _, item := range dto.GetUpdateData() {
		updateItems = append(updateItems, &repository.UpdateInfo{
			Name:  item.GetName(),
			Value: item.GetValue(),
		})
	}

	user, err := a.userRep.UpdateUser(ctx, repository.UpdateData{
		Id:         userUUID,
		UpdateInfo: updateItems,
	})
	if err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.UpdateUserRDO{
		User: user.ConvertToProto(),
	}, nil
}

func (a *UserService) Subscribe(ctx context.Context, dto *usersv1.SubscribeDTO) (*usersv1.SubscribeRDO, error) {
	op := "UserService/Subscribe"
	log := a.log.With(
		slog.String("op", op))

	bloggerUUID, err := uuid.Parse(dto.GetBloggerId())
	if err != nil {
		log.Error("can`t parse blogger uuid: ", err)
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	subscriberUUID, err := uuid.Parse(dto.GetSubscriberId())
	if err != nil {
		log.Error("can`t parse subscriber uuid: ", err)
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	err = a.userRep.Subscribe(ctx, bloggerUUID, subscriberUUID)
	if err != nil {
		log.Error("can`t subscribe to user in db: ", err)
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.SubscribeRDO{
		IsSubscribe: true,
	}, nil
}

func (a *UserService) Unsubscribe(ctx context.Context, dto *usersv1.SubscribeDTO) (*usersv1.SubscribeRDO, error) {
	op := "UserService/Subscribe"
	log := a.log.With(
		slog.String("op", op))

	bloggerUUID, err := uuid.Parse(dto.GetBloggerId())
	if err != nil {
		log.Error("can`t parse blogger uuid: ", err)
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	subscriberUUID, err := uuid.Parse(dto.GetSubscriberId())
	if err != nil {
		log.Error("can`t parse subscriber uuid: ", err)
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	err = a.userRep.Unsubscribe(ctx, bloggerUUID, subscriberUUID)
	if err != nil {
		log.Error("can`t unsubscribe in db: ", err)
		return &usersv1.SubscribeRDO{
			IsSubscribe: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.SubscribeRDO{
		IsSubscribe: true,
	}, nil
}

func (a *UserService) DeleteUser(ctx context.Context, dto *usersv1.DeleteUserDTO) (*usersv1.DeleteUserRDO, error) {
	op := "UserService/DeleteUser"
	log := a.log.With(
		slog.String("op", op))

	userUUID, err := uuid.Parse(dto.GetId())
	if err != nil {
		log.Error("can`t parse user uuid: ", err)
		return &usersv1.DeleteUserRDO{
			IsDeleted: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	err = a.userRep.DeleteUser(ctx, userUUID)
	if err != nil {
		log.Error("can`t delete user in db: ", err)
		return &usersv1.DeleteUserRDO{
			IsDeleted: false,
		}, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.DeleteUserRDO{
		IsDeleted: true,
	}, nil
}

func (a *UserService) UploadAvatar(ctx context.Context, dto *usersv1.UploadAvatarDTO) (*usersv1.UploadAvatarRDO, error) {
	op := "UserService/UploadAvatar"
	log := a.log.With(
		slog.String("op", op))

	userUUID, err := uuid.Parse(dto.GetUserId())
	if err != nil {
		log.Error("can`t parse user uuid: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	imgUrl, err := a.imgStore.UploadFile(ctx, fmt.Sprintf("%s.jpeg", uuid.New().String()), dto.GetImage(), s3client.ImageJpeg)
	if err != nil {
		log.Error("can`t upload image: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	//TODO
	_, err = a.userRep.UpdateUser(ctx, repository.UpdateData{
		Id: userUUID,
		UpdateInfo: []*repository.UpdateInfo{
			{
				Name:  "avatar",
				Value: imgUrl,
			},
			{
				Name:  "avatar_min",
				Value: imgUrl,
			},
		},
	})
	if err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &usersv1.UploadAvatarRDO{
		UserId:        dto.GetUserId(),
		AvatarUrl:     imgUrl,
		AvatarMiniUrl: imgUrl,
	}, nil
}
