// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type {Store, Action} from 'redux';
import React from 'react';

import type {GlobalState} from '@mattermost/types/store';
import type {
    PluginConfiguration,
    PluginConfigurationSection,
    PluginConfigurationCustomSetting
} from '@mattermost/types/plugins/user_settings';

import manifest from '@/manifest';
import type {PluginRegistry} from '@/types/mattermost-webapp';
import NotificationServicesSettings from './components/user_settings';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // Register user settings
        const userSettings: PluginConfiguration = {
            id: manifest.id,
            uiName: 'Notification Services',
            icon: "icon-bell",
            sections: [
                {
                    title: 'Notification Services Settings',
                    settings: [
                        {
                            type: 'custom',
                            name: 'notification_services',
                            title: 'Notification Services',
                            helpText: 'Configure services that will receive notifications. Enter URLs for external services that should be notified.',
                            component: NotificationServicesSettings
                        } as PluginConfigurationCustomSetting
                    ]
                } as PluginConfigurationSection
            ]
        };

        registry.registerUserSettings(userSettings);
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void;
    }
}

window.registerPlugin(manifest.id, new Plugin());
