package aggregate

import (
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
)

// activityLogger is the default logger for the Aggregate Activity
var activityLogger = logger.GetLogger("activity-tibco-aggregate")



const (
	ivFunction   = "function"
	ivWindowSize = "windowSize"
	ivAutoReset  = "autoReset"
	ivValue      = "value"

	ovResult = "result"
	ovReport = "report"
)

type Aggregator interface {
	Add(value float64) (report bool, result float64)
}

func init() {
	activityLogger.SetLogLevel(logger.InfoLevel)
}

// AggregateActivity is an Activity that is used to Aggregate a message to the console
// inputs : {function, windowSize, autoRest, value}
// outputs: {result, report}
type AggregateActivity struct {
	metadata *activity.Metadata
	mutex    *sync.RWMutex

	// aggregators stateful map of aggregators
	aggregators map[string]Aggregator
}

// NewActivity creates a new AppActivity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &AggregateActivity{metadata: metadata, aggregators:make(map[string]Aggregator), mutex:&sync.RWMutex{}}
}

// Metadata returns the activity's metadata
func (a *AggregateActivity) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements api.Activity.Eval - Aggregates the Message
func (a *AggregateActivity) Eval(context activity.Context) (done bool, err error) {

	aggregatorKey := context.FlowDetails().Name() + ":" + context.TaskName()

	a.mutex.RLock()
	//get aggregator for activity, assumes flow & task names are unique
	aggregator, ok := a.aggregators[aggregatorKey]

	a.mutex.RUnlock()

	//if window not create for this flow, create it

	if !ok {

		//go doesn't have lock upgrades or try, so do same check again

		a.mutex.Lock()
		aggregator, ok = a.aggregators[aggregatorKey]

		if !ok {
			windowSize, _ := context.GetInput(ivWindowSize).(int)
			autoReset, _ := context.GetInput(ivAutoReset).(bool)

			aggregator = NewMovingAverage(windowSize, autoReset)
			a.aggregators[aggregatorKey] = aggregator

			activityLogger.Debug("Aggregator created for ", aggregatorKey)
		}

		a.mutex.Unlock()
	}

	value, ok := context.GetInput(ivValue).(float64)

	if !ok {
		value,_ = data.CoerceToNumber(context.GetInput(ivValue))
	}

	report, result := aggregator.Add(value)

	context.SetOutput(ovReport, report)
	context.SetOutput(ovResult, result)

	return true, nil
}