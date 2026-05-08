// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"

	"github.com/mattermost/mattermost/server/public/model"
)

func (a *App) GetServerLimits() (*model.ServerLimits, *model.AppError) {
	limits := &model.ServerLimits{}
	limits.MaxUsersLimit = 0
	limits.MaxUsersHardLimit = 0
	limits.PostHistoryLimit = 0
	limits.LastAccessiblePostTime = 0
	limits.SingleChannelGuestLimit = 0

	activeUserCount, appErr := a.Srv().Store().User().Count(model.UserCountOptions{})
	if appErr != nil {
		return nil, model.NewAppError("GetServerLimits", "app.limits.get_app_limits.user_count.store_error", nil, "", http.StatusInternalServerError).Wrap(appErr)
	}

	if a.shouldTrackSingleChannelGuests() {
		singleChannelGuestCount, err := a.Srv().Store().User().AnalyticsGetSingleChannelGuestCount()
		if err != nil {
			return nil, model.NewAppError("GetServerLimits", "app.limits.get_app_limits.single_channel_guest_count.store_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}

		limits.ActiveUserCount = max(activeUserCount-singleChannelGuestCount, 0)
		limits.SingleChannelGuestCount = singleChannelGuestCount
	} else {
		limits.ActiveUserCount = activeUserCount
	}

	return limits, nil
}

func (a *App) shouldTrackSingleChannelGuests() bool {
	license := a.License()
	if license == nil {
		return false
	}
	if license.IsMattermostEntry() {
		return false
	}
	cfg := a.Config()
	if cfg == nil || cfg.GuestAccountsSettings.Enable == nil {
		return false
	}

	return *cfg.GuestAccountsSettings.Enable
}

func (a *App) GetPostHistoryLimit() int64 {
	return 0
}

func (a *App) isAtUserLimit() (bool, *model.AppError) {
	userLimits, appErr := a.GetServerLimits()
	if appErr != nil {
		return false, appErr
	}

	if userLimits.MaxUsersHardLimit == 0 {
		return false, nil
	}

	return userLimits.ActiveUserCount >= userLimits.MaxUsersHardLimit, appErr
}
