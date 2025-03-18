// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import type { PluginConfiguration } from '@mattermost/types/plugins/user_settings';

export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType);
    registerUserSettings(settings: PluginConfiguration): void;

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}
