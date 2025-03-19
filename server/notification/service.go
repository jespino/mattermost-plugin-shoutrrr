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
	}
}

// SendUserNotification sends a notification to a user based on their configured services
func (s *Service) SendUserNotification(userID, message string) error {
	// Get database connection
	db, err := s.client.Store.GetMasterDB()
	if err != nil {
		s.client.Log.Error("Failed to get database connection", "userId", userID, "error", err)
		return fmt.Errorf("failed to get database connection")
	}

	// Query the database directly for user preferences
	query := `
		SELECT Value
		FROM Preferences
		WHERE UserId = $1
		AND Category = 'pp_com.mattermost.plugin-shoutrr'
		AND Name = 'notification_services'
	`
	var servicesStr string
	err = db.QueryRow(query, userID).Scan(&servicesStr)
	if err != nil {
		// If no rows found, it's not an error, just no services configured
		if err.Error() == "sql: no rows in result set" {
			s.client.Log.Debug("No notification services configured for user", "userId", userID)
			return nil
		}
		s.client.Log.Error("Failed to query user preferences from database", "userId", userID, "error", err)
		return fmt.Errorf("failed to query user preferences: %w", err)
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

		err := shoutrrr.Send(serviceURL, message)
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
