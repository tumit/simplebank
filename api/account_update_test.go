package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "tumit.ga/simplebank/db/mock"
	db "tumit.ga/simplebank/db/sqlc"
)

func TestUpdateAccount(t *testing.T) {
	account := randomAccount()

	testCases := []struct {
		name          string
		accountID     int64
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recoder *httptest.ResponseRecorder)
	}{
		{
			name:      "InvalidID",
			accountID: 0,
			body: gin.H{
				"balance": account.Balance,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InvalidBalance",
			accountID: account.ID,
			body: gin.H{
				"balance": -1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "UpdateError",
			accountID: account.ID,
			body: gin.H{
				"balance": account.Balance,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "OK",
			accountID: account.ID,
			body: gin.H{
				"balance": account.Balance,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateAccountParams{
					ID:      account.ID,
					Balance: account.Balance,
				}
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})
	}
}
