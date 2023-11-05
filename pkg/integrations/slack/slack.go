package slack

import (
	"errors"
	"fmt"

	"github.com/hatchet-dev/hatchet-workflows/pkg/workflows/types"
	"github.com/slack-go/slack"
)

type SlackIntegration struct {
	api    *slack.Client
	teamId string
}

func NewSlackIntegration(authToken string, teamId string, debug bool) *SlackIntegration {
	api := slack.New(authToken, slack.OptionDebug(debug))

	return &SlackIntegration{
		api:    api,
		teamId: teamId,
	}
}

func (s *SlackIntegration) GetId() string {
	return "slack"
}

func (s *SlackIntegration) Actions() []string {
	return []string{
		"create-channel",
		"send-message",
		"add-users-to-channel",
	}
}

func (s *SlackIntegration) PerformAction(action types.Action, data map[string]interface{}) (map[string]interface{}, error) {
	fmt.Println("GOT SLACK", action.String())
	switch action.Verb {
	case "create-channel":
		return s.createChannel(data)
	case "add-users-to-channel":
		return s.addUsersToChannel(data)
	case "send-message":
		return s.sendMessageToChannel(data)
	default:
		return nil, fmt.Errorf("unsupported action: %s", action)
	}
}

func (s *SlackIntegration) createChannel(data map[string]interface{}) (map[string]interface{}, error) {
	dataName, ok := data["channelName"]

	if !ok || dataName == nil {
		return nil, errors.New("missing required field: name")
	}

	name, ok := dataName.(string)

	if !ok {
		return nil, errors.New("invalid type for field: name")
	}

	channel, err := s.api.CreateConversation(slack.CreateConversationParams{
		IsPrivate:   true,
		ChannelName: name,
		TeamID:      s.teamId,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating slack channel: %w", err)
	}

	return map[string]interface{}{
		"channelId": channel.ID,
	}, nil
}

func (s *SlackIntegration) addUsersToChannel(data map[string]interface{}) (map[string]interface{}, error) {
	channelId, ok := data["channelId"]

	if !ok || channelId == nil {
		return nil, errors.New("missing required field: channelId")
	}

	channelIdStr, ok := channelId.(string)

	if !ok {
		return nil, errors.New("invalid type for field: channelId")
	}

	userIds, ok := data["userIds"]

	if !ok || userIds == nil {
		return nil, errors.New("missing required field: userIds")
	}

	userIdsArr, ok := userIds.([]any)

	if !ok {
		return nil, errors.New("invalid type for field: userIds")
	}

	userIdsStrArr := make([]string, len(userIdsArr))

	for i, userId := range userIdsArr {
		userIdStr, ok := userId.(string)

		if !ok {
			return nil, errors.New("invalid type for field: userIds")
		}

		userIdsStrArr[i] = userIdStr
	}

	_, err := s.api.InviteUsersToConversation(channelIdStr, userIdsStrArr...)

	if err != nil {
		return nil, fmt.Errorf("error adding users to slack channel: %w", err)
	}

	return map[string]interface{}{}, nil
}

func (s *SlackIntegration) sendMessageToChannel(data map[string]interface{}) (map[string]interface{}, error) {
	channelId, ok := data["channelId"]

	if !ok || channelId == nil {
		return nil, errors.New("missing required field: channelId")
	}

	channelIdStr, ok := channelId.(string)

	if !ok {
		return nil, errors.New("invalid type for field: channelId")
	}

	message, ok := data["message"]

	if !ok || message == nil {
		return nil, errors.New("missing required field: message")
	}

	messageStr, ok := message.(string)

	if !ok {
		return nil, errors.New("invalid type for field: message")
	}

	_, _, err := s.api.PostMessage(channelIdStr, slack.MsgOptionText(messageStr, false))

	if err != nil {
		return nil, fmt.Errorf("error sending message to slack channel: %w", err)
	}

	return map[string]interface{}{}, nil
}
