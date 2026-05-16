// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from "react";

import type { ChannelType } from "@mattermost/types/channels";

import { renderWithContext, screen } from "tests/react_testing_utils";
import Constants, { RHSStates } from "utils/constants";
import { TestHelper } from "utils/test_helper";

import ChannelHeader from "./channel_header";

jest.mock("plugins/call_button", () => {
    return {
        __esModule: true,
        default: () => <div data-testid="channel-header-call-button" />,
    };
});

describe("components/ChannelHeader call button visibility", () => {
    const baseProps = {
        actions: {
            showPinnedPosts: jest.fn(),
            showChannelFiles: jest.fn(),
            closeRightHandSide: jest.fn(),
            getCustomEmojisInText: jest.fn(),
            updateChannelNotifyProps: jest.fn(),
            showChannelMembers: jest.fn(),
            fetchChannelRemotes: jest.fn(),
        },
        team: TestHelper.getTeamMock({ id: "team_id" }),
        channel: TestHelper.getChannelMock({
            id: "channel_id",
            type: Constants.OPEN_CHANNEL as ChannelType,
        }),
        channelMember: TestHelper.getChannelMembershipMock({
            channel_id: "channel_id",
            user_id: "user_id",
        }),
        currentUser: TestHelper.getUserMock({ id: "user_id" }),
        isCustomStatusEnabled: false,
        isCustomStatusExpired: false,
        isFileAttachmentsEnabled: true,
        lastActivityTimestamp: 0,
        isLastActiveEnabled: false,
        memberCount: 2,
        dmUser: undefined,
        gmMembers: undefined,
        rhsState: RHSStates.CHANNEL_INFO,
        isChannelMuted: false,
        hasGuests: false,
        pinnedPostsCount: 0,
        customStatus: undefined,
        timestampUnits: [],
        hideGuestTags: false,
        remoteNames: [],
        sharedChannelsPluginsEnabled: false,
        isChannelAutotranslated: false,
    };

    test("should render call button in non-shared channel", () => {
        renderWithContext(<ChannelHeader {...baseProps} />);

        expect(
            screen.getByTestId("channel-header-call-button")
        ).toBeInTheDocument();
    });

    test("should render call button in shared channel when shared channel plugins are disabled", () => {
        renderWithContext(
            <ChannelHeader
                {...baseProps}
                channel={TestHelper.getChannelMock({
                    id: "shared_channel_id",
                    type: Constants.OPEN_CHANNEL as ChannelType,
                    shared: true,
                })}
                sharedChannelsPluginsEnabled={false}
            />
        );

        expect(
            screen.getByTestId("channel-header-call-button")
        ).toBeInTheDocument();
    });

    test("should render call button in shared channel when shared channel plugins are enabled", () => {
        renderWithContext(
            <ChannelHeader
                {...baseProps}
                channel={TestHelper.getChannelMock({
                    id: "shared_channel_id",
                    type: Constants.OPEN_CHANNEL as ChannelType,
                    shared: true,
                })}
                sharedChannelsPluginsEnabled={true}
            />
        );

        expect(
            screen.getByTestId("channel-header-call-button")
        ).toBeInTheDocument();
    });
});
