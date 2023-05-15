// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"github.com/ethereum/go-ethereum/core/types"
	"pkg.furychain.dev/gridiron/cosmos/x/evm/plugins/state/events"
	"sync"
)

// Ensure, that LogsDBMock does implement events.LogsDB.
// If this is not the case, regenerate this file with moq.
var _ events.LogsDB = &LogsDBMock{}

// LogsDBMock is a mock implementation of events.LogsDB.
//
//	func TestSomethingThatUsesLogsDB(t *testing.T) {
//
//		// make and configure a mocked events.LogsDB
//		mockedLogsDB := &LogsDBMock{
//			AddLogFunc: func(log *types.Log)  {
//				panic("mock out the AddLog method")
//			},
//		}
//
//		// use mockedLogsDB in code that requires events.LogsDB
//		// and then make assertions.
//
//	}
type LogsDBMock struct {
	// AddLogFunc mocks the AddLog method.
	AddLogFunc func(log *types.Log)

	// calls tracks calls to the methods.
	calls struct {
		// AddLog holds details about calls to the AddLog method.
		AddLog []struct {
			// Log is the log argument value.
			Log *types.Log
		}
	}
	lockAddLog sync.RWMutex
}

// AddLog calls AddLogFunc.
func (mock *LogsDBMock) AddLog(log *types.Log) {
	if mock.AddLogFunc == nil {
		panic("LogsDBMock.AddLogFunc: method is nil but LogsDB.AddLog was just called")
	}
	callInfo := struct {
		Log *types.Log
	}{
		Log: log,
	}
	mock.lockAddLog.Lock()
	mock.calls.AddLog = append(mock.calls.AddLog, callInfo)
	mock.lockAddLog.Unlock()
	mock.AddLogFunc(log)
}

// AddLogCalls gets all the calls that were made to AddLog.
// Check the length with:
//
//	len(mockedLogsDB.AddLogCalls())
func (mock *LogsDBMock) AddLogCalls() []struct {
	Log *types.Log
} {
	var calls []struct {
		Log *types.Log
	}
	mock.lockAddLog.RLock()
	calls = mock.calls.AddLog
	mock.lockAddLog.RUnlock()
	return calls
}
