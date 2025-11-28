/*
 * Copyright 2025 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/Nerzal/gocloak/v13"
	"github.com/SENERGY-Platform/analytics-cleanup/lib"
)

type KeycloakService struct {
	client       *gocloak.GoCloak
	token        *gocloak.JWT
	clientId     string
	clientSecret string
	realm        string
	userName     string
	password     string
	url          string
}

func NewKeycloakService(url string, clientId string, clientSecret string, realm string, userName string, password string) *KeycloakService {
	client := gocloak.NewClient(url + "/auth")
	return &KeycloakService{client, nil, clientId, clientSecret, realm, userName, password, url}
}

func (k *KeycloakService) Login() {
	ctx := context.Background()
	token, err := k.client.Login(ctx, k.clientId, k.clientSecret, k.realm, k.userName, k.password)
	if err != nil {
		fmt.Println("Login failed:" + err.Error())
	}
	k.token = token
}

func (k *KeycloakService) Logout() {
	ctx := context.Background()
	err := k.client.Logout(ctx, k.clientId, k.clientSecret, k.realm, k.token.RefreshToken)
	if err != nil {
		fmt.Println("Logout failed:" + err.Error())
	}
}

func (k *KeycloakService) GetAccessToken() string {
	return k.token.AccessToken
}

func (k *KeycloakService) GetUserInfo() (*gocloak.UserInfo, error) {
	ctx := context.Background()
	user, err := k.client.GetUserInfo(ctx, k.token.AccessToken, k.realm)
	return user, err
}

func (k *KeycloakService) GetUserByID(id string) (user *gocloak.User, err error) {
	ctx := context.Background()
	user, err = k.client.GetUserByID(ctx, k.token.AccessToken, k.realm, id)
	return
}

func (k *KeycloakService) GetImpersonateToken(userId string) (token string, err error) {
	resp, err := http.PostForm(k.url+"/auth/realms/"+k.realm+"/protocol/openid-connect/token", url.Values{
		"client_id":         {k.clientId},
		"client_secret":     {k.clientSecret},
		"grant_type":        {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"requested_subject": {userId},
	})
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: GetUserToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		return "", resp.Body.Close()
	}
	var openIdToken lib.OpenIdToken
	err = json.NewDecoder(resp.Body).Decode(&openIdToken)
	if err != nil {
		return
	}
	return openIdToken.AccessToken, nil
}
