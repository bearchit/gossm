package ssm

import (
	"testing"

	"fmt"

	"reflect"

	testify "github.com/stretchr/testify/assert"
)

var (
	sm *StateMachine
)

func init() {
	Reset()
}

type state string
type event string

const (
	EventAtoB event = "a-b"
	EventBtoC event = "b-c"
	EventLoop event = "loop"

	StateA state = "a"
	StateB state = "b"
	StateC state = "c"
)

type Info struct {
	Message string
}

func Reset() {
	sm = New(
		WithInitial(StateA),
		WithEvents(
			Events{
				{EventAtoB, States{StateA}, StateB},
				{EventBtoC, States{StateB}, StateC},
			},
		),
		WithLoops(
			LoopEvents{
				{EventLoop, States{StateA, StateB}},
			},
		),
		WithEventCallbacks(
			EventCallbacks{
				{Type: Before, Event: EventAtoB, Callback: func(current State, args ...interface{}) error {
					fmt.Printf("before_a-b: %v\n", current)
					return nil
				}},
				{Type: After, Event: EventAtoB, Callback: func(current State, args ...interface{}) error {
					fmt.Printf("after_a-b: %v\n", current)
					return nil
				}},
			},
		),
		WithStateCallbacks(
			StateCallbacks{
				{Type: Enter, State: StateB, Callback: func(current State, args ...interface{}) error {
					fmt.Printf("enter_b: %v\n", current)
					return nil
				}},
				{Type: Leave, State: StateB, Callback: func(current State, args ...interface{}) error {
					fmt.Printf("leave_b: %v\n", current)
					return nil
				}},
			},
		),
		WithAfterCallback(
			func(current State, args ...interface{}) error {
				fmt.Println(current)
				return nil
			},
		),
	)
}

func TestCan(t *testing.T) {
	assert := testify.New(t)

	Reset()

	assert.True(sm.Can(EventAtoB))
	assert.False(sm.Can(EventBtoC))
}

func TestTransition(t *testing.T) {
	assert := testify.New(t)

	Reset()

	assert.NoError(sm.Event(EventAtoB, Info{Message: "Hello"}))
	assert.Equal(StateB, sm.Current())

	assert.NoError(sm.Event(EventBtoC))
	assert.Equal(StateC, sm.Current())
}

func TestLoopTransition(t *testing.T) {
	assert := testify.New(t)

	Reset()

	assert.NoError(sm.Event(EventLoop))
	assert.Equal(StateA, sm.Current())

	assert.NoError(sm.Event(EventAtoB))
	assert.Equal(StateB, sm.Current())

	assert.NoError(sm.Event(EventLoop))
	assert.Equal(StateB, sm.Current())
}

func TestCustomTypeEquality(t *testing.T) {
	assert := testify.New(t)

	assert.NotEqual(StateA, "a")
	assert.Equal(StateA, state("a"))

	for k := range sm.cbEvent[Before] {
		t.Log(reflect.TypeOf(k))
	}
}
