package keycloak

import "time"

type token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"-"`
	RefreshToken string    `json:"refresh_token"`
	SessionState string    `json:"-"`
	ClientID     string    `json:"-"`
}

type token_response struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
	SessionState string `json:"session_state"`
	// error fields
	// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

type User_Credentials struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

type User struct {
	Id               string              `json:"id"`
	CreatedTimestamp int64               `json:"createdTimestamp"`
	Username         string              `json:"username"`
	Enabled          bool                `json:"enabled"`
	Totp             bool                `json:"totp"`
	EmailVerified    bool                `json:"emailVerified"`
	FirstName        string              `json:"firstName"`
	LastName         string              `json:"lastName"`
	Email            string              `json:"email"`
	Attributes       map[string][]string `json:"attributes,omitempty"`
	NotBefore        int                 `json:"notBefore,omitempty"`
	Access           struct {
		ManageGroupMembership bool `json:"manageGroupMembership,omitempty"`
		View                  bool `json:"view,omitempty"`
		MapRoles              bool `json:"mapRoles,omitempty"`
		Impersonate           bool `json:"impersonate,omitempty"`
		Manage                bool `json:"manage,omitempty"`
	} `json:"access,omitempty"`
	Credentials []User_Credentials `json:"credentials,omitempty"`
}

type Role struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Composite   bool   `json:"composite"`
	ClientRole  bool   `json:"clientRole"`
	ContainerId string `json:"containerId"`
}

type ClientRep struct {
	Id       string `json:"id"`
	ClientId string `json:"clientId"`
	Name     string `json:"name"`
}
