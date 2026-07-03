/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/coze-dev/coze-studio/backend/domain/schedule"
)

func TestBuildWeComScheduleContent(t *testing.T) {
	assert.Equal(t, "开会\n\nAI 已完成会议提醒", buildWeComScheduleContent(&domain.AgentScheduleTask{
		Title:  "开会",
		Prompt: "开会",
	}, "AI 已完成会议提醒"))
	assert.Equal(t, "杯子的价格汇总结果。\n\n杯子价格区间为 10-30 元。", buildWeComScheduleContent(&domain.AgentScheduleTask{
		Title: "：杯子的价格汇总结果。",
	}, "杯子价格区间为 10-30 元。"))
	assert.Equal(t, "AI 输出", buildWeComScheduleContent(&domain.AgentScheduleTask{}, "AI 输出"))
}

func TestCleanScheduleTitle(t *testing.T) {
	assert.Equal(t, "杯子的价格汇总结果。", cleanScheduleTitle("：杯子的价格汇总结果。"))
	assert.Equal(t, "开会", cleanScheduleTitle("  : 开会"))
}

func TestSendWeComText(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	require.NoError(t, sendWeComText(context.Background(), server.URL, "定时提醒：开会"))
	assert.Equal(t, "text", got["msgtype"])
	text, ok := got["text"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "定时提醒：开会", text["content"])
}

func TestSendWeComTextReturnsWebhookError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":93000,"errmsg":"invalid webhook"}`))
	}))
	defer server.Close()

	err := sendWeComText(context.Background(), server.URL, "定时提醒：开会")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "errcode 93000")
}
