// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';
import {useIntl} from 'react-intl';
import {useSelector} from 'react-redux';

import {WithTooltip} from '@mattermost/shared/components/tooltip';

import {getChannelByName} from 'mattermost-redux/selectors/entities/channels';

import {
    isCallsEnabled as getIsCallsEnabled,
    getSessionsInCalls,
    callsChannelExplicitlyDisabled,
} from 'selectors/calls';

import ProfilePopoverCallButton from 'components/profile_popover/profile_popover_calls_button';

import {getDirectChannelName} from 'utils/utils';

import type {GlobalState} from 'types/store';

type Props = {
    userId: string;
    currentUserId: string;
    fullname: string;
    username: string;
};

export function isUserInCall(
    state: GlobalState,
    userId: string,
    channelId: string,
) {
    const sessionsInCall = getSessionsInCalls(state)[channelId] || {};

    for (const session of Object.values(sessionsInCall)) {
        if (session.user_id === userId) {
            return true;
        }
    }

    return false;
}

const CallButton = ({userId, currentUserId, fullname, username}: Props) => {
    const {formatMessage} = useIntl();

    const isCallsEnabled = useSelector((state: GlobalState) =>
        getIsCallsEnabled(state),
    );
    const dmChannel = useSelector((state: GlobalState) =>
        getChannelByName(state, getDirectChannelName(currentUserId, userId)),
    );
    const isCallsPluginEnabledInState = useSelector((state: GlobalState) => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        return state['plugins-com.mattermost.calls']?.enabled !== false;
    });

    const shouldRenderButton = useSelector((state: GlobalState) => {
        // 1. No one should get the button if the plugin is disabled.
        if (!isCallsEnabled || !isCallsPluginEnabledInState) {
            return false;
        }

        // 2. No one should get the button if calls in channel have been explicitly disabled in the DM channel.
        if (callsChannelExplicitlyDisabled(state, dmChannel?.id ?? '')) {
            return false;
        }

        return true;
    });

    const hasDMCall = useSelector((state: GlobalState) => {
        if (isCallsEnabled && dmChannel) {
            return (
                isUserInCall(state, currentUserId, dmChannel.id) ||
                isUserInCall(state, userId, dmChannel.id)
            );
        }
        return false;
    });

    if (!shouldRenderButton) {
        return null;
    }

    // We disable the button if there's already a call ongoing with the user.
    const disabled = hasDMCall;
    const startCallMessage = hasDMCall ? formatMessage(
        {
            id: 'user_profile.call.ongoing',
            defaultMessage: 'Call with {user} is ongoing',
        },
        {user: fullname || username},
    ) : formatMessage({
        id: 'user_profile.call.start',
        defaultMessage: 'Start Call',
    });
    const callButton = (
        <WithTooltip title={startCallMessage}>
            <button
                id='startCallButton'
                type='button'
                disabled={disabled}
                className='btn btn-icon btn-sm style--none'
                aria-label={startCallMessage}
            >
                <span
                    className='icon icon-phone'
                    aria-hidden='true'
                />
            </button>
        </WithTooltip>
    );

    if (disabled) {
        return callButton;
    }

    return (
        <ProfilePopoverCallButton
            dmChannel={dmChannel}
            userId={userId}
            customButton={callButton}
        />
    );
};

export default CallButton;
