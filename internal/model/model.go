package model

import "time"

type User struct {
	UID                 string    `json:"uid"`
	Email               string    `json:"email,omitempty"`
	PhoneNumber         string    `json:"phone_number,omitempty"`
	IsPhoneVerified     bool      `json:"isPhoneVerified"`
	IsEmailVerified     bool      `json:"isEmailVerified"`
	IsGuestUser         bool      `json:"isGuestUser"`
	Password            string    `json:"password,omitempty"`
	Joint               []string  `json:"joint"`
	IsBillableUser      bool      `json:"isBillableUser"`
	Is2FNeeded          bool      `json:"is2FNeeded"`
	UserFirstName       string    `json:"userFirstName,omitempty"`
	UserSecondName      string    `json:"userSecondName,omitempty"`
	UserCreatedDate     time.Time `json:"userCreatedDate"`
	UserLastLoginDetail time.Time `json:"userLastLoginDetail"`
	CountryOfOrigin     string    `json:"countryOfOrigin,omitempty"`
	Address             string    `json:"address,omitempty"`
}
