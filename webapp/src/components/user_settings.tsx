// Copyright (c) 2023-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';

import manifest from '@/manifest';
import type { PluginCustomSettingComponent } from '@mattermost/types/plugins/user_settings';

const NotificationServicesSettings: PluginCustomSettingComponent = ({ informChange }) => {
    const [services, setServices] = useState<string[]>([]);
    const [currentService, setCurrentService] = useState<string>('');
    const userPreferences = useSelector((state: any) => state.entities.preferences.myPreferences);
    const savedServices = (userPreferences[`pp_com.mattermost.plugin-shoutrr--notification_services`] || {}).value || '';

    useEffect(() => {
        if (savedServices) {
            setServices(savedServices.split(',').map((s: string) => s.trim()).filter(Boolean));
        }
    }, []);

    const handleAddService = () => {
        if (currentService && !services.includes(currentService)) {
            const newServices = [...services, currentService];
            setServices(newServices);
            informChange('notification_services', newServices.join(','));
            setCurrentService('');
        }
    };

    const handleRemoveService = (serviceToRemove: string) => {
        const newServices = services.filter(service => service !== serviceToRemove);
        setServices(newServices);
        informChange('notification_services', newServices.join(','));
    };

    return (
        <div className='form-group'>
            <div className='input-group' style={{display: 'flex'}}>
                <input
                    type='text'
                    className='form-control'
                    style={{marginRight: '10px'}}
                    placeholder='Enter notification service URL'
                    value={currentService}
                    onChange={(e) => setCurrentService(e.target.value)}
                />
                <div className='input-group-append'>
                    <button
                        className='btn btn-primary'
                        type='button'
                        onClick={handleAddService}
                        disabled={!currentService}
                    >
                        Add
                    </button>
                </div>
            </div>
            
            <details className='mt-2 mb-3'>
                <summary className='cursor-pointer text-blue-600' style={{cursor: 'pointer'}}>
                    <i className='fa fa-chevron-right mr-2' style={{fontSize: '0.8em', transition: 'transform 0.2s', display: 'inline-block'}}></i>
                    Service URL Format Help
                </summary>
                <style>{`
                    details[open] > summary > i.fa-chevron-right {
                        transform: rotate(90deg);
                    }
                `}</style>
                <div className='mt-2 p-3 border rounded bg-white'>
                    <p className='mb-2'>Enter service URLs using one of these formats:</p>
                    <div style={{maxHeight: '250px', overflowY: 'auto'}}>
                        <ul className='list-unstyled pl-2'>
                            <li className='mb-1'><strong>Bark:</strong> <code>bark://<b>devicekey</b>@<b>host</b></code></li>
                            <li className='mb-1'><strong>Discord:</strong> <code>discord://<b>token</b>@<b>id</b></code></li>
                            <li className='mb-1'><strong>Email:</strong> <code>smtp://<b>username</b>:<b>password</b>@<b>host</b>:<b>port</b>/?from=<b>fromAddress</b>&to=<b>recipient1</b>[,<b>recipient2</b>,...]</code></li>
                            <li className='mb-1'><strong>Gotify:</strong> <code>gotify://<b>gotify-host</b>/<b>token</b></code></li>
                            <li className='mb-1'><strong>Google Chat:</strong> <code>googlechat://chat.googleapis.com/v1/spaces/FOO/messages?key=bar&token=baz</code></li>
                            <li className='mb-1'><strong>IFTTT:</strong> <code>ifttt://<b>key</b>/?events=<b>event1</b>[,<b>event2</b>,...]&value1=<b>value1</b>&value2=<b>value2</b>&value3=<b>value3</b></code></li>
                            <li className='mb-1'><strong>Join:</strong> <code>join://shoutrrr:<b>api-key</b>@join/?devices=<b>device1</b>[,<b>device2</b>, ...][&icon=<b>icon</b>][&title=<b>title</b>]</code></li>
                            <li className='mb-1'><strong>Mattermost:</strong> <code>mattermost://[<b>username</b>@]<b>mattermost-host</b>/<b>token</b>[/<b>channel</b>]</code></li>
                            <li className='mb-1'><strong>Matrix:</strong> <code>matrix://<b>username</b>:<b>password</b>@<b>host</b>:<b>port</b>/[?rooms=<b>!roomID1</b>[,<b>roomAlias2</b>]]</code></li>
                            <li className='mb-1'><strong>Ntfy:</strong> <code>ntfy://<b>username</b>:<b>password</b>@ntfy.sh/<b>topic</b></code></li>
                            <li className='mb-1'><strong>OpsGenie:</strong> <code>opsgenie://<b>host</b>/token?responders=<b>responder1</b>[,<b>responder2</b>]</code></li>
                            <li className='mb-1'><strong>Pushbullet:</strong> <code>pushbullet://<b>api-token</b>[/<b>device</b>/#<b>channel</b>/<b>email</b>]</code></li>
                            <li className='mb-1'><strong>Pushover:</strong> <code>pushover://shoutrrr:<b>apiToken</b>@<b>userKey</b>/?devices=<b>device1</b>[,<b>device2</b>, ...]</code></li>
                            <li className='mb-1'><strong>Rocketchat:</strong> <code>rocketchat://[<b>username</b>@]<b>rocketchat-host</b>/<b>token</b>[/<b>channel</b>|<b>@recipient</b>]</code></li>
                            <li className='mb-1'><strong>Slack:</strong> <code>slack://[<b>botname</b>@]<b>token-a</b>/<b>token-b</b>/<b>token-c</b></code></li>
                            <li className='mb-1'><strong>Teams:</strong> <code>teams://<b>group</b>@<b>tenant</b>/<b>altId</b>/<b>groupOwner</b>?host=<b>organization</b>.webhook.office.com</code></li>
                            <li className='mb-1'><strong>Telegram:</strong> <code>telegram://<b>token</b>@telegram?chats=<b>@channel-1</b>[,<b>chat-id-1</b>,...]</code></li>
                            <li className='mb-1'><strong>Zulip Chat:</strong> <code>zulip://<b>bot-mail</b>:<b>bot-key</b>@<b>zulip-domain</b>/?stream=<b>name-or-id</b>&topic=<b>name</b></code></li>
                        </ul>
                    </div>
                    <p className='mt-2 mb-0 text-muted small'>For more details, visit <a href="https://containrrr.dev/shoutrrr/v0.8/services/overview/" target="_blank" rel="noopener noreferrer">Shoutrrr documentation</a>.</p>
                </div>
            </details>

            {services.length > 0 && (
                <div className='mt-3'>
                    <label>
                        Configured Services:
                    </label>
                    <ul className='list-group'>
                        {services.map((service, index) => (
                            <li key={index} className='list-group-item d-flex justify-content-between align-items-center'>
                                {service}
                                <button
                                    className='btn btn-sm btn-danger'
                                    style={{position: 'absolute', right: '10px'}}
                                    onClick={() => handleRemoveService(service)}
                                >
                                    Remove
                                </button>
                            </li>
                        ))}
                    </ul>
                </div>
            )}
        </div>
    );
};

export default NotificationServicesSettings;
