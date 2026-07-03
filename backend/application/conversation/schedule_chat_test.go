/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package conversation

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coze-dev/coze-studio/backend/api/model/conversation/run"
)

func TestLastUserTextFromChatV3(t *testing.T) {
	got := lastUserTextFromChatV3(&run.ChatV3Request{
		AdditionalMessages: []*run.EnterMessage{
			{
				Role:        "assistant",
				ContentType: run.ContentTypeText,
				Content:     "你好",
			},
			{
				Role:        "user",
				ContentType: run.ContentTypeText,
				Content:     " 请在2分钟后提醒我开会 ",
			},
		},
	})

	assert.Equal(t, "请在2分钟后提醒我开会", got)
}

func TestLastUserTextFromChatV3MixApi(t *testing.T) {
	content, err := json.Marshal([]*run.AdditionalContent{
		{
			Type: "text",
			Text: ptrString("请在2分钟后"),
		},
		{
			Type: "text",
			Text: ptrString("提醒我开会"),
		},
	})
	require.NoError(t, err)

	got := lastUserTextFromChatV3(&run.ChatV3Request{
		AdditionalMessages: []*run.EnterMessage{
			{
				Role:        "user",
				ContentType: run.ContentTypeMixApi,
				Content:     string(content),
			},
		},
	})

	assert.Equal(t, "请在2分钟后 提醒我开会", got)
}

func TestReminderTitle(t *testing.T) {
	assert.Equal(t, "开会", reminderTitle("提醒我开会"))
	assert.Equal(t, "提交周报", reminderTitle("通知我提交周报"))
	assert.Equal(t, "汇总的现在涨停股票", reminderTitle("发给我汇总的现在涨停股票"))
	assert.Equal(t, "喝水", reminderTitle("喝水"))
}

func TestBuildScheduleToolResponseContent(t *testing.T) {
	content := buildScheduleToolResponseContent(&scheduleChatResult{
		taskID:  123,
		content: "已为你创建提醒：开会",
	})

	var got map[string]string
	require.NoError(t, json.Unmarshal([]byte(content), &got))
	assert.Equal(t, "Task created successfully", got["response_for_model"])
	assert.Equal(t, "123", got["schedule_task_id"])
	assert.Equal(t, "已为你创建提醒：开会", got["message"])
}

func ptrString(v string) *string {
	return &v
}
