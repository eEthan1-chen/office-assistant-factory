/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNaturalLanguage(t *testing.T) {
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, time.FixedZone("CST", 8*3600))
	tests := []struct {
		name string
		text string
		want string
	}{
		{name: "daily hour", text: "每天早上 9 点", want: "0 9 * * *"},
		{name: "weekly", text: "每周一上午 10 点", want: "0 10 * * 1"},
		{name: "workday", text: "每个工作日 18:00", want: "0 18 * * 1-5"},
		{name: "interval", text: "每隔 2 小时", want: "0 */2 * * *"},
		{name: "chinese number interval", text: "每三分钟提醒我：光头快写代码", want: "*/3 * * * *"},
		{name: "pm", text: "每天晚上 9 点", want: "0 21 * * *"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNaturalLanguage(tt.text, "", now)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.CronExpr)
			assert.Equal(t, "Asia/Shanghai", got.Timezone)
			assert.Len(t, got.PreviewNextRuns, 5)
		})
	}
}

func TestParseNaturalLanguageOneTimeReminder(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)

	got, err := ParseNaturalLanguage("请在2分钟后提醒我开会", "", now)
	require.NoError(t, err)
	assert.Equal(t, TriggerTypeOneTime, got.TriggerType)
	assert.Empty(t, got.CronExpr)
	assert.Equal(t, time.Date(2026, 6, 5, 8, 2, 0, 0, loc).UnixMilli(), got.RunAt)
	assert.Equal(t, []int64{got.RunAt}, got.PreviewNextRuns)
}

func TestParseNaturalLanguageOneTimeChineseNumber(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)

	got, err := ParseNaturalLanguage("三分钟后提醒我开会", "", now)
	require.NoError(t, err)
	assert.Equal(t, TriggerTypeOneTime, got.TriggerType)
	assert.Equal(t, time.Date(2026, 6, 5, 8, 3, 0, 0, loc).UnixMilli(), got.RunAt)
}

func TestParseNaturalLanguageOneTimeSendMessage(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)

	got, err := ParseNaturalLanguage("1分钟发给我汇总的现在涨停股票", "", now)
	require.NoError(t, err)
	assert.Equal(t, TriggerTypeOneTime, got.TriggerType)
	assert.Empty(t, got.CronExpr)
	assert.Equal(t, time.Date(2026, 6, 5, 8, 1, 0, 0, loc).UnixMilli(), got.RunAt)
	assert.Equal(t, []int64{got.RunAt}, got.PreviewNextRuns)
}

func TestExtractReminderPrompt(t *testing.T) {
	got, ok := ExtractReminderPrompt("请在2分钟后提醒我开会")
	require.True(t, ok)
	assert.Equal(t, "提醒我开会", got)

	got, ok = ExtractReminderPrompt("1分钟发给我汇总的现在涨停股票")
	require.True(t, ok)
	assert.Equal(t, "发给我汇总的现在涨停股票", got)
}

func TestParseScheduleIntentRegexFallbackNotificationOnly(t *testing.T) {
	t.Setenv("SCHEDULE_BUILTIN_CM_TYPE", "")
	t.Setenv("BUILTIN_CM_TYPE", "")
	t.Setenv("SCHEDULE_MODEL_ID", "")
	t.Setenv("MODEL_OPENCOZE_ID_0", "")
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)

	got, ok, err := ParseScheduleIntent(context.Background(), "过一分钟提醒我打球", "", now)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, got.Intent)
	require.NotNil(t, got.Parsed)
	assert.Equal(t, string(TaskKindNotificationOnly), got.Intent.TaskKind)
	assert.Equal(t, "打球", got.Intent.TaskPrompt)
	assert.Equal(t, TriggerTypeOneTime, got.Parsed.TriggerType)
	assert.Equal(t, time.Date(2026, 6, 5, 8, 1, 0, 0, loc).UnixMilli(), got.Parsed.RunAt)
}

func TestParseScheduleIntentRegexFallbackAgentTask(t *testing.T) {
	t.Setenv("SCHEDULE_BUILTIN_CM_TYPE", "")
	t.Setenv("BUILTIN_CM_TYPE", "")
	t.Setenv("SCHEDULE_MODEL_ID", "")
	t.Setenv("MODEL_OPENCOZE_ID_0", "")
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)

	got, ok, err := ParseScheduleIntent(context.Background(), "一分钟后提醒我有哪些任务", "", now)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, got.Intent)
	require.NotNil(t, got.Parsed)
	assert.Equal(t, string(TaskKindAgentTask), got.Intent.TaskKind)
	assert.Equal(t, "有哪些任务", got.Intent.TaskPrompt)
	assert.Equal(t, "有哪些任务", got.Intent.Title)
	assert.Equal(t, TriggerTypeOneTime, got.Parsed.TriggerType)
	assert.Equal(t, time.Date(2026, 6, 5, 8, 1, 0, 0, loc).UnixMilli(), got.Parsed.RunAt)
}

func TestNormalizeScheduleIntentConvertsSecondTimestamp(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)
	runAt := now.Add(time.Minute).Unix()

	got, ok, err := NormalizeScheduleIntent(&ScheduleIntent{
		IsScheduleRequest: true,
		TaskKind:          string(TaskKindAgentTask),
		TriggerType:       string(TriggerTypeOneTime),
		RunAt:             runAt,
		Timezone:          "Asia/Shanghai",
		Title:             "有哪些任务",
		TaskPrompt:        "提醒我有哪些任务",
		Confidence:        0.95,
	}, "", now)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, got.Intent)
	assert.Equal(t, now.Add(time.Minute).UnixMilli(), got.Intent.RunAt)
	assert.Equal(t, "有哪些任务", got.Intent.TaskPrompt)
}

func TestParseNaturalLanguageInvalid(t *testing.T) {
	_, err := ParseNaturalLanguage("下次有空的时候提醒我", "", time.Now())
	require.Error(t, err)
}

func TestNextRuns(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 6, 5, 8, 0, 0, 0, loc)
	runs, err := NextRuns("0 9 * * *", "Asia/Shanghai", now, 2)
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 6, 5, 9, 0, 0, 0, loc).UnixMilli(), runs[0])
	assert.Equal(t, time.Date(2026, 6, 6, 9, 0, 0, 0, loc).UnixMilli(), runs[1])
}
