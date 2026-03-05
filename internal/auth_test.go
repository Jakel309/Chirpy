package internal

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTCreationAndValidation (t *testing.T) {
	userID := uuid.Max
	tokenSecret := "mysecret"
	expiresIn := 4000000000000

	token, err := MakeJWT(userID, tokenSecret, time.Duration(expiresIn))

	if err != nil {
		t.Errorf(`Failure to generate token: %v`, err)
		return
	}

	validatedToken, err := ValidateJWT(token, tokenSecret)

	if validatedToken != userID || err != nil {
		t.Errorf(`Expected user id %v to match user id %v. Error follows: %v`, validatedToken, userID, err)
		return
	}
}

func TestRetreivingBearerToken (t *testing.T) {
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	bearer, err := GetBearerToken(headers)

	if bearer != "" || err.Error() != "No bearer token" {
		t.Errorf(`Expected empty bearer but got %v. Expected error message 'No bearer token' but got: %v`, bearer, err)
	}

	token := "dsfgsfdgrfgfdger3234"
	headers.Add("Authorization", "Bearer " + token)

	bearer, err = GetBearerToken(headers)

	if bearer != token || err != nil {
		t.Errorf(`Expected bearer to be %v but got %v. Error message: %v`, token, bearer, err)
	}
}