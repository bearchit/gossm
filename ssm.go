package ssm

type (
	StateMachine struct {
		current     State
		transitions transitions
		cbEvent     eventCallbacks
		cbState     stateCallbacks
		cbAfter     AfterCallback
	}

	State  interface{}
	States []State

	Event interface{}

	transitions map[node]State

	node struct {
		event Event
		from  State
	}

	eventDesc struct {
		Event Event
		From  States
		To    State
	}

	Events []eventDesc

	loopDesc struct {
		Event Event
		Stay  States
	}

	LoopEvents []loopDesc

	callbackFn func(current State, args ...interface{}) error

	eventCallbacks map[int]map[Event]callbackFn
	stateCallbacks map[int]map[State]callbackFn

	eventCallbackDesc struct {
		Type     int
		Event    Event
		Callback callbackFn
	}

	stateCallbackDesc struct {
		Type     int
		State    State
		Callback callbackFn
	}

	EventCallbacks []eventCallbackDesc
	StateCallbacks []stateCallbackDesc
	AfterCallback  callbackFn
)

const (
	Before = iota + 1
	After
	Enter
	Leave
)

func New(options ...func(*StateMachine)) *StateMachine {
	sm := new(StateMachine)

	for _, option := range options {
		option(sm)
	}

	return sm
}

func WithInitial(state State) func(*StateMachine) {
	return func(m *StateMachine) {
		m.current = state
	}
}

func WithEvents(events Events) func(*StateMachine) {
	return func(m *StateMachine) {
		if m.transitions == nil {
			m.transitions = make(transitions)
		}

		for _, e := range events {
			for _, from := range e.From {
				m.transitions[node{e.Event, from}] = e.To
			}
		}
	}
}

func WithLoops(loops LoopEvents) func(*StateMachine) {
	return func(m *StateMachine) {
		if m.transitions == nil {
			m.transitions = make(transitions)
		}

		for _, e := range loops {
			for _, stay := range e.Stay {
				m.transitions[node{e.Event, stay}] = stay
			}
		}
	}
}

func WithEventCallbacks(callbacks EventCallbacks) func(*StateMachine) {
	return func(m *StateMachine) {
		if m.cbEvent == nil {
			m.cbEvent = make(eventCallbacks)
		}

		for _, cb := range callbacks {
			if m.cbEvent[cb.Type] == nil {
				m.cbEvent[cb.Type] = make(map[Event]callbackFn)
			}
			m.cbEvent[cb.Type][cb.Event] = cb.Callback
		}
	}
}

func WithStateCallbacks(callbacks StateCallbacks) func(*StateMachine) {
	return func(m *StateMachine) {
		if m.cbState == nil {
			m.cbState = make(stateCallbacks)
		}

		for _, cb := range callbacks {
			if m.cbState[cb.Type] == nil {
				m.cbState[cb.Type] = make(map[State]callbackFn)
			}
			m.cbState[cb.Type][cb.State] = cb.Callback
		}
	}
}

func WithAfterCallback(callback AfterCallback) func(*StateMachine) {
	return func(m *StateMachine) {
		m.cbAfter = callback
	}
}

func (sm *StateMachine) SetCurrent(state State) {
	sm.current = state
}

func (sm StateMachine) Current() State {
	return sm.current
}

func (sm *StateMachine) Event(e Event, args ...interface{}) error {
	dst, ok := sm.transitions[node{e, sm.Current()}]
	if !ok {
		return &InvalidTransitionError{Event: e, From: sm.Current()}
	}

	if cb, ok := sm.cbEvent[Before][e]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return err
		}
	}

	if cb, ok := sm.cbState[Enter][dst]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return err
		}
	}

	if cb, ok := sm.cbState[Leave][sm.Current()]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return err
		}
	}

	if dst == sm.Current() {
		return nil
	}

	sm.current = dst

	if cb, ok := sm.cbEvent[After][e]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return err
		}
	}

	if sm.cbAfter != nil {
		if err := sm.cbAfter(sm.Current(), args...); err != nil {
			return err
		}
	}

	return nil
}

func (sm StateMachine) Can(e Event, args ...interface{}) (bool, error) {
	dst, ok := sm.transitions[node{e, sm.Current()}]
	if !ok {
		return false, &InvalidTransitionError{Event: e, From: sm.Current()}
	}

	if cb, ok := sm.cbEvent[Before][e]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return false, err
		}
	}

	if cb, ok := sm.cbState[Enter][dst]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return false, err
		}
	}

	if cb, ok := sm.cbState[Leave][sm.Current()]; ok {
		if err := cb(sm.Current(), args...); err != nil {
			return false, err
		}
	}

	return true, nil
}
