/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 */

package schedule

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"
)

type cronSpec struct {
	Minute     map[int]bool
	Hour       map[int]bool
	DayOfMonth map[int]bool
	Month      map[int]bool
	DayOfWeek  map[int]bool
}

func ValidateCron(expr string) error {
	_, err := parseCron(expr)
	return err
}

func NextRuns(expr, timezone string, from time.Time, count int) ([]int64, error) {
	spec, err := parseCron(expr)
	if err != nil {
		return nil, err
	}
	loc, err := loadLocation(timezone)
	if err != nil {
		return nil, err
	}
	if count <= 0 {
		count = 1
	}

	cursor := from.In(loc).Truncate(time.Minute).Add(time.Minute)
	deadline := cursor.AddDate(2, 0, 0)
	runs := make([]int64, 0, count)
	for len(runs) < count && cursor.Before(deadline) {
		if spec.match(cursor) {
			runs = append(runs, cursor.UnixMilli())
		}
		cursor = cursor.Add(time.Minute)
	}
	if len(runs) == 0 {
		return nil, fmt.Errorf("cron has no next run within two years")
	}
	return runs, nil
}

func (c cronSpec) match(t time.Time) bool {
	return c.Minute[t.Minute()] &&
		c.Hour[t.Hour()] &&
		c.DayOfMonth[t.Day()] &&
		c.Month[int(t.Month())] &&
		c.DayOfWeek[int(t.Weekday())]
}

func parseCron(expr string) (*cronSpec, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron expression must contain 5 fields")
	}
	spec := &cronSpec{}
	var err error
	if spec.Minute, err = parseCronField(fields[0], 0, 59); err != nil {
		return nil, fmt.Errorf("minute: %w", err)
	}
	if spec.Hour, err = parseCronField(fields[1], 0, 23); err != nil {
		return nil, fmt.Errorf("hour: %w", err)
	}
	if spec.DayOfMonth, err = parseCronField(fields[2], 1, 31); err != nil {
		return nil, fmt.Errorf("day of month: %w", err)
	}
	if spec.Month, err = parseCronField(fields[3], 1, 12); err != nil {
		return nil, fmt.Errorf("month: %w", err)
	}
	if spec.DayOfWeek, err = parseCronField(fields[4], 0, 6); err != nil {
		return nil, fmt.Errorf("day of week: %w", err)
	}
	return spec, nil
}

func parseCronField(field string, min, max int) (map[int]bool, error) {
	out := map[int]bool{}
	if field == "*" {
		for i := min; i <= max; i++ {
			out[i] = true
		}
		return out, nil
	}
	for _, part := range strings.Split(field, ",") {
		if strings.Contains(part, "/") {
			pair := strings.Split(part, "/")
			if len(pair) != 2 {
				return nil, fmt.Errorf("invalid step %q", part)
			}
			step, err := strconv.Atoi(pair[1])
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid step %q", part)
			}
			start, end := min, max
			if pair[0] != "*" {
				bounds := strings.Split(pair[0], "-")
				if len(bounds) != 2 {
					return nil, fmt.Errorf("invalid step range %q", part)
				}
				start, err = strconv.Atoi(bounds[0])
				if err != nil {
					return nil, err
				}
				end, err = strconv.Atoi(bounds[1])
				if err != nil {
					return nil, err
				}
			}
			if start < min || end > max || start > end {
				return nil, fmt.Errorf("range out of bounds %q", part)
			}
			for i := start; i <= end; i += step {
				out[i] = true
			}
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid range %q", part)
			}
			start, err := strconv.Atoi(bounds[0])
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(bounds[1])
			if err != nil {
				return nil, err
			}
			if start < min || end > max || start > end {
				return nil, fmt.Errorf("range out of bounds %q", part)
			}
			for i := start; i <= end; i++ {
				out[i] = true
			}
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid value %q", part)
		}
		if v < min || v > max {
			return nil, fmt.Errorf("value %d out of bounds", v)
		}
		out[v] = true
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty field")
	}
	return out, nil
}

func sortedCronValues(values map[int]bool) []int {
	out := make([]int, 0, len(values))
	for v := range values {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func defaultTimezone(tz string) string {
	if strings.TrimSpace(tz) == "" {
		return "Asia/Shanghai"
	}
	return strings.TrimSpace(tz)
}

func loadLocation(tz string) (*time.Location, error) {
	return time.LoadLocation(defaultTimezone(tz))
}
