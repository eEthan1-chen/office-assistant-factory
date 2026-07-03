/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/api/model/conversation/common"
	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	messageModel "github.com/coze-dev/coze-studio/backend/crossdomain/message/model"
	agentEntity "github.com/coze-dev/coze-studio/backend/domain/agent/singleagent/entity"
	agentrun "github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/entity"
	agentrunService "github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/service"
	convEntity "github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/entity"
	conversationService "github.com/coze-dev/coze-studio/backend/domain/conversation/conversation/service"
	domain "github.com/coze-dev/coze-studio/backend/domain/schedule"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/conv"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
	"github.com/coze-dev/coze-studio/backend/types/consts"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

type ServiceComponents struct {
	DB                    *gorm.DB
	IDGen                 idgen.IDGenerator
	AgentRunDomainSVC     agentrunService.Run
	ConversationDomainSVC conversationService.Conversation
	SingleAgentDomainSVC  interface {
		ObtainAgentByIdentity(ctx context.Context, identity *agentEntity.AgentIdentity) (*agentEntity.SingleAgent, error)
	}
}

type ApplicationService struct {
	db                    *gorm.DB
	repo                  domain.Repository
	agentRunDomainSVC     agentrunService.Run
	conversationDomainSVC conversationService.Conversation
	singleAgentDomainSVC  interface {
		ObtainAgentByIdentity(ctx context.Context, identity *agentEntity.AgentIdentity) (*agentEntity.SingleAgent, error)
	}
}

var SVC = &ApplicationService{}

func InitService(ctx context.Context, c *ServiceComponents) (*ApplicationService, error) {
	repo := domain.NewRepository(c.DB, c.IDGen)
	if err := repo.Migrate(ctx); err != nil {
		return nil, err
	}
	SVC = &ApplicationService{
		db:                    c.DB,
		repo:                  repo,
		agentRunDomainSVC:     c.AgentRunDomainSVC,
		conversationDomainSVC: c.ConversationDomainSVC,
		singleAgentDomainSVC:  c.SingleAgentDomainSVC,
	}
	return SVC, nil
}

func (s *ApplicationService) Parse(ctx context.Context, req *ParseTaskRequest) (*ParseTaskResponse, error) {
	parsed, err := domain.ParseNaturalLanguage(req.ScheduleText, req.Timezone, time.Now())
	if err != nil {
		return nil, err
	}
	return &ParseTaskResponse{Code: 0, Msg: "success", Data: parsed}, nil
}

func (s *ApplicationService) Create(ctx context.Context, req *CreateTaskRequest) (*TaskResponse, error) {
	if s == nil || s.repo == nil || s.singleAgentDomainSVC == nil || s.conversationDomainSVC == nil {
		return nil, fmt.Errorf("schedule service is not initialized")
	}
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed)
	}
	if req.AgentID <= 0 || strings.TrimSpace(req.Prompt) == "" || strings.TrimSpace(req.ScheduleText) == "" {
		return nil, fmt.Errorf("agent_id, prompt and schedule_text are required")
	}
	connectorID := req.ConnectorID
	if connectorID == 0 {
		connectorID = consts.APIConnectorID
	}

	agent, err := s.singleAgentDomainSVC.ObtainAgentByIdentity(ctx, &agentEntity.AgentIdentity{
		AgentID:     req.AgentID,
		IsDraft:     req.IsDraft,
		ConnectorID: connectorID,
	})
	if err != nil {
		return nil, err
	}
	if agent == nil || agent.CreatorID != userID {
		return nil, errorx.New(errno.ErrConversationPermissionCode, errorx.KV("msg", "agent not match"))
	}

	conversationID := req.ConversationID
	if conversationID == 0 {
		convData, err := s.conversationDomainSVC.Create(ctx, &convEntity.CreateMeta{
			AgentID:     req.AgentID,
			CreatorID:   userID,
			Scene:       common.Scene_Default,
			ConnectorID: connectorID,
		})
		if err != nil {
			return nil, err
		}
		conversationID = convData.ID
	} else if err := s.checkConversationOwner(ctx, conversationID, userID); err != nil {
		return nil, err
	}

	parsed := req.ParsedSchedule
	if parsed == nil {
		var err error
		parsed, err = domain.ParseNaturalLanguage(req.ScheduleText, req.Timezone, time.Now())
		if err != nil {
			return nil, err
		}
	}
	id, err := s.repo.GenID(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSpace(req.ScheduleText)
	}
	task := &domain.AgentScheduleTask{
		ID:             id,
		UserID:         userID,
		SpaceID:        agent.SpaceID,
		AgentID:        req.AgentID,
		ConnectorID:    connectorID,
		ConversationID: conversationID,
		IsDraft:        req.IsDraft,
		Title:          title,
		Prompt:         strings.TrimSpace(req.Prompt),
		ScheduleText:   strings.TrimSpace(req.ScheduleText),
		TaskKind:       domain.NormalizeTaskKind(req.TaskKind),
		CronExpr:       parsed.CronExpr,
		TriggerType:    parsed.TriggerType,
		Timezone:       parsed.Timezone,
		Status:         domain.TaskStatusEnabled,
		NextRunAt:      parsed.PreviewNextRuns[0],
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	return &TaskResponse{Code: 0, Msg: "success", Data: task}, nil
}

func (s *ApplicationService) List(ctx context.Context, agentID int64) (*ListTasksResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed)
	}
	tasks, err := s.repo.ListTasks(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	return &ListTasksResponse{Code: 0, Msg: "success", Data: tasks}, nil
}

func (s *ApplicationService) Update(ctx context.Context, req *UpdateTaskRequest) (*TaskResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed)
	}
	task, err := s.repo.GetTask(ctx, req.TaskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, errorx.New(errno.ErrConversationPermissionCode)
	}
	if req.Title != nil {
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.Prompt != nil {
		task.Prompt = strings.TrimSpace(*req.Prompt)
	}
	if req.Enabled != nil {
		if *req.Enabled {
			task.Status = domain.TaskStatusEnabled
		} else {
			task.Status = domain.TaskStatusDisabled
		}
	}
	if req.ScheduleText != nil || req.Timezone != nil {
		scheduleText := task.ScheduleText
		timezone := task.Timezone
		if req.ScheduleText != nil {
			scheduleText = strings.TrimSpace(*req.ScheduleText)
		}
		if req.Timezone != nil {
			timezone = strings.TrimSpace(*req.Timezone)
		}
		parsed, err := domain.ParseNaturalLanguage(scheduleText, timezone, time.Now())
		if err != nil {
			return nil, err
		}
		task.ScheduleText = scheduleText
		task.CronExpr = parsed.CronExpr
		task.TriggerType = parsed.TriggerType
		task.Timezone = parsed.Timezone
		task.NextRunAt = parsed.PreviewNextRuns[0]
	}
	task.UpdatedAt = time.Now().UnixMilli()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		return nil, err
	}
	return &TaskResponse{Code: 0, Msg: "success", Data: task}, nil
}

func (s *ApplicationService) Delete(ctx context.Context, req *DeleteTaskRequest) (*TaskResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed)
	}
	if err := s.repo.DeleteTask(ctx, req.TaskID, userID, time.Now().UnixMilli()); err != nil {
		return nil, err
	}
	return &TaskResponse{Code: 0, Msg: "success"}, nil
}

func (s *ApplicationService) Runs(ctx context.Context, taskID int64, limit int) (*ListRunsResponse, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, errorx.New(errno.ErrUserAuthenticationFailed)
	}
	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, errorx.New(errno.ErrConversationPermissionCode)
	}
	runs, err := s.repo.ListRuns(ctx, taskID, limit)
	if err != nil {
		return nil, err
	}
	runData := make([]*ScheduleRunWithData, 0, len(runs))
	for _, run := range runs {
		item := &ScheduleRunWithData{AgentScheduleRun: run}
		if run.ChatID > 0 {
			input, output, followUps, err := s.getRunMessages(ctx, run.ChatID)
			if err != nil {
				logs.CtxWarnf(ctx, "get schedule run messages failed, chatID=%d, err=%v", run.ChatID, err)
			} else {
				item.Input = input
				item.Output = output
				item.FollowUps = followUps
			}
		}
		runData = append(runData, item)
	}
	return &ListRunsResponse{Code: 0, Msg: "success", Data: runData}, nil
}

type scheduleRunMessage struct {
	Role        string `gorm:"column:role"`
	MessageType string `gorm:"column:message_type"`
	Content     string `gorm:"column:content"`
	CreatedAt   int64  `gorm:"column:created_at"`
}

func (s *ApplicationService) getRunMessages(ctx context.Context, chatID int64) (input string, output string, followUps []string, err error) {
	var messages []*scheduleRunMessage
	err = s.db.WithContext(ctx).
		Table("message").
		Select("role, message_type, content, created_at").
		Where("run_id = ? AND status = ?", chatID, 1).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return "", "", nil, err
	}
	for _, msg := range messages {
		switch {
		case msg.Role == "user" && msg.MessageType == "question" && input == "":
			input = msg.Content
		case msg.Role == "assistant" && msg.MessageType == "answer" && output == "":
			output = msg.Content
		case msg.Role == "assistant" && msg.MessageType == "follow_up":
			followUps = append(followUps, msg.Content)
		}
	}
	return input, output, followUps, nil
}

func (s *ApplicationService) checkConversationOwner(ctx context.Context, conversationID, userID int64) error {
	convData, err := s.conversationDomainSVC.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}
	if convData == nil || convData.CreatorID != userID {
		return errorx.New(errno.ErrConversationPermissionCode, errorx.KV("msg", "conversation not match"))
	}
	return nil
}

func userIDFromContext(ctx context.Context) (int64, bool) {
	if uid := ctxutil.GetUIDFromCtx(ctx); uid != nil {
		return *uid, true
	}
	if apiAuth := ctxutil.GetApiAuthFromCtx(ctx); apiAuth != nil && apiAuth.UserID > 0 {
		return apiAuth.UserID, true
	}
	return 0, false
}

func (s *ApplicationService) executeTask(ctx context.Context, task *domain.AgentScheduleTask) (int64, error) {
	convData, err := s.conversationDomainSVC.GetByID(ctx, task.ConversationID)
	if err != nil {
		return 0, err
	}
	if convData == nil {
		return 0, errors.New("conversation not found")
	}
	streamer, err := s.agentRunDomainSVC.AgentRun(ctx, &agentrun.AgentRunMeta{
		ConversationID: task.ConversationID,
		AgentID:        task.AgentID,
		Content: []*messageModel.InputMetaData{{
			Type: messageModel.InputTypeText,
			Text: task.Prompt,
		}},
		DisplayContent: task.Prompt,
		SpaceID:        task.SpaceID,
		UserID:         conv.Int64ToStr(task.UserID),
		CozeUID:        task.UserID,
		SectionID:      convData.SectionID,
		IsDraft:        task.IsDraft,
		ConnectorID:    task.ConnectorID,
		ContentType:    messageModel.ContentTypeText,
		Ext: map[string]string{
			"schedule_task_id": conv.Int64ToStr(task.ID),
			"task_type":        "schedule",
		},
	})
	if err != nil {
		return 0, err
	}
	var chatID int64
	for {
		chunk, recvErr := streamer.Recv()
		if recvErr != nil {
			if errors.Is(recvErr, io.EOF) {
				break
			}
			return chatID, recvErr
		}
		if chunk == nil {
			continue
		}
		if chunk.ChunkRunItem != nil && chunk.ChunkRunItem.ID > 0 {
			chatID = chunk.ChunkRunItem.ID
		}
		if chunk.Event == agentrun.RunEventError && chunk.Error != nil {
			return chatID, errors.New(chunk.Error.Msg)
		}
		if chunk.Event == agentrun.RunEventStreamDone {
			break
		}
	}
	return chatID, nil
}

func (s *ApplicationService) StartScheduler(ctx context.Context) {
	go s.schedulerLoop(ctx)
}

func (s *ApplicationService) schedulerLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scanDueTasks(context.Background())
		}
	}
}

func (s *ApplicationService) scanDueTasks(ctx context.Context) {
	now := time.Now()
	tasks, err := s.repo.DueTasks(ctx, now.UnixMilli(), 20)
	if err != nil {
		logs.CtxErrorf(ctx, "[schedule] scan due tasks failed: %v", err)
		return
	}
	for _, task := range tasks {
		task := task
		go s.runDueTask(context.Background(), task)
	}
}

func (s *ApplicationService) runDueTask(ctx context.Context, task *domain.AgentScheduleTask) {
	now := time.Now()
	nextRunAt := int64(0)
	if task.TriggerType != domain.TriggerTypeOneTime {
		nextRuns, err := domain.NextRuns(task.CronExpr, task.Timezone, now, 1)
		if err != nil {
			logs.CtxErrorf(ctx, "[schedule] compute next run failed task=%d: %v", task.ID, err)
			return
		}
		nextRunAt = nextRuns[0]
	}
	claimed, err := s.repo.ClaimTask(ctx, task.ID, task.NextRunAt, nextRunAt, now.UnixMilli())
	if err != nil || !claimed {
		if err != nil {
			logs.CtxErrorf(ctx, "[schedule] claim task failed task=%d: %v", task.ID, err)
		}
		return
	}

	runID, err := s.repo.GenID(ctx)
	if err != nil {
		logs.CtxErrorf(ctx, "[schedule] gen run id failed task=%d: %v", task.ID, err)
		return
	}
	run := &domain.AgentScheduleRun{
		ID:             runID,
		TaskID:         task.ID,
		RunStatus:      domain.RunStatusRunning,
		ConversationID: task.ConversationID,
		StartedAt:      now.UnixMilli(),
	}
	if err := s.repo.CreateRun(ctx, run); err != nil {
		logs.CtxErrorf(ctx, "[schedule] create run failed task=%d: %v", task.ID, err)
		return
	}

	var (
		chatID  int64
		execErr error
		output  string
	)
	if task.TaskKind == domain.TaskKindNotificationOnly {
		output = task.Prompt
	} else {
		chatID, execErr = s.executeTask(ctx, task)
	}
	var weComErr error
	if execErr == nil {
		if task.TaskKind != domain.TaskKindNotificationOnly {
			var msgErr error
			_, output, _, msgErr = s.getRunMessages(ctx, chatID)
			if msgErr != nil {
				weComErr = fmt.Errorf("get schedule run output failed: %w", msgErr)
			}
		}
		if weComErr == nil {
			weComErr = notifyWeComForScheduleTask(ctx, task, output)
		}
		if weComErr != nil {
			logs.CtxErrorf(ctx, "[schedule] notify wecom failed task=%d: %v", task.ID, weComErr)
		}
	}
	status := domain.RunStatusSuccess
	errMsg := ""
	if weComErr != nil || execErr != nil {
		status = domain.RunStatusFailed
		switch {
		case weComErr != nil && execErr != nil:
			errMsg = fmt.Sprintf("wecom notify failed: %v; agent run failed: %v", weComErr, execErr)
		case weComErr != nil:
			errMsg = fmt.Sprintf("wecom notify failed: %v", weComErr)
		default:
			errMsg = execErr.Error()
		}
	}
	if err := s.repo.FinishRun(ctx, runID, status, chatID, errMsg, time.Now().UnixMilli()); err != nil {
		logs.CtxErrorf(ctx, "[schedule] finish run failed run=%d: %v", runID, err)
	}
}

// TriggerDueTasksOnce is intentionally exported for smoke tests and local debugging.
func (s *ApplicationService) TriggerDueTasksOnce(ctx context.Context) {
	s.scanDueTasks(ctx)
}
