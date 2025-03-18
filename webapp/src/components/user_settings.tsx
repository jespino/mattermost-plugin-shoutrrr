// Copyright (c) 2023-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import { FormattedMessage } from 'react-intl';

import manifest from '@/manifest';
import type { PluginCustomSettingComponent } from '@mattermost/types/plugins/user_settings';

const NotificationServicesSettings: PluginCustomSettingComponent = ({ informChange }) => {
    const [services, setServices] = useState<string[]>([]);
    const [currentService, setCurrentService] = useState<string>('');
    const userPreferences = useSelector((state: any) => state.entities.preferences.myPreferences);
    const savedServices = (userPreferences[`${manifest.id}--notification_services`] || {}).value || '';
    
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
            <div className='input-group'>
                <input
                    type='text'
                    className='form-control'
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
                        <FormattedMessage defaultMessage='Add' />
                    </button>
                </div>
            </div>
            
            {services.length > 0 && (
                <div className='mt-3'>
                    <label>
                        <FormattedMessage defaultMessage='Configured Services:' />
                    </label>
                    <ul className='list-group'>
                        {services.map((service, index) => (
                            <li key={index} className='list-group-item d-flex justify-content-between align-items-center'>
                                {service}
                                <button
                                    className='btn btn-sm btn-danger'
                                    onClick={() => handleRemoveService(service)}
                                >
                                    <FormattedMessage defaultMessage='Remove' />
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
