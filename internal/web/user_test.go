package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectify/internal/domain"
	"connectify/internal/service"
	svcmocks "connectify/internal/service/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Signup(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		reqBuilder func(t *testing.T) *http.Request
		wantCode   int
		wantBody   string
	}{
		{
			name: "registration successful",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "test@example.com",
					Password: "Test@1234",
				}).Return(nil)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "test@example.com",
"password": "Test@1234",
"confirmPassword": "Test@1234"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"message":"registration successful"}`,
		},
		{
			name: "bind error",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "test@example.com",
"password": "Test@1234
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"invalid request"}`,
		},
		{
			name: "email format error",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "invalid-email",
"password": "Test@1234",
"confirmPassword": "Test@1234"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"email format error"}`,
		},
		{
			name: "password format error",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "test@example.com",
"password": "weak",
"confirmPassword": "weak"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"password format error"}`,
		},
		{
			name: "two input passwords are not consistent",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "test@example.com",
"password": "Test@1234",
"confirmPassword": "Test@5678"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"two input passwords are not consistent"}`,
		},
		{
			name: "email conflict",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "test@example.com",
					Password: "Test@1234",
				}).Return(service.ErrUserDuplicateEmail)
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "test@example.com",
"password": "Test@1234",
"confirmPassword": "Test@1234"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusOK,
			wantBody: `{"message":"email conflict"}`,
		},
		{
			name: "system error",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := svcmocks.NewMockUserService(ctrl)
				userSvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "test@example.com",
					Password: "Test@1234",
				}).Return(errors.New("db error"))
				codeSvc := svcmocks.NewMockCodeService(ctrl)
				return userSvc, codeSvc
			},
			reqBuilder: func(t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost,
					"/user/signup", bytes.NewReader([]byte(`{
"email": "test@example.com",
"password": "Test@1234",
"confirmPassword": "Test@1234"
}`)))
				req.Header.Set("Content-Type", "application/json")
				assert.NoError(t, err)
				return req
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"system error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// build handler
			userSvc, codeSvc := tc.mock(ctrl)
			handler := NewUserHandler(userSvc, codeSvc)

			// prepare server, register routes
			gin.SetMode(gin.TestMode)
			server := gin.New()
			handler.RegisterRouter(server)

			// prepare request and record the recorder
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()

			// execute
			server.ServeHTTP(recorder, req)

			// assert result
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.JSONEq(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestValidateEmail(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool
	}{
		{
			name:  "no @",
			email: "123456",
			match: false,
		},
		{
			name:  "with @ but no suffix",
			email: "123456@",
			match: false,
		},
		{
			name:  "valid email",
			email: "123456@qq.com",
			match: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := ValidateEmail(tc.email)
			assert.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		match    bool
	}{
		{
			name:     "too short",
			password: "Aa1@",
			match:    false,
		},
		{
			name:     "no uppercase letter",
			password: "test@1234",
			match:    false,
		},
		{
			name:     "no lowercase letter",
			password: "TEST@1234",
			match:    false,
		},
		{
			name:     "no number",
			password: "Test@abcd",
			match:    false,
		},
		{
			name:     "no special character",
			password: "Test12345",
			match:    false,
		},
		{
			name:     "valid password",
			password: "Test@1234",
			match:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := ValidatePassword(tc.password)
			assert.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}
