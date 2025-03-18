package notification

import (
	"fmt"
	"strings"

	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/router"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

// Service handles sending notifications to different services through Shoutrrr
type Service struct {
	client *pluginapi.Client
	router router.ServiceRouter
}

// NewService creates a new notification service
func NewService(client *pluginapi.Client) *Service {
	return &Service{
		client: client,
		router: shoutrrr.CreateSender(),
	}
}

// SendUserNotification sends a notification to a user based on their configured services
func (s *Service) SendUserNotification(userID, message string) error {
	// Get user preferences directly from the database
	preferences, err := s.client.Preference.GetForUser(userID)
	if err != nil {
		s.client.Log.Error("Failed to get user preferences", "userId", userID, "error", err)
		return err
	}

	// Find the notification_services preference
	var servicesStr string
	for _, pref := range preferences {
		if pref.Category == "plugin_com.mattermost.plugin-shoutrrr" && pref.Name == "notification_services" {
			servicesStr = pref.Value
			break
		}
	}

	if servicesStr == "" {
		s.client.Log.Debug("No notification services configured for user", "userId", userID)
		return nil
	}

	// Split into individual services
	services := strings.Split(servicesStr, ",")
	for i, service := range services {
		services[i] = strings.TrimSpace(service)
	}

	var errs []string
	for _, serviceURL := range services {
		if serviceURL == "" {
			continue
		}

		err := s.router.Send(message, serviceURL)
		if err != nil {
			s.client.Log.Error("Failed to send notification",
				"userId", userID,
				"service", serviceURL,
				"error", err)
			errs = append(errs, fmt.Sprintf("%s: %v", serviceURL, err))
		} else {
			s.client.Log.Debug("Notification sent successfully",
				"userId", userID,
				"service", serviceURL)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to send notifications: %s", strings.Join(errs, "; "))
	}
	return nil
}

// SendMentionNotification sends a notification about a mention to a user
func (s *Service) SendMentionNotification(userID, postID, channel, mentionedBy, message string) error {
	notificationMsg := fmt.Sprintf("You were mentioned by @%s in %s: %s",
		mentionedBy, channel, message)

	return s.SendUserNotification(userID, notificationMsg)
}
