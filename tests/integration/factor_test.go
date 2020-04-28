/*
 * Copyright 2018 - Present Okta, Inc.
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

package integration

import (
	"context"
	"testing"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/okta/okta-sdk-golang/v2/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_exercise_factor_lifecycle(t *testing.T) {
	client, _ := tests.NewClient()

	user, _, err := client.User.GetUser(context.TODO(), "john-factor-lifecycle@example.com")
	if user != nil {
		client.User.DeactivateUser(context.TODO(), user.Id, nil)
		client.User.DeactivateOrDeleteUser(context.TODO(), user.Id, nil)
	}

	p := &okta.PasswordCredential{
		Value: "Abcd1234",
	}
	uc := &okta.UserCredentials{
		Password: p,
	}
	profile := okta.UserProfile{}
	profile["firstName"] = "John"
	profile["lastName"] = "Factor-Lifecycle"
	profile["email"] = "john-factor-lifecycle@example.com"
	profile["login"] = "john-factor-lifecycle@example.com"
	u := &okta.CreateUserRequest{
		Credentials: uc,
		Profile:     &profile,
	}
	qp := query.NewQueryParams(query.WithActivate(false))

	user, _, err = client.User.CreateUser(context.TODO(), *u, qp)
	require.NoError(t, err, "Creating an user should not error")

	allowedFactors, _, _ := client.UserFactor.ListSupportedFactors(context.TODO(), user.Id)
	continueTesting := false
	for _, f := range allowedFactors {
		if f.(*okta.UserFactor).FactorType == "sms" {
			continueTesting = true
		}
	}

	if continueTesting {
		factors, _, listFactorsError := client.UserFactor.ListFactors(context.TODO(), user.Id)
		require.NoError(t, listFactorsError, "Should not error when listing factors")

		assert.Empty(t, factors, "Factors list should be empty")

		factorProfile := okta.NewSmsUserFactorProfile()
		factorProfile.PhoneNumber = "16284001133"

		factor := okta.NewSmsUserFactor()
		factor.Profile = factorProfile

		addedFactor, resp, err := client.UserFactor.EnrollFactor(context.TODO(), user.Id, factor, nil)
		require.NotEmpty(t, resp, "Response should not be empty")
		require.NoError(t, err, "Adding factor should not error")
		assert.IsType(t, okta.NewUserFactor(), addedFactor)

		foundFactor, _, err := client.UserFactor.GetFactor(context.TODO(), user.Id, addedFactor.(*okta.UserFactor).Id, okta.NewSmsUserFactor())
		require.NoError(t, err, "Getting the factor should not error")

		client.UserFactor.DeleteFactor(context.TODO(), user.Id, foundFactor.(*okta.UserFactor).Id)
	} else {
		t.Skip("Skipping exercise factor lifecycle testing. SMS factor was not enabled")
	}

	client.User.DeactivateUser(context.TODO(), user.Id, nil)
	client.User.DeactivateOrDeleteUser(context.TODO(), user.Id, nil)
}
