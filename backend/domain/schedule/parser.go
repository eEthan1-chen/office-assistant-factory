/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	timePattern      = regexp.MustCompile(`(?:(凌晨|早上|上午|中午|下午|晚上|晚间)?\s*(\d{1,2})(?:[:：点时](\d{1,2})?分?)?)`)
	numberPattern    = `[0-9]+|[一二两三四五六七八九十]+`
	intervalPattern  = regexp.MustCompile(`每(?:隔)?\s*(` + numberPattern + `)\s*(分钟|分|小时|个小时)`)
	afterPattern     = regexp.MustCompile(`(` + numberPattern + `)\s*(秒|分钟|分|小时|个小时|天|日)\s*(?:后|以后|之后)`)
	sendAfterPattern = regexp.MustCompile(`(` + numberPattern + `)\s*(秒|分钟|分|小时|个小时|天|日)\s*(?:后|以后|之后)?\s*(?:提醒我|通知我|叫我|发给我|发送给我|推送给我|告诉我|发我|给我发)`)
	weekdayMap       = map[string]int{"日": 0, "天": 0, "一": 1, "二": 2, "三": 3, "四": 4, "五": 5, "六": 6}
)

func ParseNaturalLanguage(text, timezone string, now time.Time) (*ParsedSchedule, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("schedule text is required")
	}
	tz := defaultTimezone(timezone)
	if _, err := loadLocation(tz); err != nil {
		return nil, err
	}

	if m := afterPattern.FindStringSubmatch(text); len(m) == 3 {
		return buildRelativeOneTimeParsed(m[1], m[2], tz, text, now)
	}

	if m := intervalPattern.FindStringSubmatch(text); len(m) == 3 {
		n, err := parsePositiveNumber(m[1])
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			return nil, fmt.Errorf("interval must be positive")
		}
		unit := m[2]
		var cron string
		if strings.Contains(unit, "小时") {
			if n > 23 {
				return nil, fmt.Errorf("hour interval must be <= 23")
			}
			cron = fmt.Sprintf("0 */%d * * *", n)
		} else {
			if n > 59 {
				return nil, fmt.Errorf("minute interval must be <= 59")
			}
			cron = fmt.Sprintf("*/%d * * * *", n)
		}
		return buildParsed(cron, tz, text, 0.9, now)
	}

	if m := sendAfterPattern.FindStringSubmatch(text); len(m) == 3 {
		return buildRelativeOneTimeParsed(m[1], m[2], tz, text, now)
	}

	hour, minute, ok := extractTime(text)
	if !ok {
		return nil, fmt.Errorf("无法解析定时表达，请改写为“每天 9 点”或“每周一 10:30”这类格式")
	}

	switch {
	case strings.Contains(text, "工作日"):
		return buildParsed(fmt.Sprintf("%d %d * * 1-5", minute, hour), tz, text, 0.95, now)
	case strings.Contains(text, "每周"):
		wd, ok := extractWeekday(text)
		if !ok {
			return nil, fmt.Errorf("每周任务需要包含周几")
		}
		return buildParsed(fmt.Sprintf("%d %d * * %d", minute, hour, wd), tz, text, 0.95, now)
	case strings.Contains(text, "每天") || strings.Contains(text, "每日") || strings.Contains(text, "天天"):
		return buildParsed(fmt.Sprintf("%d %d * * *", minute, hour), tz, text, 0.95, now)
	case strings.Contains(text, "每月"):
		day := 1
		if m := regexp.MustCompile(`每月\s*(\d{1,2})\s*[号日]`).FindStringSubmatch(text); len(m) == 2 {
			day, _ = strconv.Atoi(m[1])
		}
		if day < 1 || day > 31 {
			return nil, fmt.Errorf("day of month out of range")
		}
		return buildParsed(fmt.Sprintf("%d %d %d * *", minute, hour, day), tz, text, 0.85, now)
	default:
		return nil, fmt.Errorf("无法识别重复周期，请包含每天、每周、每月、工作日或每隔")
	}
}

func buildRelativeOneTimeParsed(nText, unit, timezone, normalized string, now time.Time) (*ParsedSchedule, error) {
	n, err := parsePositiveNumber(nText)
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, fmt.Errorf("relative time must be positive")
	}
	runAt := now.Add(relativeDuration(n, unit))
	return buildOneTimeParsed(runAt, timezone, normalized, 0.92)
}

func parsePositiveNumber(text string) (int, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0, fmt.Errorf("number is required")
	}
	if n, err := strconv.Atoi(text); err == nil {
		return n, nil
	}
	digits := map[rune]int{
		'一': 1,
		'二': 2,
		'两': 2,
		'三': 3,
		'四': 4,
		'五': 5,
		'六': 6,
		'七': 7,
		'八': 8,
		'九': 9,
	}
	if text == "十" {
		return 10, nil
	}
	runes := []rune(text)
	if len(runes) == 1 {
		if n, ok := digits[runes[0]]; ok {
			return n, nil
		}
	}
	if strings.ContainsRune(text, '十') {
		parts := strings.Split(text, "十")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid number: %s", text)
		}
		tens := 1
		if parts[0] != "" {
			r := []rune(parts[0])
			if len(r) != 1 {
				return 0, fmt.Errorf("invalid number: %s", text)
			}
			v, ok := digits[r[0]]
			if !ok {
				return 0, fmt.Errorf("invalid number: %s", text)
			}
			tens = v
		}
		ones := 0
		if parts[1] != "" {
			r := []rune(parts[1])
			if len(r) != 1 {
				return 0, fmt.Errorf("invalid number: %s", text)
			}
			v, ok := digits[r[0]]
			if !ok {
				return 0, fmt.Errorf("invalid number: %s", text)
			}
			ones = v
		}
		return tens*10 + ones, nil
	}
	return 0, fmt.Errorf("invalid number: %s", text)
}

func buildParsed(cron, timezone, normalized string, confidence float64, now time.Time) (*ParsedSchedule, error) {
	runs, err := NextRuns(cron, timezone, now, 5)
	if err != nil {
		return nil, err
	}
	return &ParsedSchedule{
		CronExpr:        cron,
		TriggerType:     TriggerTypeCron,
		Timezone:        timezone,
		NormalizedText:  normalized,
		Confidence:      confidence,
		PreviewNextRuns: runs,
	}, nil
}

func buildOneTimeParsed(runAt time.Time, timezone, normalized string, confidence float64) (*ParsedSchedule, error) {
	loc, err := loadLocation(timezone)
	if err != nil {
		return nil, err
	}
	runAt = runAt.In(loc)
	return &ParsedSchedule{
		TriggerType:     TriggerTypeOneTime,
		RunAt:           runAt.UnixMilli(),
		Timezone:        timezone,
		NormalizedText:  normalized,
		Confidence:      confidence,
		PreviewNextRuns: []int64{runAt.UnixMilli()},
	}, nil
}

func relativeDuration(n int, unit string) time.Duration {
	switch unit {
	case "秒":
		return time.Duration(n) * time.Second
	case "小时", "个小时":
		return time.Duration(n) * time.Hour
	case "天", "日":
		return time.Duration(n) * 24 * time.Hour
	default:
		return time.Duration(n) * time.Minute
	}
}

func ExtractReminderPrompt(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", false
	}
	for _, marker := range reminderPrefixes() {
		if idx := strings.Index(text, marker); idx >= 0 {
			prompt := strings.TrimSpace(text[idx:])
			if prompt != "" {
				return prompt, true
			}
		}
	}
	return "", false
}

func extractTime(text string) (int, int, bool) {
	matches := timePattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return 0, 0, false
	}
	m := matches[len(matches)-1]
	period := m[1]
	hour, _ := strconv.Atoi(m[2])
	minute := 0
	if m[3] != "" {
		minute, _ = strconv.Atoi(m[3])
	}
	if hour > 23 || minute > 59 {
		return 0, 0, false
	}
	if (period == "下午" || period == "晚上" || period == "晚间") && hour < 12 {
		hour += 12
	}
	if period == "中午" && hour < 11 {
		hour += 12
	}
	return hour, minute, true
}

func extractWeekday(text string) (int, bool) {
	m := regexp.MustCompile(`(?:周|星期|礼拜)([一二三四五六日天])`).FindStringSubmatch(text)
	if len(m) != 2 {
		return 0, false
	}
	wd, ok := weekdayMap[m[1]]
	return wd, ok
}
