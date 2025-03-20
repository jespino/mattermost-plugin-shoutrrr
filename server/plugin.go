package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-shoutrrr/server/command"
	"github.com/mattermost/mattermost-plugin-shoutrrr/server/notification"
	"github.com/mattermost/mattermost-plugin-shoutrrr/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore kvstore.KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// commandClient is the client used to register and execute slash commands.
	commandClient command.Command

	// notificationService is the client used to send notifications using Shoutrrr
	notificationService *notification.Service

	backgroundJob *cluster.Job

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	p.kvstore = kvstore.NewKVStore(p.client)

	p.commandClient = command.NewCommandHandler(p.client)

	// Initialize notification service
	p.notificationService = notification.NewService(p.client)

	job, err := cluster.Schedule(
		p.API,
		"BackgroundJob",
		cluster.MakeWaitForRoundedInterval(1*time.Hour),
		p.runJob,
	)
	if err != nil {
		return errors.Wrap(err, "failed to schedule background job")
	}

	p.backgroundJob = job

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	if p.backgroundJob != nil {
		if err := p.backgroundJob.Close(); err != nil {
			p.API.LogError("Failed to close background job", "err", err)
		}
	}
	return nil
}

// MessageHasBeenPosted is called after a message has been posted.
// This hook extracts all mentions from the post and logs them.
func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	mentions, err := p.GetAllMentions(post)
	if err != nil {
		p.API.LogError("Failed to get mentions from post", "error", err.Error())
		return
	}

	// Log the mentions
	p.API.LogInfo("Message mentions detected",
		"post_id", post.Id,
		"user_mentions", formatMentionsForLog(mentions.Mentions),
		"here_mentioned", mentions.HereMentioned,
		"channel_mentioned", mentions.ChannelMentioned,
		"all_mentioned", mentions.AllMentioned,
		"group_mentions", formatMentionsForLog(mentions.GroupMentions),
		"other_potential_mentions", mentions.OtherPotentialMentions)

	// Send notifications to mentioned users
	sender, err := p.API.GetUser(post.UserId)
	if err != nil {
		p.API.LogError("Failed to get sender for notification", "error", err.Error())
		return
	}

	// Extract post message to use in notification
	message := post.Message
	if len(message) > 100 {
		message = message[:97] + "..."
	}

	// Send notifications to all mentioned users
	for userID := range mentions.Mentions {
		// Don't send notifications to the post author
		if userID == post.UserId {
			continue
		}

		appErr := p.notificationService.SendMentionNotification(
			userID,
			post.Id,
			channel.DisplayName,
			sender.Username,
			message,
		)
		if appErr != nil {
			p.API.LogError("Failed to send mention notification",
				"error", appErr.Error(),
				"userId", userID)
		}
	}
}

// formatMentionsForLog converts a map of mentions to a comma-separated string for logging
func formatMentionsForLog(mentions map[string]MentionType) string {
	if len(mentions) == 0 {
		return "none"
	}

	result := "["
	first := true
	for id, mentionType := range mentions {
		if !first {
			result += ", "
		}
		result += id + ":" + formatMentionType(mentionType)
		first = false
	}
	return result + "]"
}

// formatMentionType converts a MentionType to a string representation
func formatMentionType(mentionType MentionType) string {
	switch mentionType {
	case NoMention:
		return "none"
	case GMMention:
		return "gm"
	case ThreadMention:
		return "thread"
	case CommentMention:
		return "comment"
	case ChannelMention:
		return "channel"
	case DMMention:
		return "dm"
	case KeywordMention:
		return "keyword"
	case GroupMention:
		return "group"
	default:
		return "unknown"
	}
}

// This will execute the commands that were registered in the NewCommandHandler function.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, err := p.commandClient.Handle(args)
	if err != nil {
		return nil, model.NewAppError("ExecuteCommand", "plugin.command.execute_command.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return response, nil
}
