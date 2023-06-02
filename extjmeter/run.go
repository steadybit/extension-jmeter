/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extjmeter

import (
	"context"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	extension_kit "github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extcmd"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/extfile"
	"github.com/steadybit/extension-kit/extutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type JmeterLoadTestRunAction struct{}

type JmeterLoadTestRunState struct {
	Command         []string  `json:"command"`
	Pid             int       `json:"pid"`
	CmdStateID      string    `json:"cmdStateId"`
	Timestamp       string    `json:"timestamp"`
	StdOutLineCount int       `json:"stdOutLineCount"`
	ExecutionId     uuid.UUID `json:"executionId"`
	CloudRunId      string    `json:"cloudRunId"`
}

// Make sure action implements all required interfaces
var (
	_ action_kit_sdk.Action[JmeterLoadTestRunState]           = (*JmeterLoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStatus[JmeterLoadTestRunState] = (*JmeterLoadTestRunAction)(nil)
	_ action_kit_sdk.ActionWithStop[JmeterLoadTestRunState]   = (*JmeterLoadTestRunAction)(nil)
)

func NewJmeterLoadTestRunAction() action_kit_sdk.Action[JmeterLoadTestRunState] {
	return &JmeterLoadTestRunAction{}
}

func (l *JmeterLoadTestRunAction) NewEmptyState() JmeterLoadTestRunState {
	return JmeterLoadTestRunState{}
}

func (l *JmeterLoadTestRunAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          actionId,
		Label:       "JMeter",
		Description: "Execute a JMeter load test.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Icon:        extutil.Ptr(actionIcon),
		Kind:        action_kit_api.LoadTest,
		TimeControl: action_kit_api.Internal,
		Hint: &action_kit_api.ActionHint{
			Content: "Please note that load tests are executed by the jmeter extension participating in the experiment, consuming resources of the system that it is installed in.",
			Type:    action_kit_api.HintWarning,
		},
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:        "file",
				Label:       "JMeter JMX File",
				Description: extutil.Ptr("Upload your JMeter Script"),
				Type:        action_kit_api.File,
				Required:    extutil.Ptr(true),
				AcceptedFileTypes: extutil.Ptr([]string{
					".jmx",
				}),
			},
			{
				Name:        "parameter",
				Label:       "JMeter Parameter",
				Description: extutil.Ptr("Parameters will be accessible from your JMeter Script by ${__P(FOOBAR)}"),
				Type:        action_kit_api.KeyValue,
				Required:    extutil.Ptr(true),
			},
		},
		Status: extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("5s"),
		}),
		Stop: extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

type JMeterLoadTestRunConfig struct {
	Parameter []map[string]string
	File      string
}

func (l *JmeterLoadTestRunAction) Prepare(_ context.Context, state *JmeterLoadTestRunState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var config JMeterLoadTestRunConfig
	if err := extconversion.Convert(request.Config, &config); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}
	logfile := fmt.Sprintf("/tmp/steadybit/%v/result.jtl", request.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	command := []string{
		"jmeter",
		"--nongui",
		"--testfile",
		config.File,
		"--logfile",
		logfile,
		"-Djmeter.save.saveservice.output_format=xml",
	}

	if config.Parameter != nil {
		for _, value := range config.Parameter {
			command = append(command, "--jmeterproperty")
			command = append(command, fmt.Sprintf("%s=%s", value["key"], value["value"]))
		}
	}

	state.ExecutionId = request.ExecutionId
	state.Timestamp = time.Now().Format(time.RFC3339)
	state.Command = command

	return nil, nil
}

func (l *JmeterLoadTestRunAction) Start(_ context.Context, state *JmeterLoadTestRunState) (*action_kit_api.StartResult, error) {
	log.Info().Msgf("Starting JMeter load test with command: %s", strings.Join(state.Command, " "))
	cmd := exec.Command(state.Command[0], state.Command[1:]...)
	cmdState := extcmd.NewCmdState(cmd)
	state.CmdStateID = cmdState.Id
	err := cmd.Start()
	if err != nil {
		return nil, extension_kit.ToError("Failed to start command.", err)
	}

	state.Pid = cmd.Process.Pid
	go func() {
		cmdErr := cmd.Wait()
		if cmdErr != nil {
			log.Error().Msgf("Failed to execute jmeter: %s", cmdErr)
		}
	}()
	log.Info().Msgf("Started load test.")

	state.Command = nil
	return nil, nil
}

func (l *JmeterLoadTestRunAction) Status(_ context.Context, state *JmeterLoadTestRunState) (*action_kit_api.StatusResult, error) {
	log.Debug().Msgf("Checking JMeter status for %d\n", state.Pid)

	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find command state", err)
	}

	var result action_kit_api.StatusResult

	// check if jmeter is still running
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	stdOut := cmdState.GetLines(false)
	stdOutToLog(stdOut)
	if exitCode == -1 {
		log.Debug().Msgf("JMeter is still running")
		result.Completed = false
	} else if exitCode == 0 {
		log.Info().Msgf("JMeter run completed successfully")
		result.Completed = true
	} else {
		title := fmt.Sprintf("JMeter run failed, exit-code %d", exitCode)
		result.Completed = true
		result.Error = &action_kit_api.ActionKitError{
			Status: extutil.Ptr(action_kit_api.Errored),
			Title:  title,
		}
	}

	filename := fmt.Sprintf("/tmp/steadybit/%v/log.txt", state.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	if err := extfile.AppendToFile(filename, stdOut); err != nil {
		return nil, extension_kit.ToError("Failed to append log to file", err)
	}
	messages := stdOutToMessages(stdOut)
	log.Debug().Msgf("Returning %d messages", len(messages))

	result.Messages = extutil.Ptr(messages)
	return &result, nil
}

func (l *JmeterLoadTestRunAction) Stop(_ context.Context, state *JmeterLoadTestRunState) (*action_kit_api.StopResult, error) {
	if state.CmdStateID == "" {
		log.Info().Msg("JMeter not yet started, nothing to stop.")
		return nil, nil
	}

	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find command state", err)
	}
	extcmd.RemoveCmdState(state.CmdStateID)

	// gracefully stop JMeter
	if err := exec.Command("stoptest.sh").Run(); err != nil {
		return nil, extension_kit.ToError("Failed to stop jmeter gracefully.", err)
	}

	// kill JMeter if it is still running
	var pid = state.Pid
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, extension_kit.ToError("Failed to find process", err)
	}
	_ = process.Kill()

	// read Stout and Stderr and send it as Messages
	stdOut := cmdState.GetLines(true)
	stdOutToLog(stdOut)
	filename := fmt.Sprintf("/tmp/steadybit/%v/log.txt", state.ExecutionId) //Folder is managed by action_kit_sdk's file download handling
	if err := extfile.AppendToFile(filename, stdOut); err != nil {
		return nil, extension_kit.ToError("Failed to append log to file", err)
	}
	messages := stdOutToMessages(stdOut)

	// read return code and send it as Message
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	if exitCode != 0 && exitCode != -1 {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Error),
			Message: fmt.Sprintf("JMeter run failed with exit code %d", exitCode),
		})
	}

	var artifacts []action_kit_api.Artifact

	// check if log file exists and send it as artifact
	_, err = os.Stat(filename)
	if err == nil { // file exists
		content, err := extfile.File2Base64(filename)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, action_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_log.txt",
			Data:  content,
		})
	}

	resultFilename := fmt.Sprintf("/tmp/steadybit/%v/result.jtl", state.ExecutionId)
	_, err = os.Stat(resultFilename)
	var resultFailure *action_kit_api.ActionKitError
	if err == nil { // file exists
		content, err := extfile.File2Base64(resultFilename)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, action_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_result.jtl",
			Data:  content,
		})

		//try to find assertion failures
		resultFile, err := os.Open(resultFilename)
		if err != nil {
			return nil, err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Error().Msgf("Failed to close file: %s", err)
			}
		}(resultFile)
		resultXml, err := xmlquery.Parse(resultFile)
		if err != nil {
			log.Error().Msgf("Failed to parse result xml: %s", err)
			return nil, err
		}
		failure := xmlquery.Find(resultXml, "//failureMessage[not(*) and normalize-space(.)]")
		if len(failure) > 0 {
			resultFailure = &action_kit_api.ActionKitError{
				Status: extutil.Ptr(action_kit_api.Failed),
				Title:  fmt.Sprintf("%d assertion failures found.", len(failure)),
			}
		}
	}

	log.Debug().Msgf("Returning %d messages", len(messages))
	return &action_kit_api.StopResult{
		Artifacts: extutil.Ptr(artifacts),
		Messages:  extutil.Ptr(messages),
		Error:     resultFailure,
	}, nil
}
