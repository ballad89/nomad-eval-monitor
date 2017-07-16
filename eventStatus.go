package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/client"
)

// formatTime formats the time to string based on RFC822
func formatTime(t time.Time) string {
	return t.Format("01/02/06 15:04:05 MST")
}

// formatUnixNanoTime is a helper for formatting time for output.
func formatUnixNanoTime(nano int64) string {
	t := time.Unix(0, nano)
	return formatTime(t)
}

// outputTaskStatus prints out a list of the most recent events for the given
// task state.
func formatTaskStatus(state *api.TaskState) []string {

	events := make([]string, len(state.Events))

	size := len(state.Events)
	for i, event := range state.Events {
		formatedTime := formatUnixNanoTime(event.Time)

		// Build up the description based on the event type.
		var desc string
		switch event.Type {
		case api.TaskSetup:
			desc = event.Message
		case api.TaskStarted:
			desc = "Task started by client"
		case api.TaskReceived:
			desc = "Task received by client"
		case api.TaskFailedValidation:
			if event.ValidationError != "" {
				desc = event.ValidationError
			} else {
				desc = "Validation of task failed"
			}
		case api.TaskSetupFailure:
			if event.SetupError != "" {
				desc = event.SetupError
			} else {
				desc = "Task setup failed"
			}
		case api.TaskDriverFailure:
			if event.DriverError != "" {
				desc = event.DriverError
			} else {
				desc = "Failed to start task"
			}
		case api.TaskDownloadingArtifacts:
			desc = "Client is downloading artifacts"
		case api.TaskArtifactDownloadFailed:
			if event.DownloadError != "" {
				desc = event.DownloadError
			} else {
				desc = "Failed to download artifacts"
			}
		case api.TaskKilling:
			if event.KillReason != "" {
				desc = fmt.Sprintf("Killing task: %v", event.KillReason)
			} else if event.KillTimeout != 0 {
				desc = fmt.Sprintf("Sent interrupt. Waiting %v before force killing", event.KillTimeout)
			} else {
				desc = "Sent interrupt"
			}
		case api.TaskKilled:
			if event.KillError != "" {
				desc = event.KillError
			} else {
				desc = "Task successfully killed"
			}
		case api.TaskTerminated:
			var parts []string
			parts = append(parts, fmt.Sprintf("Exit Code: %d", event.ExitCode))

			if event.Signal != 0 {
				parts = append(parts, fmt.Sprintf("Signal: %d", event.Signal))
			}

			if event.Message != "" {
				parts = append(parts, fmt.Sprintf("Exit Message: %q", event.Message))
			}
			desc = strings.Join(parts, ", ")
		case api.TaskRestarting:
			in := fmt.Sprintf("Task restarting in %v", time.Duration(event.StartDelay))
			if event.RestartReason != "" && event.RestartReason != client.ReasonWithinPolicy {
				desc = fmt.Sprintf("%s - %s", event.RestartReason, in)
			} else {
				desc = in
			}
		case api.TaskNotRestarting:
			if event.RestartReason != "" {
				desc = event.RestartReason
			} else {
				desc = "Task exceeded restart policy"
			}
		case api.TaskSiblingFailed:
			if event.FailedSibling != "" {
				desc = fmt.Sprintf("Task's sibling %q failed", event.FailedSibling)
			} else {
				desc = "Task's sibling failed"
			}
		case api.TaskSignaling:
			sig := event.TaskSignal
			reason := event.TaskSignalReason

			if sig == "" && reason == "" {
				desc = "Task being sent a signal"
			} else if sig == "" {
				desc = reason
			} else if reason == "" {
				desc = fmt.Sprintf("Task being sent signal %v", sig)
			} else {
				desc = fmt.Sprintf("Task being sent signal %v: %v", sig, reason)
			}
		case api.TaskRestartSignal:
			if event.RestartReason != "" {
				desc = event.RestartReason
			} else {
				desc = "Task signaled to restart"
			}
		case api.TaskDriverMessage:
			desc = event.DriverMessage
		case api.TaskLeaderDead:
			desc = "Leader Task in Group dead"
		}

		// Reverse order so we are sorted by time
		events[size-i-1] = fmt.Sprintf("%s | %s | %s\n", formatedTime, event.Type, desc)

	}

	return events
}
