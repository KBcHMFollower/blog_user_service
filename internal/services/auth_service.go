package services

import (
	"context"
	"errors"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	tokenshelper "github.com/KBcHMFollower/blog_user_service/internal/lib/tokens"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	servicesutils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type authSvcUserStore interface {
	dep.UserCreator
	dep.UserGetter
}

type AuthService struct {
	userRep     authSvcUserStore
	log         logger.Logger
	tokenTtl    time.Duration
	tokenSecret string
	txCreator   dep.TransactionCreator
}

func NewAuthService(userRep authSvcUserStore, log logger.Logger, tokenTtl time.Duration, tokenSecret string, txCreator dep.TransactionCreator) *AuthService {
	return &AuthService{
		userRep:     userRep,
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		txCreator:   txCreator,
	}
}

func (as *AuthService) Register(ctx context.Context, req *transfer.RegisterInfo) (resToken *transfer.TokenResult, resErr error) {
	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionEmailKey, req.Email)
	as.log.InfoContext(ctx, "trying to register user")

	tx, err := as.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t start transaction", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(err, tx)
	}()

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t generate hashPass", err))
	}

	as.log.DebugContext(ctx, "hash pass is generated successfully")

	userId, err := as.userRep.Create(ctx, &repositoriestransfer.CreateUserInfo{
		Email:    req.Email,
		FName:    req.FName,
		LName:    req.LName,
		HashPass: hashPass,
	}, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create user in db", err))
	}

	ctx = logger.UpdateLoggerCtx(ctx, createdUserIdLogKey, userId)
	as.log.DebugContext(ctx, "user created in db successfully")

	token, err := tokenshelper.CreateNewJwt(userId, req.Email, as.tokenTtl, as.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create new jwt", err))
	}

	ctx = logger.UpdateLoggerCtx(ctx, accessTokenLogKey, token)
	as.log.DebugContext(ctx, "token created successfully")

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t commit transaction", err))
	}

	as.log.InfoContext(ctx, "user registered successfully")

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (as *AuthService) Login(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error) {
	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionEmailKey, loginInfo.Email)

	as.log.InfoContext(ctx, "user try to login")

	user, err := as.userRep.User(ctx, repositoriestransfer.GetUserInfo{
		Condition: map[repositoriestransfer.UserFieldTarget]interface{}{
			repositoriestransfer.UserEmailCondition: loginInfo.Email,
		},
	}, nil)
	if err != nil {
		if errors.Is(err, ctxerrors.ErrNotFound) {
			return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("email not found", ctxerrors.ErrBadRequest))
		}
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user from db", err))
	}

	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, user.Id)
	as.log.DebugContext(ctx, "email is exists")

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(loginInfo.Password))
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("passwords not eq", ctxerrors.ErrBadRequest))
	}

	as.log.DebugContext(ctx, "password is correct")

	token, err := tokenshelper.CreateNewJwt(user.Id, user.Email, as.tokenTtl, as.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t generate jwt", err))
	}

	as.log.DebugContext(ctx, "token created successfully", "as-token", token)
	as.log.InfoContext(ctx, "user logged in successfully")

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (as *AuthService) CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error) {
	parsedToken, err := tokenshelper.Parse(authInfo.AccessToken, as.tokenSecret)

	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t parse token", ctxerrors.ErrUnauthorized))
	}
	if !parsedToken.Valid {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("token is invalid", ctxerrors.ErrUnauthorized))
	}

	tokenClaims, err := tokenshelper.GetClaimsValues(parsedToken)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t  parse jwt claims", ctxerrors.ErrUnauthorized))
	}

	newToken, err := tokenshelper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, as.tokenTtl, as.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create jwt", err))
	}

	return &transfer.TokenResult{
		AccessToken: newToken,
	}, nil
}
