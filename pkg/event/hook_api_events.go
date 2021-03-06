package event

import (
	"time"

	"github.com/sirupsen/logrus"
)

// APIEvent represents simplified Event object to be returned from the API
type APIEvent struct {
	Time     time.Time
	LogLevel string `yaml:"level"`
	Message  string
}

// AsAPIEvents takes all buffered event log entries and saves them as APIEvents
func (eventLog *Log) AsAPIEvents() []*APIEvent {
	saver := &HookAPIEvents{}
	eventLog.Save(saver)
	return saver.events
}

// HookAPIEvents saves all events as APIEvents that holds only time, level and message
type HookAPIEvents struct {
	events []*APIEvent
}

// Levels defines on which log levels this hook should be fired
func (hook *HookAPIEvents) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire processes a single log entry
func (hook *HookAPIEvents) Fire(e *logrus.Entry) error {
	apiEvent := &APIEvent{Time: e.Time, LogLevel: e.Level.String(), Message: e.Message}
	hook.events = append(hook.events, apiEvent)
	return nil
}
