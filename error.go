package ssm

import "fmt"

type InvalidTransitionError struct {
	Event Event
	From  State
}

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("Invalid transition error [Event: %v, From: %v]",
		e.Event, e.From,
	)
}
