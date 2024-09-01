package validators

import (
	"errors"
	"fmt"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/lib"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// todo: нужно будет в файл локализации вынести
const (
	reqTag   validationTag = "required"
	minTag   validationTag = "min"
	maxTag   validationTag = "max"
	emailTag validationTag = "email"
	alphaTag validationTag = "alpha"
	uuidTag  validationTag = "uuid"
	gteTag   validationTag = "gte"
	lteTag   validationTag = "lte"
)

const (
	requiredErrMessage = "{0} is a required field"
	minErrMessage      = "{0} must be greater than {1}"
	maxErrMessage      = "{0} must be less than {1}"
	emailErrMessage    = "{0} must be a valid email address"
	alphaErrMessage    = "{0} must be a valid alpha numeric value"
	uuidErrMessage     = "{0} must be a valid uuid value"
	gteErrMessage      = "{0} must be greater than {1}"
	lteErrMessage      = "{0} must be less than {1}"
)

var (
	messagesWithParams = map[validationTag]string{
		minTag: minErrMessage,
		maxTag: maxErrMessage,
		gteTag: gteErrMessage,
		lteTag: lteErrMessage,
	}
	messagesWithoutParams = map[validationTag]string{
		emailTag: emailErrMessage,
		alphaTag: alphaErrMessage,
		uuidTag:  uuidErrMessage,
		reqTag:   requiredErrMessage,
	}
)

type validationTag string

var (
	ErrNotFoundTranslator = errors.New("uni.GetTranslator('en') failed")
)

type Validator struct {
	*validator.Validate
}

func NewValidator() (*Validator, error) {
	valid := validator.New()
	uni := ut.New(en.New(), en.New())
	trans, ok := uni.GetTranslator("en")
	if !ok {
		return nil, ErrNotFoundTranslator
	}

	if err := valid.RegisterValidation("mapkeys-user-update", validateMapKeys([]servicestransfer.UserFieldTarget{
		servicestransfer.UserLNameUpdateTarget,
		servicestransfer.UserEmailUpdateTarget,
		servicestransfer.UserFNameUpdateTarget,
	})); err != nil {
		return nil, errors.New(fmt.Sprint("Error registering validation:", err))
	}

	for key, val := range messagesWithParams {
		if err := registerTranslation(translationInfo{
			v:          valid,
			trans:      trans,
			message:    val,
			tag:        key,
			withParams: true,
		}); err != nil {
			return nil, errors.New(fmt.Sprint("Error registering translation:", err))
		}
	}

	for key, val := range messagesWithoutParams {
		if err := registerTranslation(translationInfo{
			v:          valid,
			trans:      trans,
			tag:        key,
			withParams: false,
			message:    val,
		}); err != nil {
			return nil, errors.New(fmt.Sprint("Error registering translation:", err))
		}
	}

	return &Validator{valid}, nil
}

type translationInfo struct {
	v          *validator.Validate
	trans      ut.Translator
	message    string
	tag        validationTag
	withParams bool
}

func registerTranslation(info translationInfo) error {
	return info.v.RegisterTranslation(string(info.tag), info.trans, func(ut ut.Translator) error {
		return ut.Add(info.tag, info.message, true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		if info.withParams {
			t, _ := ut.T(info.tag, fe.Field(), fe.Param())
			return t
		}
		t, _ := ut.T(info.tag, fe.Field())
		return t
	})
}

func validateMapKeys[T ~string](validKeys []T) func(level validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		fields, ok := fl.Field().Interface().(map[T]any)
		if !ok {
			return false
		}

		for key := range fields {
			if !lib.Contains(validKeys, key) {
				return false
			}
		}
		return true
	}
}
