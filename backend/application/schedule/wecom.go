/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	domain "github.com/coze-dev/coze-studio/backend/domain/schedule"
)

const weComWebhookURLEnv = "WECOM_WEBHOOK_URL"

var weComHTTPClient = &http.Client{Timeout: 5 * time.Second}

type weComTextMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func notifyWeComForScheduleTask(ctx context.Context, task *domain.AgentScheduleTask, output string) error {
	webhookURL := strings.TrimSpace(os.Getenv(weComWebhookURLEnv))
	if webhookURL == "" {
		return nil
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return fmt.Errorf("schedule agent output is empty")
	}
	return sendWeComText(ctx, webhookURL, buildWeComScheduleContent(task, output))
}

func buildWeComScheduleContent(task *domain.AgentScheduleTask, output string) string {
	title := cleanScheduleTitle(task.Title)
	output = strings.TrimSpace(output)

	switch {
	case title != "":
		return fmt.Sprintf("%s\n\n%s", title, output)
	case output != "":
		return output
	default:
		return "定时任务执行完成，但未生成结果。"
	}
}

func cleanScheduleTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.TrimLeft(title, "：:，,。 ")
	return strings.TrimSpace(title)
}

func sendWeComText(ctx context.Context, webhookURL, content string) error {
	msg := weComTextMessage{MsgType: "text"}
	msg.Text.Content = content

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := weComHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("wecom webhook returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("decode wecom webhook response failed: %w", err)
		}
		if result.ErrCode != 0 {
			return fmt.Errorf("wecom webhook returned errcode %d: %s", result.ErrCode, result.ErrMsg)
		}
	}
	return nil
}
