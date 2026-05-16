// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from "react";

import {
    fireEvent,
    renderWithContext,
    screen,
} from "tests/react_testing_utils";
import { TestHelper } from "utils/test_helper";

import type { CallButtonAction } from "types/store/plugins";

import CallButton from "./call_button";

describe("plugins/call_button/CallButton", () => {
    const currentChannel = TestHelper.getChannelMock({ id: "channel1" });
    const channelMember = TestHelper.getChannelMembershipMock({
        channel_id: "channel1",
        user_id: "user1",
    });

    const makeCallAction = () => jest.fn();

    const makeSingleAction = (
        overrides: Partial<CallButtonAction> = {}
    ): CallButtonAction => ({
        id: "call-action-1",
        pluginId: "com.mattermost.calls",
        button: React.createElement("button", {
            "data-testid": "plugin-call-button",
        }),
        dropdownButton: React.createElement("button", {
            "data-testid": "plugin-dropdown-button",
        }),
        action: makeCallAction(),
        ...overrides,
    });

    test("should render nothing when there are no call actions", () => {
        const { container } = renderWithContext(
            <CallButton
                pluginCallComponents={[]}
                currentChannel={currentChannel}
                channelMember={channelMember}
                sidebarOpen={false}
            />
        );

        expect(container).toBeEmptyDOMElement();
    });

    test("should render a provided single plugin call button", () => {
        const action = makeSingleAction();

        renderWithContext(
            <CallButton
                pluginCallComponents={[action]}
                currentChannel={currentChannel}
                channelMember={channelMember}
                sidebarOpen={false}
            />
        );

        expect(screen.getByTestId("plugin-call-button")).toBeInTheDocument();
    });

    test("should invoke action when clicking provided plugin button", () => {
        const onCall = makeCallAction();

        renderWithContext(
            <CallButton
                pluginCallComponents={[makeSingleAction({ action: onCall })]}
                currentChannel={currentChannel}
                channelMember={channelMember}
                sidebarOpen={false}
            />
        );

        fireEvent.click(screen.getByTestId("plugin-call-button"));
        expect(onCall).toHaveBeenCalledTimes(1);
        expect(onCall).toHaveBeenCalledWith(currentChannel, channelMember);
    });

    test("should render fallback call button when plugin button is missing", () => {
        renderWithContext(
            <CallButton
                pluginCallComponents={[
                    makeSingleAction({
                        button: null as unknown as JSX.Element,
                    }),
                ]}
                currentChannel={currentChannel}
                channelMember={channelMember}
                sidebarOpen={false}
            />
        );

        expect(
            screen.getByRole("button", { name: /call/i })
        ).toBeInTheDocument();
    });

    test("should invoke action when clicking fallback call button", () => {
        const onCall = makeCallAction();

        renderWithContext(
            <CallButton
                pluginCallComponents={[
                    makeSingleAction({
                        button: null as unknown as JSX.Element,
                        action: onCall,
                    }),
                ]}
                currentChannel={currentChannel}
                channelMember={channelMember}
                sidebarOpen={false}
            />
        );

        fireEvent.click(screen.getByRole("button", { name: /call/i }));
        expect(onCall).toHaveBeenCalledTimes(1);
        expect(onCall).toHaveBeenCalledWith(currentChannel, channelMember);
    });

    test("should render dropdown call button for multiple call actions", () => {
        renderWithContext(
            <CallButton
                pluginCallComponents={[
                    makeSingleAction(),
                    makeSingleAction({
                        id: "call-action-2",
                        pluginId: "com.example.calls",
                    }),
                ]}
                currentChannel={currentChannel}
                channelMember={channelMember}
                sidebarOpen={false}
            />
        );

        expect(screen.getByRole("button", { name: /call/i })).toHaveClass(
            "dropdown"
        );
    });
});
