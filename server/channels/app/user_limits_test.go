// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	storemocks "github.com/mattermost/mattermost/server/v8/channels/store/storetest/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateActiveWithUserLimits(t *testing.T) {
	mainHelper.Parallel(t)

	t.Run("unlicensed server", func(t *testing.T) {
		t.Run("reactivation allowed below hard limit", func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			th.App.Srv().SetLicense(nil)

			// Deactivate user
			deactivatedUser, appErr := th.App.UpdateActive(th.Context, th.BasicUser, false)
			require.Nil(t, appErr)
			require.NotEqual(t, 0, deactivatedUser.DeleteAt)

			// Reactivate user (should succeed - below hard limit)
			updatedUser, appErr := th.App.UpdateActive(th.Context, th.BasicUser, true)
			require.Nil(t, appErr)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed at previous hard limit", func(t *testing.T) {
			th := SetupWithStoreMock(t)

			th.App.Srv().SetLicense(nil)

			user := &model.User{
				Id:       model.NewId(),
				Email:    "test@example.com",
				Username: "testuser",
				DeleteAt: model.GetMillis(),
			}

			// Mock user count at hard limit
			mockUserStore := storemocks.UserStore{}
			mockUserStore.On("Count", mock.Anything).Return(int64(5000), nil)
			mockUserStore.On("Update", mock.Anything, mock.Anything, true).Return(&model.UserUpdate{New: user}, nil)
			mockTeamStore := storemocks.TeamStore{}
			mockTeamStore.On("GetTeamsByUserId", user.Id).Return([]*model.Team{}, nil)
			mockStore := th.App.Srv().Store().(*storemocks.Store)
			mockStore.On("User").Return(&mockUserStore)
			mockStore.On("Team").Return(&mockTeamStore)

			// Reactivation should still succeed when user limits are disabled.
			updatedUser, appErr := th.App.UpdateActive(th.Context, user, true)
			require.Nil(t, appErr)
			require.NotNil(t, updatedUser)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed above previous hard limit", func(t *testing.T) {
			th := SetupWithStoreMock(t)

			th.App.Srv().SetLicense(nil)

			user := &model.User{
				Id:       model.NewId(),
				Email:    "test@example.com",
				Username: "testuser",
				DeleteAt: model.GetMillis(),
			}

			// Mock user count to exceed hard limit
			mockUserStore := storemocks.UserStore{}
			mockUserStore.On("Count", mock.Anything).Return(int64(6000), nil)
			mockUserStore.On("Update", mock.Anything, mock.Anything, true).Return(&model.UserUpdate{New: user}, nil)
			mockTeamStore := storemocks.TeamStore{}
			mockTeamStore.On("GetTeamsByUserId", user.Id).Return([]*model.Team{}, nil)
			mockStore := th.App.Srv().Store().(*storemocks.Store)
			mockStore.On("User").Return(&mockUserStore)
			mockStore.On("Team").Return(&mockTeamStore)

			// Reactivation should still succeed when user limits are disabled.
			updatedUser, appErr := th.App.UpdateActive(th.Context, user, true)
			require.Nil(t, appErr)
			require.NotNil(t, updatedUser)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})
	})

	t.Run("licensed server with seat count enforcement", func(t *testing.T) {
		t.Run("reactivation allowed below limit", func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			userLimit := 100
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = true
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			// Deactivate user
			_, appErr := th.App.UpdateActive(th.Context, th.BasicUser, false)
			require.Nil(t, appErr)

			// Reactivate user (should succeed - below limit)
			updatedUser, appErr := th.App.UpdateActive(th.Context, th.BasicUser, true)
			require.Nil(t, appErr)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed at previous grace limit", func(t *testing.T) {
			th := SetupWithStoreMock(t)

			userLimit := 100
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = true
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			user := &model.User{
				Id:       model.NewId(),
				Email:    "test@example.com",
				Username: "testuser",
				DeleteAt: model.GetMillis(),
			}

			// Mock user count at grace limit (105 = 100 + 5% grace period)
			mockUserStore := storemocks.UserStore{}
			mockUserStore.On("Count", mock.Anything).Return(int64(105), nil)
			mockUserStore.On("Update", mock.Anything, mock.Anything, true).Return(&model.UserUpdate{New: user}, nil)
			mockTeamStore := storemocks.TeamStore{}
			mockTeamStore.On("GetTeamsByUserId", user.Id).Return([]*model.Team{}, nil)
			mockStore := th.App.Srv().Store().(*storemocks.Store)
			mockStore.On("User").Return(&mockUserStore)
			mockStore.On("Team").Return(&mockTeamStore)

			// Reactivation should still succeed when user limits are disabled.
			updatedUser, appErr := th.App.UpdateActive(th.Context, user, true)
			require.Nil(t, appErr)
			require.NotNil(t, updatedUser)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed at base limit but below grace limit", func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			userLimit := 5 // Grace limit will be 6 (5 + 1 minimum)
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = true
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			// InitBasic creates 3 users, create 2 more to reach base limit of 5
			th.CreateUser(t)
			th.CreateUser(t)

			// Deactivate a user
			_, appErr := th.App.UpdateActive(th.Context, th.BasicUser, false)
			require.Nil(t, appErr)

			// Reactivate user (should succeed - we're at base limit 5 but below grace limit 6)
			updatedUser, appErr := th.App.UpdateActive(th.Context, th.BasicUser, true)
			require.Nil(t, appErr)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed above previous grace limit", func(t *testing.T) {
			th := SetupWithStoreMock(t)

			userLimit := 100
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = true
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			user := &model.User{
				Id:       model.NewId(),
				Email:    "test@example.com",
				Username: "testuser",
				DeleteAt: model.GetMillis(),
			}

			// Mock user count above grace limit (106 > 105 grace limit)
			mockUserStore := storemocks.UserStore{}
			mockUserStore.On("Count", mock.Anything).Return(int64(106), nil)
			mockUserStore.On("Update", mock.Anything, mock.Anything, true).Return(&model.UserUpdate{New: user}, nil)
			mockTeamStore := storemocks.TeamStore{}
			mockTeamStore.On("GetTeamsByUserId", user.Id).Return([]*model.Team{}, nil)
			mockStore := th.App.Srv().Store().(*storemocks.Store)
			mockStore.On("User").Return(&mockUserStore)
			mockStore.On("Team").Return(&mockTeamStore)

			// Reactivation should still succeed when user limits are disabled.
			updatedUser, appErr := th.App.UpdateActive(th.Context, user, true)
			require.Nil(t, appErr)
			require.NotNil(t, updatedUser)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})
	})

	t.Run("licensed server without seat count enforcement", func(t *testing.T) {
		t.Run("reactivation allowed below unenforced limit", func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			userLimit := 5
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = false
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			// Create 2 additional users to have 3 total (below limit of 5)
			th.CreateUser(t)
			th.CreateUser(t)

			// Deactivate user
			_, appErr := th.App.UpdateActive(th.Context, th.BasicUser, false)
			require.Nil(t, appErr)

			// Reactivate user (should succeed - enforcement disabled and below limit)
			updatedUser, appErr := th.App.UpdateActive(th.Context, th.BasicUser, true)
			require.Nil(t, appErr)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed at unenforced limit", func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			userLimit := 5
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = false
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			// Create 4 additional users to have 5 total (at limit of 5)
			th.CreateUser(t)
			th.CreateUser(t)
			th.CreateUser(t)
			th.CreateUser(t)

			// Create a user and then deactivate them
			testUser := th.CreateUser(t)
			_, appErr := th.App.UpdateActive(th.Context, testUser, false)
			require.Nil(t, appErr)

			// Reactivate user (should succeed - enforcement disabled)
			updatedUser, appErr := th.App.UpdateActive(th.Context, testUser, true)
			require.Nil(t, appErr)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})

		t.Run("reactivation allowed above unenforced limit", func(t *testing.T) {
			th := Setup(t).InitBasic(t)

			userLimit := 5
			license := model.NewTestLicense("")
			license.IsSeatCountEnforced = false
			license.Features.Users = &userLimit
			th.App.Srv().SetLicense(license)

			// Create 5 additional users to have 6 total (above limit of 5)
			th.CreateUser(t)
			th.CreateUser(t)
			th.CreateUser(t)
			th.CreateUser(t)
			th.CreateUser(t)

			// Create a user and then deactivate them
			testUser := th.CreateUser(t)
			_, appErr := th.App.UpdateActive(th.Context, testUser, false)
			require.Nil(t, appErr)

			// Reactivate user (should succeed - enforcement disabled)
			updatedUser, appErr := th.App.UpdateActive(th.Context, testUser, true)
			require.Nil(t, appErr)
			require.Equal(t, int64(0), updatedUser.DeleteAt)
		})
	})
}

func TestCreateUserOrGuestSeatCountEnforcement(t *testing.T) {
	mainHelper.Parallel(t)

	t.Run("seat count enforced - allows user creation when under limit", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		userLimit := 5
		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = true
		license.Features.Users = &userLimit
		th.App.Srv().SetLicense(license)

		// InitBasic creates 3 users, so we're under the limit of 5
		user := &model.User{
			Email:         "TestCreateUserOrGuest@example.com",
			Username:      "username_123",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, false)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "username_123", createdUser.Username)
	})

	t.Run("seat count enforced - allows user creation at previous limit", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		userLimit := 5
		extraUsers := 1
		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = true
		license.Features.Users = &userLimit
		license.ExtraUsers = &extraUsers
		th.App.Srv().SetLicense(license)

		// Create 3 additional users to reach the hard limit of 6 (3 from InitBasic + 3)
		// Hard limit = 5 base users + 1 extra user = 6 total
		th.CreateUser(t)
		th.CreateUser(t)
		th.CreateUser(t)

		// Limit enforcement is disabled, so user creation should still succeed.
		user := &model.User{
			Email:         "TestSeatCount@example.com",
			Username:      "seat_test_user",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, false)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_user", createdUser.Username)
	})

	t.Run("seat count enforced - allows user creation over previous limit", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		userLimit := 5
		extraUsers := 0

		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = true
		license.Features.Users = &userLimit
		license.ExtraUsers = &extraUsers
		th.App.Srv().SetLicense(license)

		// Go above the previous hard limit (5).
		th.CreateUser(t)
		th.CreateUser(t)
		th.CreateUser(t)

		user := &model.User{
			Email:         "TestSeatCount@example.com",
			Username:      "seat_test_user",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, false)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_user", createdUser.Username)
	})

	t.Run("seat count not enforced - allows user creation even when over limit", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		userLimit := 5
		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = false
		license.Features.Users = &userLimit
		th.App.Srv().SetLicense(license)

		// Create additional users to exceed the limit (3 from InitBasic + 3 = 6, over limit of 5)
		th.CreateUser(t)
		th.CreateUser(t)
		th.CreateUser(t)

		// Should still allow creation since enforcement is disabled
		user := &model.User{
			Email:         "TestSeatCount@example.com",
			Username:      "seat_test_user",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, false)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_user", createdUser.Username)
	})

	t.Run("no license - uses existing hard limit logic", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		th.App.Srv().SetLicense(nil)

		// Should allow creation under hard limit
		user := &model.User{
			Email:         "TestSeatCount@example.com",
			Username:      "seat_test_user",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, false)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_user", createdUser.Username)
	})

	t.Run("license without Users feature - no seat count enforcement", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = true
		license.Features.Users = nil
		th.App.Srv().SetLicense(license)

		// Should allow creation since Users feature is nil
		user := &model.User{
			Email:         "TestSeatCount@example.com",
			Username:      "seat_test_user",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, false)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_user", createdUser.Username)
	})

	t.Run("guest creation with seat count enforcement - allows at previous limit", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		userLimit := 5
		extraUsers := 1
		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = true
		license.Features.Users = &userLimit
		license.ExtraUsers = &extraUsers
		th.App.Srv().SetLicense(license)

		// Create 3 additional users to reach the hard limit of 6 (3 from InitBasic + 3)
		// Hard limit = 5 base users + 1 extra user = 6 total
		th.CreateUser(t)
		th.CreateUser(t)
		th.CreateUser(t)

		// Limit enforcement is disabled, so guest creation should still succeed.
		user := &model.User{
			Email:         "TestSeatCountGuest@example.com",
			Username:      "seat_test_guest",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, true)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_guest", createdUser.Username)
	})

	t.Run("guest creation with seat count enforcement - allows when under limit", func(t *testing.T) {
		th := Setup(t).InitBasic(t)

		userLimit := 5
		extraUsers := 0
		license := model.NewTestLicense("")
		license.IsSeatCountEnforced = true
		license.Features.Users = &userLimit
		license.ExtraUsers = &extraUsers
		th.App.Srv().SetLicense(license)

		// InitBasic creates 3 users, so we're under the limit of 5
		user := &model.User{
			Email:         "TestSeatCountGuest@example.com",
			Username:      "seat_test_guest",
			Password:      model.NewTestPassword(),
			EmailVerified: true,
		}

		createdUser, appErr := th.App.createUserOrGuest(th.Context, user, true)
		require.Nil(t, appErr)
		require.NotNil(t, createdUser)
		require.Equal(t, "seat_test_guest", createdUser.Username)
	})
}
