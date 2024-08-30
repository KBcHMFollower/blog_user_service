package services

import (
	"context"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	tokens_helper "github.com/KBcHMFollower/blog_user_service/internal/lib/tokens"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	services_utils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type userStore interface {
	dep.UserCreator
	dep.UserGetter
}

type AuthService struct {
	userRep     userStore
	log         *slog.Logger
	tokenTtl    time.Duration
	tokenSecret string
	txCreator   dep.TransactionCreator
}

func NewAuthService(userRep userStore, log *slog.Logger, tokenTtl time.Duration, tokenSecret string, txCreator dep.TransactionCreator) *AuthService {
	return &AuthService{
		userRep:     userRep,
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		txCreator:   txCreator,
	}
}

// TODO: fname и lname не работают
func (as *AuthService) Register(ctx context.Context, req *transfer.RegisterInfo) (resToken *transfer.TokenResult, resErr error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionEmailKey, req.Email)

	as.log.InfoContext(ctx, "trying to register user")

	tx, err := as.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t start transaction", err))
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(err, tx)
	}()

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t generate hashPass", err))
	}

	as.log.Debug("hash pass is generated successfully")

	userId, err := as.userRep.Create(ctx, &repositories_transfer.CreateUserInfo{
		Email:    req.Email,
		FName:    req.FName,
		LName:    req.LName,
		HashPass: hashPass,
	}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create user in db", err))
	}

	logger.UpdateLoggerCtx(ctx, createdUserIdLogKey, userId)
	as.log.Debug("user created in db successfully")

	token, err := tokens_helper.CreateNewJwt(userId, req.Email, as.tokenTtl, as.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create new jwt", err))
	}

	logger.UpdateLoggerCtx(ctx, accessTokenLogKey, token)
	as.log.Debug("token created successfully")

	as.log.InfoContext(ctx, "user registered successfully")

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (as *AuthService) Login(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionEmailKey, loginInfo.Email)

	as.log.InfoContext(ctx, "user try to login")

	user, err := as.userRep.User(ctx, repositories_transfer.GetUserInfo{
		Target: repositories_transfer.UserEmailCondition,
		Value:  loginInfo.Email,
	}, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user from db", err))
	}

	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, user.Id)
	as.log.Debug("email is exists")

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(loginInfo.Password))
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("passwords not eq", err))
	}

	as.log.Debug("password is correct")

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, as.tokenTtl, as.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t generate jwt", err))
	}

	as.log.Debug("token created successfully", "as-token", token)
	as.log.InfoContext(ctx, "user logged in successfully")

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (as *AuthService) CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error) {
	parsedToken, err := tokens_helper.Parse(authInfo.AccessToken, as.tokenSecret)

	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t parse token", err))
	}
	if !parsedToken.Valid {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("token is invalid", err))
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t  parse jwt claims", err))
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, as.tokenTtl, as.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create jwt", err))
	}

	return &transfer.TokenResult{
		AccessToken: newToken,
	}, nil
}
