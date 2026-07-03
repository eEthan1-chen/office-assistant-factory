/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

const parseScheduleIntentToolName = "parse_schedule_intent"

func ParseScheduleIntent(ctx context.Context, text, timezone string, now time.Time) (*ScheduleIntentResult, bool, error) {
	intent, ok, err := ParseScheduleIntentWithLLM(ctx, text, timezone, now)
	if err != nil {
		logs.CtxWarnf(ctx, "parse schedule intent with llm failed: %v", err)
		return parseScheduleIntentWithRegex(text, timezone, now)
	}
	if !ok {
		return parseScheduleIntentWithRegex(text, timezone, now)
	}
	result, normalized, err := NormalizeScheduleIntent(intent, timezone, now)
	if err != nil {
		logs.CtxWarnf(ctx, "normalize schedule intent from llm failed: %v", err)
		return parseScheduleIntentWithRegex(text, timezone, now)
	}
	if !normalized {
		return parseScheduleIntentWithRegex(text, timezone, now)
	}
	return result, true, nil
}

func ParseScheduleIntentWithLLM(ctx context.Context, text, timezone string, now time.Time) (*ScheduleIntent, bool, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, false, nil
	}
	chatModel, configured, err := scheduleChatModel(ctx)
	if err != nil {
		return nil, false, err
	}
	if !configured || chatModel == nil {
		return nil, false, nil
	}

	toolCallingModel, ok := chatModel.(model.ToolCallingChatModel)
	if ok {
		toolModel, err := toolCallingModel.WithTools([]*schema.ToolInfo{scheduleIntentToolInfo()})
		if err != nil {
			return nil, false, err
		}
		out, err := toolModel.Generate(ctx, scheduleIntentMessages(text, timezone, now), model.WithToolChoice(schema.ToolChoiceAllowed))
		if err != nil {
			return nil, false, err
		}
		if intent, ok, err := intentFromModelOutput(out); err != nil || ok {
			return intent, ok, err
		}
		return nil, false, nil
	}

	out, err := chatModel.Generate(ctx, scheduleIntentMessages(text, timezone, now))
	if err != nil {
		return nil, false, err
	}
	return intentFromModelOutput(out)
}

func scheduleChatModel(ctx context.Context) (model.BaseChatModel, bool, error) {
	modelIDText := strings.TrimSpace(firstNonEmpty(os.Getenv("SCHEDULE_MODEL_ID"), os.Getenv("MODEL_OPENCOZE_ID_0")))
	if modelIDText != "" {
		modelID, err := strconv.ParseInt(modelIDText, 10, 64)
		if err != nil {
			return nil, false, fmt.Errorf("invalid schedule model id %q: %w", modelIDText, err)
		}
		chatModel, _, err := modelbuilder.BuildModelByID(ctx, modelID, nil)
		if err != nil {
			return nil, false, err
		}
		return chatModel, true, nil
	}

	if strings.TrimSpace(os.Getenv("SCHEDULE_BUILTIN_CM_TYPE")) == "" && strings.TrimSpace(os.Getenv("BUILTIN_CM_TYPE")) == "" {
		return nil, false, nil
	}

	return modelbuilder.GetBuiltinChatModel(ctx, "SCHEDULE_")
}

func scheduleIntentMessages(text, timezone string, now time.Time) []*schema.Message {
	tz := defaultTimezone(timezone)
	systemPrompt := `You classify and parse schedule creation requests.
Use the parse_schedule_intent tool when the user clearly wants to create a reminder, notification, recurring task, or scheduled send.
If it is not a schedule creation request, return JSON {"is_schedule_request":false,"confidence":0}.
Only create schedules when there is an explicit trigger time or recurring rule and a concrete task to perform.
Classify task_kind as notification_only when the scheduled action is only sending a fixed reminder text and does not need search, analysis, summarization, generation, or tools.
Classify task_kind as agent_task when the scheduled action needs an agent to query information, call tools, analyze, summarize, generate, inspect data, or prepare a dynamic result.
Use unknown when unsure.
Use 5-field cron: minute hour day month weekday. Use one_time with run_at as Unix milliseconds for one-off tasks.
Do not invent missing times. Do not create schedules in the past.`
	userPrompt := fmt.Sprintf("Current time: %s\nDefault timezone: %s\nUser message: %s", now.In(mustLoadLocation(tz)).Format("2006-01-02 15:04:05 MST"), tz, text)
	return []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}
}

func scheduleIntentToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "parse_schedule_intent",
		Desc: "Return structured schedule intent parsed from the user's message.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"is_schedule_request": {Type: schema.Boolean, Desc: "Whether the user is asking to create a scheduled task.", Required: true},
			"task_kind":           {Type: schema.String, Desc: "notification_only for fixed reminders, agent_task for dynamic work, unknown when unsure.", Enum: []string{"notification_only", "agent_task", "unknown"}, Required: true},
			"trigger_type":        {Type: schema.String, Desc: "cron or one_time. Empty when is_schedule_request is false.", Enum: []string{"cron", "one_time"}, Required: false},
			"cron_expr":           {Type: schema.String, Desc: "5-field cron expression for recurring schedules.", Required: false},
			"run_at":              {Type: schema.Integer, Desc: "Unix milliseconds for one-time schedules.", Required: false},
			"timezone":            {Type: schema.String, Desc: "IANA timezone, for example Asia/Shanghai.", Required: true},
			"title":               {Type: schema.String, Desc: "Short title for the scheduled task.", Required: false},
			"task_prompt":         {Type: schema.String, Desc: "The task content to run when the schedule fires.", Required: false},
			"confidence":          {Type: schema.Number, Desc: "Confidence from 0 to 1.", Required: true},
		}),
	}
}

func intentFromModelOutput(out *schema.Message) (*ScheduleIntent, bool, error) {
	if out == nil {
		return nil, false, nil
	}
	for _, toolCall := range out.ToolCalls {
		if toolCall.Function.Name != parseScheduleIntentToolName {
			continue
		}
		intent, err := decodeScheduleIntent(toolCall.Function.Arguments)
		if err != nil {
			return nil, false, err
		}
		return intent, intent.IsScheduleRequest, nil
	}
	intent, err := decodeScheduleIntent(out.Content)
	if err != nil {
		return nil, false, nil
	}
	return intent, intent.IsScheduleRequest, nil
}

func decodeScheduleIntent(content string) (*ScheduleIntent, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("empty schedule intent")
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var intent ScheduleIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, err
	}
	return &intent, nil
}

func mustLoadLocation(timezone string) *time.Location {
	loc, err := loadLocation(timezone)
	if err != nil {
		return time.Local
	}
	return loc
}

func parseScheduleIntentWithRegex(text, timezone string, now time.Time) (*ScheduleIntentResult, bool, error) {
	prompt, ok := ExtractReminderPrompt(text)
	if !ok {
		return nil, false, nil
	}
	parsed, err := ParseNaturalLanguage(text, timezone, now)
	if err != nil {
		return nil, false, nil
	}
	taskPrompt := NormalizeReminderPrompt(prompt)
	taskKind := string(TaskKindNotificationOnly)
	if IsDynamicSchedulePrompt(taskPrompt) {
		taskKind = string(TaskKindAgentTask)
	}
	title := reminderTitleFromPrompt(taskPrompt)
	return &ScheduleIntentResult{
		Intent: &ScheduleIntent{
			IsScheduleRequest: true,
			TaskKind:          taskKind,
			TriggerType:       string(parsed.TriggerType),
			CronExpr:          parsed.CronExpr,
			RunAt:             parsed.RunAt,
			Timezone:          parsed.Timezone,
			Title:             title,
			TaskPrompt:        taskPrompt,
			Confidence:        parsed.Confidence,
		},
		Parsed: parsed,
	}, true, nil
}

func reminderTitleFromPrompt(prompt string) string {
	return NormalizeReminderPrompt(prompt)
}
