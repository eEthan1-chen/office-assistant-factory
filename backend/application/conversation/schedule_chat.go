/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package conversation

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/api/model/conversation/run"
	appschedule "github.com/coze-dev/coze-studio/backend/application/schedule"
	crossmessage "github.com/coze-dev/coze-studio/backend/crossdomain/message/model"
	"github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/entity"
	convEntity "github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/entity"
	scheduledomain "github.com/coze-dev/coze-studio/backend/domain/schedule"
	sseImpl "github.com/coze-dev/coze-studio/backend/infra/sse/impl/sse"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/ptr"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

func (c *ConversationApplicationService) tryCreateScheduleFromAgentRun(ctx context.Context, sseSender *sseImpl.SSenderImpl, ar *run.AgentRunRequest, spaceID int64, conversationData *convEntity.Conversation) (bool, error) {
	result, ok, err := c.createScheduleFromAgentRun(ctx, ar, conversationData)
	if !ok || err != nil {
		return ok, err
	}
	c.sendAgentRunScheduleCreatedEvents(ctx, sseSender, ar, spaceID, conversationData, result)
	return true, nil
}

func (c *ConversationApplicationService) createScheduleFromAgentRun(ctx context.Context, ar *run.AgentRunRequest, conversationData *convEntity.Conversation) (*scheduleChatResult, bool, error) {
	text := strings.TrimSpace(ar.GetQuery())
	timezone := ""
	if ar.Extra != nil {
		timezone = strings.TrimSpace(ar.Extra["timezone"])
	}
	intentResult, ok, err := scheduledomain.ParseScheduleIntent(ctx, text, timezone, time.Now())
	if !ok || err != nil {
		return nil, ok, err
	}
	resp, err := appschedule.SVC.Create(ctx, &appschedule.CreateTaskRequest{
		AgentID:        ar.BotID,
		ConnectorID:    consts.CozeConnectorID,
		ConversationID: conversationData.ID,
		IsDraft:        true,
		Title:          intentResult.Intent.Title,
		Prompt:         intentResult.Intent.TaskPrompt,
		ScheduleText:   text,
		Timezone:       intentResult.Parsed.Timezone,
		TaskKind:       intentResult.Intent.TaskKind,
		ParsedSchedule: intentResult.Parsed,
	})
	if err != nil {
		return nil, true, err
	}
	taskID := int64(0)
	if resp != nil && resp.Data != nil {
		taskID = resp.Data.ID
	}
	return &scheduleChatResult{
		taskID:  taskID,
		content: buildScheduleConfirmation(intentResult.Intent.Title, intentResult.Parsed),
	}, true, nil
}

func (a *OpenapiAgentRunApplication) tryCreateScheduleFromChat(ctx context.Context, sseSender *sseImpl.SSenderImpl, ar *run.ChatV3Request, connectorID, spaceID int64, conversationData *convEntity.Conversation) (bool, error) {
	result, ok, err := a.createScheduleFromChat(ctx, ar, connectorID, conversationData)
	if !ok || err != nil {
		return ok, err
	}
	a.sendScheduleCreatedEvents(ctx, sseSender, ar, connectorID, spaceID, conversationData, result)
	return true, nil
}

func (a *OpenapiAgentRunApplication) tryCreateScheduleFromChatSync(ctx context.Context, ar *run.ChatV3Request, connectorID, spaceID int64, conversationData *convEntity.Conversation) (*run.RetrieveChatOpenResponse, bool, error) {
	result, ok, err := a.createScheduleFromChat(ctx, ar, connectorID, conversationData)
	if !ok || err != nil {
		return nil, ok, err
	}
	now := time.Now().UnixMilli()
	chatID := a.genScheduleChatID(ctx)
	return &run.RetrieveChatOpenResponse{
		ChatDetail: &run.ChatV3ChatDetail{
			ID:             chatID,
			ConversationID: conversationData.ID,
			BotID:          ar.BotID,
			Status:         string(entity.RunStatusCompleted),
			SectionID:      ptr.Of(conversationData.SectionID),
			CreatedAt:      ptr.Of(int32(now / 1000)),
			CompletedAt:    ptr.Of(int32(now / 1000)),
			MetaData: map[string]string{
				"schedule_task_id": strconv.FormatInt(result.taskID, 10),
				"task_type":        "schedule",
			},
		},
	}, true, nil
}

type scheduleChatResult struct {
	taskID  int64
	content string
}

func (a *OpenapiAgentRunApplication) createScheduleFromChat(ctx context.Context, ar *run.ChatV3Request, connectorID int64, conversationData *convEntity.Conversation) (*scheduleChatResult, bool, error) {
	text := lastUserTextFromChatV3(ar)
	timezone := scheduleTimezone(ar)
	intentResult, ok, err := scheduledomain.ParseScheduleIntent(ctx, text, timezone, time.Now())
	if !ok || err != nil {
		return nil, ok, err
	}
	resp, err := appschedule.SVC.Create(ctx, &appschedule.CreateTaskRequest{
		AgentID:        ar.BotID,
		ConnectorID:    connectorID,
		ConversationID: conversationData.ID,
		Title:          intentResult.Intent.Title,
		Prompt:         intentResult.Intent.TaskPrompt,
		ScheduleText:   text,
		Timezone:       intentResult.Parsed.Timezone,
		TaskKind:       intentResult.Intent.TaskKind,
		ParsedSchedule: intentResult.Parsed,
	})
	if err != nil {
		return nil, true, err
	}
	taskID := int64(0)
	if resp != nil && resp.Data != nil {
		taskID = resp.Data.ID
	}
	return &scheduleChatResult{
		taskID:  taskID,
		content: buildScheduleConfirmation(intentResult.Intent.Title, intentResult.Parsed),
	}, true, nil
}

func lastUserTextFromChatV3(ar *run.ChatV3Request) string {
	for i := len(ar.AdditionalMessages) - 1; i >= 0; i-- {
		msg := ar.AdditionalMessages[i]
		if msg == nil || msg.Role != string(schema.User) {
			continue
		}
		switch msg.ContentType {
		case run.ContentTypeText:
			return strings.TrimSpace(msg.Content)
		case run.ContentTypeMixApi:
			var inputs []*run.AdditionalContent
			if err := json.Unmarshal([]byte(msg.Content), &inputs); err != nil {
				return ""
			}
			var parts []string
			for _, input := range inputs {
				if input != nil && input.Type == string(crossmessage.InputTypeText) {
					parts = append(parts, strings.TrimSpace(ptr.From(input.Text)))
				}
			}
			return strings.TrimSpace(strings.Join(parts, " "))
		}
	}
	return ""
}

func scheduleTimezone(ar *run.ChatV3Request) string {
	if ar.ExtraParams != nil {
		return strings.TrimSpace(ar.ExtraParams["timezone"])
	}
	return ""
}

func reminderTitle(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	for _, prefix := range []string{"提醒我", "通知我", "叫我", "发给我", "发送给我", "推送给我", "告诉我", "发我", "给我发"} {
		if strings.HasPrefix(prompt, prefix) {
			title := strings.TrimSpace(strings.TrimPrefix(prompt, prefix))
			if title != "" {
				return title
			}
		}
	}
	return prompt
}

func buildScheduleConfirmation(prompt string, parsed *scheduledomain.ParsedSchedule) string {
	runAt := ""
	if len(parsed.PreviewNextRuns) > 0 {
		loc, err := time.LoadLocation(parsed.Timezone)
		if err == nil {
			runAt = time.UnixMilli(parsed.PreviewNextRuns[0]).In(loc).Format("2006-01-02 15:04")
		}
	}
	if runAt == "" {
		return "已为你创建提醒：" + prompt
	}
	return "已为你创建提醒：" + prompt + "，将在 " + runAt + " 触发。"
}

func (a *OpenapiAgentRunApplication) sendScheduleCreatedEvents(ctx context.Context, sseSender *sseImpl.SSenderImpl, ar *run.ChatV3Request, connectorID, spaceID int64, conversationData *convEntity.Conversation, result *scheduleChatResult) {
	now := time.Now().UnixMilli()
	chatID := a.genScheduleChatID(ctx)
	messageID := a.genScheduleChatID(ctx)
	toolMessageID := a.genScheduleChatID(ctx)
	runItem := &entity.ChunkRunItem{
		ID:             chatID,
		ConversationID: conversationData.ID,
		SectionID:      conversationData.SectionID,
		AgentID:        ar.BotID,
		Status:         entity.RunStatusCompleted,
		CreatedAt:      now,
		UpdatedAt:      now,
		CompletedAt:    now,
		CreatorID:      conversationData.CreatorID,
	}
	msgItem := &entity.ChunkMessageItem{
		ID:             messageID,
		ConversationID: conversationData.ID,
		SectionID:      conversationData.SectionID,
		RunID:          chatID,
		AgentID:        ar.BotID,
		Role:           entity.RoleTypeAssistant,
		Content:        result.content,
		ContentType:    crossmessage.ContentTypeText,
		MessageType:    crossmessage.MessageTypeAnswer,
		Type:           crossmessage.MessageTypeAnswer,
		Ext: map[string]string{
			"schedule_task_id": strconv.FormatInt(result.taskID, 10),
			"task_type":        "schedule",
			"connector_id":     strconv.FormatInt(connectorID, 10),
			"space_id":         strconv.FormatInt(spaceID, 10),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	msgDoneItem := *msgItem
	msgDoneItem.IsFinish = true
	toolMsgItem := buildScheduleToolResponseMessage(toolMessageID, chatID, ar.BotID, conversationData, result, now)
	sseSender.Send(ctx, buildMessageChunkEvent(string(entity.RunEventCreated), buildARSM2ApiChatMessage(&entity.AgentRunResponse{Event: entity.RunEventCreated, ChunkRunItem: runItem})))
	sseSender.Send(ctx, buildMessageChunkEvent(string(entity.RunEventMessageCompleted), buildARSM2ApiMessage(&entity.AgentRunResponse{Event: entity.RunEventMessageCompleted, ChunkMessageItem: toolMsgItem})))
	sseSender.Send(ctx, buildMessageChunkEvent(string(entity.RunEventMessageDelta), buildARSM2ApiMessage(&entity.AgentRunResponse{Event: entity.RunEventMessageDelta, ChunkMessageItem: msgItem})))
	sseSender.Send(ctx, buildMessageChunkEvent(string(entity.RunEventMessageCompleted), buildARSM2ApiMessage(&entity.AgentRunResponse{Event: entity.RunEventMessageCompleted, ChunkMessageItem: &msgDoneItem})))
	sseSender.Send(ctx, buildMessageChunkEvent(string(entity.RunEventCompleted), buildARSM2ApiChatMessage(&entity.AgentRunResponse{Event: entity.RunEventCompleted, ChunkRunItem: runItem})))
	sseSender.Send(ctx, buildDoneEvent(string(entity.RunEventStreamDone)))
}

func (c *ConversationApplicationService) sendAgentRunScheduleCreatedEvents(ctx context.Context, sseSender *sseImpl.SSenderImpl, ar *run.AgentRunRequest, spaceID int64, conversationData *convEntity.Conversation, result *scheduleChatResult) {
	now := time.Now().UnixMilli()
	chatID := c.genScheduleChatID(ctx)
	ackMessageID := c.genScheduleChatID(ctx)
	toolMessageID := c.genScheduleChatID(ctx)
	answerMessageID := c.genScheduleChatID(ctx)
	ackMsgItem := &entity.ChunkMessageItem{
		ID:             ackMessageID,
		ConversationID: conversationData.ID,
		SectionID:      conversationData.SectionID,
		RunID:          chatID,
		AgentID:        ar.BotID,
		Role:           entity.RoleTypeUser,
		Content:        ar.GetQuery(),
		ContentType:    crossmessage.ContentTypeText,
		MessageType:    crossmessage.MessageTypeAck,
		Type:           crossmessage.MessageTypeAck,
		ReplyID:        ackMessageID,
		CreatedAt:      now,
		UpdatedAt:      now,
		IsFinish:       true,
	}
	toolMsgItem := buildScheduleToolResponseMessage(toolMessageID, chatID, ar.BotID, conversationData, result, now)
	toolMsgItem.ReplyID = ackMessageID
	answerMsgItem := &entity.ChunkMessageItem{
		ID:             answerMessageID,
		ConversationID: conversationData.ID,
		SectionID:      conversationData.SectionID,
		RunID:          chatID,
		AgentID:        ar.BotID,
		Role:           entity.RoleTypeAssistant,
		Content:        result.content,
		ContentType:    crossmessage.ContentTypeText,
		MessageType:    crossmessage.MessageTypeAnswer,
		Type:           crossmessage.MessageTypeAnswer,
		ReplyID:        ackMessageID,
		Ext: map[string]string{
			"schedule_task_id": strconv.FormatInt(result.taskID, 10),
			"task_type":        "schedule",
			"connector_id":     strconv.FormatInt(consts.CozeConnectorID, 10),
			"space_id":         strconv.FormatInt(spaceID, 10),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	answerDoneItem := *answerMsgItem
	answerDoneItem.IsFinish = true
	sseSender.Send(ctx, buildMessageChunkEvent(run.RunEventMessage, buildARSM2Message(&entity.AgentRunResponse{Event: entity.RunEventAck, ChunkMessageItem: ackMsgItem}, ar)))
	sseSender.Send(ctx, buildMessageChunkEvent(run.RunEventMessage, buildARSM2Message(&entity.AgentRunResponse{Event: entity.RunEventMessageCompleted, ChunkMessageItem: toolMsgItem}, ar)))
	sseSender.Send(ctx, buildMessageChunkEvent(run.RunEventMessage, buildARSM2Message(&entity.AgentRunResponse{Event: entity.RunEventMessageDelta, ChunkMessageItem: answerMsgItem}, ar)))
	sseSender.Send(ctx, buildMessageChunkEvent(run.RunEventMessage, buildARSM2Message(&entity.AgentRunResponse{Event: entity.RunEventMessageCompleted, ChunkMessageItem: &answerDoneItem}, ar)))
	sseSender.Send(ctx, buildDoneEvent(run.RunEventDone))
}

func buildScheduleToolResponseMessage(messageID, chatID, agentID int64, conversationData *convEntity.Conversation, result *scheduleChatResult, now int64) *entity.ChunkMessageItem {
	return &entity.ChunkMessageItem{
		ID:             messageID,
		ConversationID: conversationData.ID,
		SectionID:      conversationData.SectionID,
		RunID:          chatID,
		AgentID:        agentID,
		Role:           entity.RoleTypeTool,
		Content:        buildScheduleToolResponseContent(result),
		ContentType:    crossmessage.ContentTypeText,
		MessageType:    crossmessage.MessageTypeToolResponse,
		Type:           crossmessage.MessageTypeToolResponse,
		Ext: map[string]string{
			"schedule_task_id": strconv.FormatInt(result.taskID, 10),
			"task_type":        "schedule",
		},
		CreatedAt: now,
		UpdatedAt: now,
		IsFinish:  true,
	}
}

func buildScheduleToolResponseContent(result *scheduleChatResult) string {
	data := map[string]string{
		"response_for_model": "Task created successfully",
		"schedule_task_id":   strconv.FormatInt(result.taskID, 10),
		"message":            result.content,
	}
	b, _ := json.Marshal(data)
	return string(b)
}

func (a *OpenapiAgentRunApplication) genScheduleChatID(ctx context.Context) int64 {
	if ConversationSVC != nil && ConversationSVC.appContext != nil && ConversationSVC.appContext.IDGen != nil {
		if id, err := ConversationSVC.appContext.IDGen.GenID(ctx); err == nil {
			return id
		}
	}
	return time.Now().UnixNano()
}

func (c *ConversationApplicationService) genScheduleChatID(ctx context.Context) int64 {
	if c != nil && c.appContext != nil && c.appContext.IDGen != nil {
		if id, err := c.appContext.IDGen.GenID(ctx); err == nil {
			return id
		}
	}
	return time.Now().UnixNano()
}
