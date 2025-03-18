package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-shoutrrr/server/command"
	"github.com/mattermost/mattermost-plugin-shoutrrr/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
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

// This will execute the commands that were registered in the NewCommandHandler function.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, err := p.commandClient.Handle(args)
	if err != nil {
		return nil, model.NewAppError("ExecuteCommand", "plugin.command.execute_command.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return response, nil
}

func (p *Plugin) IsCRTEnabledForUser(userID string) bool {
	appCRT := *p.API.GetConfig().ServiceSettings.CollapsedThreads
	if appCRT == model.CollapsedThreadsDisabled {
		return false
	}
	if appCRT == model.CollapsedThreadsAlwaysOn {
		return true
	}
	threadsEnabled := appCRT == model.CollapsedThreadsDefaultOn
	// check if a participant has overridden collapsed threads settings
	if preference, err := p.API.GetPreferenceForUser(userID, model.PreferenceCategoryDisplaySettings, model.PreferenceNameCollapsedThreadsEnabled); err == nil {
		threadsEnabled = preference.Value == "on"
	}
	return threadsEnabled
}

// allowChannelMentions returns whether or not the channel mentions are allowed for the given post.
func (p *Plugin) allowChannelMentions(post *model.Post, numProfiles int) bool {
	if !p.API.HasPermissionToChannel(post.UserId, post.ChannelId, model.PermissionUseChannelMentions) {
		return false
	}

	if post.Type == model.PostTypeHeaderChange || post.Type == model.PostTypePurposeChange {
		return false
	}

	if int64(numProfiles) >= *p.API.GetConfig().TeamSettings.MaxNotificationsPerChannel {
		return false
	}

	return true
}

func (p *Plugin) getMentionKeywordsInChannel(profiles map[string]*model.User, allowChannelMentions bool, channelMemberNotifyPropsMap map[string]model.StringMap, groups map[string]*model.Group) MentionKeywords {
	keywords := make(MentionKeywords)

	for _, profile := range profiles {
		status, _ := p.API.GetUserStatus(profile.Id)
		keywords.AddUser(
			profile,
			channelMemberNotifyPropsMap[profile.Id],
			status,
			allowChannelMentions,
		)
	}

	keywords.AddGroupsMap(groups)

	return keywords
}

func (p *Plugin) getExplicitMentionsAndKeywords(post *model.Post, channel *model.Channel, profileMap map[string]*model.User, groups map[string]*model.Group, channelMemberNotifyPropsMap map[string]model.StringMap, parentPostList *model.PostList) (*MentionResults, MentionKeywords) {
	mentions := &MentionResults{}
	var isAllowChannelMentions bool
	var keywords MentionKeywords

	if channel.Type == model.ChannelTypeDirect {
		isWebhook := post.GetProp("from_webhook") == "true"

		// A bot can post in a DM where it doesn't belong to.
		// Therefore, we cannot "guess" who is the other user,
		// so we add the mention to any user that is not the
		// poster unless the post comes from a webhook.
		user1, user2 := channel.GetBothUsersForDM()
		if (post.UserId != user1) || isWebhook {
			if _, ok := profileMap[user1]; ok {
				mentions.addMention(user1, DMMention)
			} else {
				p.API.LogDebug("missing profile: DM user not in profiles", mlog.String("userId", user1), mlog.String("channelId", channel.Id))
			}
		}

		if user2 != "" {
			if (post.UserId != user2) || isWebhook {
				if _, ok := profileMap[user2]; ok {
					mentions.addMention(user2, DMMention)
				} else {
					p.API.LogDebug("missing profile: DM user not in profiles", mlog.String("userId", user2), mlog.String("channelId", channel.Id))
				}
			}
		}
	} else {
		isAllowChannelMentions = p.allowChannelMentions(post, len(profileMap))
		keywords = p.getMentionKeywordsInChannel(profileMap, isAllowChannelMentions, channelMemberNotifyPropsMap, groups)

		mentions = getExplicitMentions(post, keywords)

		// Add a GM mention to all members of a GM channel
		if channel.Type == model.ChannelTypeGroup {
			for id := range channelMemberNotifyPropsMap {
				if _, ok := profileMap[id]; ok {
					mentions.addMention(id, GMMention)
				} else {
					p.API.LogDebug("missing profile: GM user not in profiles", mlog.String("userId", id), mlog.String("channelId", channel.Id))
				}
			}
		}

		// Add an implicit mention when a user is added to a channel
		// even if the user has set 'username mentions' to false in account settings.
		if post.Type == model.PostTypeAddToChannel {
			if addedUserId, ok := post.GetProp(model.PostPropsAddedUserId).(string); ok {
				if _, ok := profileMap[addedUserId]; ok {
					mentions.addMention(addedUserId, KeywordMention)
				} else {
					p.API.LogDebug("missing profile: user added to channel not in profiles", mlog.String("userId", addedUserId), mlog.String("channelId", channel.Id))
				}
			}
		}

		// Get users that have comment thread mentions enabled
		if post.RootId != "" && parentPostList != nil {
			for _, threadPost := range parentPostList.Posts {
				profile := profileMap[threadPost.UserId]
				if profile == nil {
					// Not logging missing profile since this is relatively expected
					continue
				}

				// If this is the root post and it was posted by an OAuth bot, don't notify the user
				if threadPost.Id == parentPostList.Order[0] && threadPost.IsFromOAuthBot() {
					continue
				}
				if p.IsCRTEnabledForUser(profile.Id) {
					continue
				}
				if profile.NotifyProps[model.CommentsNotifyProp] == model.CommentsNotifyAny || (profile.NotifyProps[model.CommentsNotifyProp] == model.CommentsNotifyRoot && threadPost.Id == parentPostList.Order[0]) {
					mentionType := ThreadMention
					if threadPost.Id == parentPostList.Order[0] {
						mentionType = CommentMention
					}

					mentions.addMention(threadPost.UserId, mentionType)
				}
			}
		}

		// Prevent the user from mentioning themselves
		if post.GetProp("from_webhook") != "true" {
			mentions.removeMention(post.UserId)
		}
	}

	return mentions, keywords
}
