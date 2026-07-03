/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"fmt"
	"strings"
	"time"
)

const (
	intentConfidenceThreshold = 0.8
	minOneTimeDelay           = 30 * time.Second
)

type ScheduleIntent struct {
	IsScheduleRequest bool    `json:"is_schedule_request"`
	TaskKind          string  `json:"task_kind"`
	TriggerType       string  `json:"trigger_type"`
	CronExpr          string  `json:"cron_expr"`
	RunAt             int64   `json:"run_at"`
	Timezone          string  `json:"timezone"`
	Title             string  `json:"title"`
	TaskPrompt        string  `json:"task_prompt"`
	Confidence        float64 `json:"confidence"`
}

type ScheduleIntentResult struct {
	Intent *ScheduleIntent
	Parsed *ParsedSchedule
}

func NormalizeScheduleIntent(intent *ScheduleIntent, fallbackTimezone string, now time.Time) (*ScheduleIntentResult, bool, error) {
	if intent == nil || !intent.IsScheduleRequest {
		return nil, false, nil
	}
	if intent.Confidence < intentConfidenceThreshold {
		return nil, false, nil
	}
	intent.TaskKind = normalizeTaskKind(intent.TaskKind)
	if intent.TaskKind == string(TaskKindUnknown) {
		return nil, false, nil
	}

	intent.Timezone = defaultTimezone(firstNonEmpty(intent.Timezone, fallbackTimezone))
	if _, err := loadLocation(intent.Timezone); err != nil {
		return nil, false, err
	}

	intent.Title = strings.TrimSpace(intent.Title)
	intent.TaskPrompt = NormalizeReminderPrompt(intent.TaskPrompt)
	if intent.TaskPrompt == "" {
		return nil, false, fmt.Errorf("schedule task prompt is required")
	}
	if intent.TaskKind == string(TaskKindNotificationOnly) && IsDynamicSchedulePrompt(intent.TaskPrompt) {
		intent.TaskKind = string(TaskKindAgentTask)
	}
	if intent.Title == "" {
		intent.Title = intent.TaskPrompt
	}

	switch TriggerType(intent.TriggerType) {
	case TriggerTypeCron:
		cronExpr := strings.TrimSpace(intent.CronExpr)
		if cronExpr == "" {
			return nil, false, fmt.Errorf("cron expression is required")
		}
		runs, err := NextRuns(cronExpr, intent.Timezone, now, 5)
		if err != nil {
			return nil, false, err
		}
		return &ScheduleIntentResult{
			Intent: intent,
			Parsed: &ParsedSchedule{
				CronExpr:        cronExpr,
				TriggerType:     TriggerTypeCron,
				Timezone:        intent.Timezone,
				NormalizedText:  intent.Title,
				Confidence:      intent.Confidence,
				PreviewNextRuns: runs,
			},
		}, true, nil
	case TriggerTypeOneTime:
		if intent.RunAt > 0 && intent.RunAt < 100000000000 {
			intent.RunAt *= 1000
		}
		if intent.RunAt <= now.Add(minOneTimeDelay).UnixMilli() {
			return nil, false, fmt.Errorf("one-time schedule must be in the future")
		}
		return &ScheduleIntentResult{
			Intent: intent,
			Parsed: &ParsedSchedule{
				TriggerType:     TriggerTypeOneTime,
				RunAt:           intent.RunAt,
				Timezone:        intent.Timezone,
				NormalizedText:  intent.Title,
				Confidence:      intent.Confidence,
				PreviewNextRuns: []int64{intent.RunAt},
			},
		}, true, nil
	default:
		return nil, false, fmt.Errorf("unsupported schedule trigger type: %s", intent.TriggerType)
	}
}

func NormalizeTaskKind(taskKind string) TaskKind {
	switch TaskKind(strings.TrimSpace(taskKind)) {
	case TaskKindNotificationOnly:
		return TaskKindNotificationOnly
	case TaskKindAgentTask:
		return TaskKindAgentTask
	default:
		return TaskKindAgentTask
	}
}

func normalizeTaskKind(taskKind string) string {
	switch TaskKind(strings.TrimSpace(taskKind)) {
	case TaskKindNotificationOnly:
		return string(TaskKindNotificationOnly)
	case TaskKindAgentTask:
		return string(TaskKindAgentTask)
	default:
		return string(TaskKindUnknown)
	}
}

func NormalizeReminderPrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	for _, prefix := range reminderPrefixes() {
		if strings.HasPrefix(prompt, prefix) {
			if action := strings.TrimSpace(strings.TrimPrefix(prompt, prefix)); action != "" {
				return strings.TrimLeft(action, "：:，,。 ")
			}
		}
	}
	return prompt
}

func IsDynamicSchedulePrompt(prompt string) bool {
	prompt = strings.ToLower(strings.TrimSpace(prompt))
	if prompt == "" {
		return false
	}
	dynamicPhrases := []string{
		"有哪些任务",
		"有什么任务",
		"查询任务",
		"查一下任务",
		"待办",
		"逾期任务",
		"到期任务",
		"任务列表",
		"汇总",
		"总结",
		"整理",
		"统计",
		"分析",
		"生成",
		"报告",
		"周报",
		"股票",
		"天气",
		"新闻",
		"航班",
		"日程",
		"会议",
		"工时",
		"项目风险",
	}
	for _, phrase := range dynamicPhrases {
		if strings.Contains(prompt, phrase) {
			return true
		}
	}
	return false
}

func reminderPrefixes() []string {
	return []string{"提醒我", "通知我", "叫我", "发给我", "发送给我", "推送给我", "告诉我", "发我", "给我发"}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
