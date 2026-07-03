/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package schedule

type TaskStatus string

const (
	TaskStatusEnabled  TaskStatus = "enabled"
	TaskStatusDisabled TaskStatus = "disabled"
	TaskStatusDeleted  TaskStatus = "deleted"
)

type RunStatus string

const (
	RunStatusRunning RunStatus = "running"
	RunStatusSuccess RunStatus = "success"
	RunStatusFailed  RunStatus = "failed"
)

type TriggerType string

const (
	TriggerTypeCron    TriggerType = "cron"
	TriggerTypeOneTime TriggerType = "one_time"
)

type TaskKind string

const (
	TaskKindNotificationOnly TaskKind = "notification_only"
	TaskKindAgentTask        TaskKind = "agent_task"
	TaskKindUnknown          TaskKind = "unknown"
)

type AgentScheduleTask struct {
	ID             int64       `gorm:"column:id;primaryKey;autoIncrement:false" json:"id,string"`
	UserID         int64       `gorm:"column:user_id;not null;index:idx_user_agent_status" json:"user_id,string"`
	SpaceID        int64       `gorm:"column:space_id;not null" json:"space_id,string"`
	AgentID        int64       `gorm:"column:agent_id;not null;index:idx_user_agent_status" json:"agent_id,string"`
	ConnectorID    int64       `gorm:"column:connector_id;not null" json:"connector_id,string"`
	ConversationID int64       `gorm:"column:conversation_id;not null" json:"conversation_id,string"`
	IsDraft        bool        `gorm:"column:is_draft;not null;default:false" json:"is_draft"`
	Title          string      `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Prompt         string      `gorm:"column:prompt;type:text;not null" json:"prompt"`
	ScheduleText   string      `gorm:"column:schedule_text;type:varchar(512);not null" json:"schedule_text"`
	TaskKind       TaskKind    `gorm:"column:task_kind;type:varchar(32);not null;default:agent_task" json:"task_kind"`
	CronExpr       string      `gorm:"column:cron_expr;type:varchar(64);not null" json:"cron_expr"`
	TriggerType    TriggerType `gorm:"column:trigger_type;type:varchar(32);not null;default:cron" json:"trigger_type"`
	Timezone       string      `gorm:"column:timezone;type:varchar(64);not null" json:"timezone"`
	Status         TaskStatus  `gorm:"column:status;type:varchar(32);not null;index:idx_user_agent_status;index:idx_status_next_run" json:"status"`
	NextRunAt      int64       `gorm:"column:next_run_at;not null;index:idx_status_next_run" json:"next_run_at"`
	LastRunAt      int64       `gorm:"column:last_run_at;not null;default:0" json:"last_run_at"`
	CreatedAt      int64       `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt      int64       `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt      int64       `gorm:"column:deleted_at;not null;default:0" json:"deleted_at"`
}

func (AgentScheduleTask) TableName() string {
	return "agent_schedule_task"
}

type AgentScheduleRun struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement:false" json:"id,string"`
	TaskID         int64     `gorm:"column:task_id;not null;index:idx_task_started_at" json:"task_id,string"`
	RunStatus      RunStatus `gorm:"column:run_status;type:varchar(32);not null" json:"run_status"`
	ChatID         int64     `gorm:"column:chat_id;not null;default:0" json:"chat_id,string"`
	ConversationID int64     `gorm:"column:conversation_id;not null" json:"conversation_id,string"`
	StartedAt      int64     `gorm:"column:started_at;not null;index:idx_task_started_at" json:"started_at"`
	FinishedAt     int64     `gorm:"column:finished_at;not null;default:0" json:"finished_at"`
	ErrorMessage   string    `gorm:"column:error_message;type:text" json:"error_message"`
}

func (AgentScheduleRun) TableName() string {
	return "agent_schedule_run"
}

type ParsedSchedule struct {
	CronExpr        string      `json:"cron_expr"`
	TriggerType     TriggerType `json:"trigger_type"`
	RunAt           int64       `json:"run_at,omitempty"`
	Timezone        string      `json:"timezone"`
	NormalizedText  string      `json:"normalized_text"`
	Confidence      float64     `json:"confidence"`
	PreviewNextRuns []int64     `json:"preview_next_runs"`
}
