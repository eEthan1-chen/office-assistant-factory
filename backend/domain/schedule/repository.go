/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

type Repository interface {
	Migrate(ctx context.Context) error
	CreateTask(ctx context.Context, task *AgentScheduleTask) error
	UpdateTask(ctx context.Context, task *AgentScheduleTask) error
	GetTask(ctx context.Context, id int64) (*AgentScheduleTask, error)
	ListTasks(ctx context.Context, userID, agentID int64) ([]*AgentScheduleTask, error)
	DeleteTask(ctx context.Context, id, userID int64, now int64) error
	DueTasks(ctx context.Context, now int64, limit int) ([]*AgentScheduleTask, error)
	ClaimTask(ctx context.Context, id, oldNextRunAt, nextRunAt, now int64) (bool, error)
	CreateRun(ctx context.Context, run *AgentScheduleRun) error
	FinishRun(ctx context.Context, id int64, status RunStatus, chatID int64, errMsg string, finishedAt int64) error
	ListRuns(ctx context.Context, taskID int64, limit int) ([]*AgentScheduleRun, error)
	GenID(ctx context.Context) (int64, error)
}

type gormRepository struct {
	db    *gorm.DB
	idGen idgen.IDGenerator
}

func NewRepository(db *gorm.DB, idGen idgen.IDGenerator) Repository {
	return &gormRepository{db: db, idGen: idGen}
}

func (r *gormRepository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(&AgentScheduleTask{}, &AgentScheduleRun{})
}

func (r *gormRepository) GenID(ctx context.Context) (int64, error) {
	return r.idGen.GenID(ctx)
}

func (r *gormRepository) CreateTask(ctx context.Context, task *AgentScheduleTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *gormRepository) UpdateTask(ctx context.Context, task *AgentScheduleTask) error {
	return r.db.WithContext(ctx).Model(&AgentScheduleTask{}).
		Where("id = ? AND user_id = ? AND status <> ?", task.ID, task.UserID, TaskStatusDeleted).
		Updates(map[string]any{
			"title":         task.Title,
			"prompt":        task.Prompt,
			"schedule_text": task.ScheduleText,
			"task_kind":     task.TaskKind,
			"is_draft":      task.IsDraft,
			"cron_expr":     task.CronExpr,
			"trigger_type":  task.TriggerType,
			"timezone":      task.Timezone,
			"status":        task.Status,
			"next_run_at":   task.NextRunAt,
			"updated_at":    task.UpdatedAt,
		}).Error
}

func (r *gormRepository) GetTask(ctx context.Context, id int64) (*AgentScheduleTask, error) {
	var task AgentScheduleTask
	err := r.db.WithContext(ctx).Where("id = ? AND status <> ?", id, TaskStatusDeleted).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *gormRepository) ListTasks(ctx context.Context, userID, agentID int64) ([]*AgentScheduleTask, error) {
	var tasks []*AgentScheduleTask
	tx := r.db.WithContext(ctx).Where("user_id = ? AND status <> ?", userID, TaskStatusDeleted)
	if agentID > 0 {
		tx = tx.Where("agent_id = ?", agentID)
	}
	err := tx.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *gormRepository) DeleteTask(ctx context.Context, id, userID int64, now int64) error {
	return r.db.WithContext(ctx).Model(&AgentScheduleTask{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{"status": TaskStatusDeleted, "deleted_at": now, "updated_at": now}).Error
}

func (r *gormRepository) DueTasks(ctx context.Context, now int64, limit int) ([]*AgentScheduleTask, error) {
	var tasks []*AgentScheduleTask
	if limit <= 0 {
		limit = 20
	}
	err := r.db.WithContext(ctx).
		Where("status = ? AND next_run_at > 0 AND next_run_at <= ?", TaskStatusEnabled, now).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *gormRepository) ClaimTask(ctx context.Context, id, oldNextRunAt, nextRunAt, now int64) (bool, error) {
	status := TaskStatusEnabled
	if nextRunAt == 0 {
		status = TaskStatusDisabled
	}
	tx := r.db.WithContext(ctx).Model(&AgentScheduleTask{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND status = ? AND next_run_at = ?", id, TaskStatusEnabled, oldNextRunAt).
		Updates(map[string]any{"next_run_at": nextRunAt, "last_run_at": now, "updated_at": now, "status": status})
	return tx.RowsAffected > 0, tx.Error
}

func (r *gormRepository) CreateRun(ctx context.Context, run *AgentScheduleRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *gormRepository) FinishRun(ctx context.Context, id int64, status RunStatus, chatID int64, errMsg string, finishedAt int64) error {
	return r.db.WithContext(ctx).Model(&AgentScheduleRun{}).Where("id = ?", id).
		Updates(map[string]any{
			"run_status":    status,
			"chat_id":       chatID,
			"error_message": errMsg,
			"finished_at":   finishedAt,
		}).Error
}

func (r *gormRepository) ListRuns(ctx context.Context, taskID int64, limit int) ([]*AgentScheduleRun, error) {
	var runs []*AgentScheduleRun
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("started_at DESC").Limit(limit).Find(&runs).Error
	return runs, err
}

func nowMillis() int64 {
	return time.Now().UnixMilli()
}
