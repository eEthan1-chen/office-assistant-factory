/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import domain "github.com/coze-dev/coze-studio/backend/domain/schedule"

type ParseTaskRequest struct {
	ScheduleText string `json:"schedule_text"`
	Timezone     string `json:"timezone"`
}

type ParseTaskResponse struct {
	Code int32                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data *domain.ParsedSchedule `json:"data"`
}

type CreateTaskRequest struct {
	AgentID        int64                  `json:"agent_id,string"`
	ConnectorID    int64                  `json:"connector_id,string"`
	ConversationID int64                  `json:"conversation_id,string"`
	IsDraft        bool                   `json:"is_draft"`
	Title          string                 `json:"title"`
	Prompt         string                 `json:"prompt"`
	ScheduleText   string                 `json:"schedule_text"`
	Timezone       string                 `json:"timezone"`
	TaskKind       string                 `json:"-"`
	ParsedSchedule *domain.ParsedSchedule `json:"-"`
}

type TaskResponse struct {
	Code int32                     `json:"code"`
	Msg  string                    `json:"msg"`
	Data *domain.AgentScheduleTask `json:"data"`
}

type ListTasksResponse struct {
	Code int32                       `json:"code"`
	Msg  string                      `json:"msg"`
	Data []*domain.AgentScheduleTask `json:"data"`
}

type UpdateTaskRequest struct {
	TaskID       int64   `json:"task_id,string"`
	Title        *string `json:"title"`
	Prompt       *string `json:"prompt"`
	ScheduleText *string `json:"schedule_text"`
	Timezone     *string `json:"timezone"`
	Enabled      *bool   `json:"enabled"`
}

type DeleteTaskRequest struct {
	TaskID int64 `json:"task_id,string"`
}

type ListRunsResponse struct {
	Code int32                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data []*ScheduleRunWithData `json:"data"`
}

type ScheduleRunWithData struct {
	*domain.AgentScheduleRun
	Input     string   `json:"input"`
	Output    string   `json:"output"`
	FollowUps []string `json:"follow_ups"`
}
