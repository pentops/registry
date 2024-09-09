// Code generated by protoc-gen-go-psm. DO NOT EDIT.

package github_pb

import (
	context "context"
	fmt "fmt"
	psm_j5pb "github.com/pentops/j5/gen/j5/state/v1/psm_j5pb"
	psm "github.com/pentops/protostate/psm"
	sqrlx "github.com/pentops/sqrlx.go/sqrlx"
)

// PSM RepoPSM

type RepoPSM = psm.StateMachine[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
]

type RepoPSMDB = psm.DBStateMachine[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
]

type RepoPSMEventSpec = psm.EventSpec[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
]

type RepoPSMEventKey = string

const (
	RepoPSMEventNil             RepoPSMEventKey = "<nil>"
	RepoPSMEventConfigure       RepoPSMEventKey = "configure"
	RepoPSMEventConfigureBranch RepoPSMEventKey = "configure_branch"
	RepoPSMEventRemoveBranch    RepoPSMEventKey = "remove_branch"
)

// EXTEND RepoKeys with the psm.IKeyset interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoKeys) PSMIsSet() bool {
	return msg != nil
}

// PSMFullName returns the full name of state machine with package prefix
func (msg *RepoKeys) PSMFullName() string {
	return "j5.registry.github.v1.repo"
}
func (msg *RepoKeys) PSMKeyValues() (map[string]string, error) {
	keyset := map[string]string{
		"owner": msg.Owner,
		"name":  msg.Name,
	}
	return keyset, nil
}

// EXTEND RepoState with the psm.IState interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoState) PSMIsSet() bool {
	return msg != nil
}

func (msg *RepoState) PSMMetadata() *psm_j5pb.StateMetadata {
	if msg.Metadata == nil {
		msg.Metadata = &psm_j5pb.StateMetadata{}
	}
	return msg.Metadata
}

func (msg *RepoState) PSMKeys() *RepoKeys {
	return msg.Keys
}

func (msg *RepoState) SetStatus(status RepoStatus) {
	msg.Status = status
}

func (msg *RepoState) SetPSMKeys(inner *RepoKeys) {
	msg.Keys = inner
}

func (msg *RepoState) PSMData() *RepoStateData {
	if msg.Data == nil {
		msg.Data = &RepoStateData{}
	}
	return msg.Data
}

// EXTEND RepoStateData with the psm.IStateData interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoStateData) PSMIsSet() bool {
	return msg != nil
}

// EXTEND RepoEvent with the psm.IEvent interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoEvent) PSMIsSet() bool {
	return msg != nil
}

func (msg *RepoEvent) PSMMetadata() *psm_j5pb.EventMetadata {
	if msg.Metadata == nil {
		msg.Metadata = &psm_j5pb.EventMetadata{}
	}
	return msg.Metadata
}

func (msg *RepoEvent) PSMKeys() *RepoKeys {
	return msg.Keys
}

func (msg *RepoEvent) SetPSMKeys(inner *RepoKeys) {
	msg.Keys = inner
}

// PSMEventKey returns the RepoPSMEventPSMEventKey for the event, implementing psm.IEvent
func (msg *RepoEvent) PSMEventKey() RepoPSMEventKey {
	tt := msg.UnwrapPSMEvent()
	if tt == nil {
		return RepoPSMEventNil
	}
	return tt.PSMEventKey()
}

// UnwrapPSMEvent implements psm.IEvent, returning the inner event message
func (msg *RepoEvent) UnwrapPSMEvent() RepoPSMEvent {
	if msg == nil {
		return nil
	}
	if msg.Event == nil {
		return nil
	}
	switch v := msg.Event.Type.(type) {
	case *RepoEventType_Configure_:
		return v.Configure
	case *RepoEventType_ConfigureBranch_:
		return v.ConfigureBranch
	case *RepoEventType_RemoveBranch_:
		return v.RemoveBranch
	default:
		return nil
	}
}

// SetPSMEvent sets the inner event message from a concrete type, implementing psm.IEvent
func (msg *RepoEvent) SetPSMEvent(inner RepoPSMEvent) error {
	if msg.Event == nil {
		msg.Event = &RepoEventType{}
	}
	switch v := inner.(type) {
	case *RepoEventType_Configure:
		msg.Event.Type = &RepoEventType_Configure_{Configure: v}
	case *RepoEventType_ConfigureBranch:
		msg.Event.Type = &RepoEventType_ConfigureBranch_{ConfigureBranch: v}
	case *RepoEventType_RemoveBranch:
		msg.Event.Type = &RepoEventType_RemoveBranch_{RemoveBranch: v}
	default:
		return fmt.Errorf("invalid type %T for RepoEventType", v)
	}
	return nil
}

type RepoPSMEvent interface {
	psm.IInnerEvent
	PSMEventKey() RepoPSMEventKey
}

// EXTEND RepoEventType_Configure with the RepoPSMEvent interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoEventType_Configure) PSMIsSet() bool {
	return msg != nil
}

func (*RepoEventType_Configure) PSMEventKey() RepoPSMEventKey {
	return RepoPSMEventConfigure
}

// EXTEND RepoEventType_ConfigureBranch with the RepoPSMEvent interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoEventType_ConfigureBranch) PSMIsSet() bool {
	return msg != nil
}

func (*RepoEventType_ConfigureBranch) PSMEventKey() RepoPSMEventKey {
	return RepoPSMEventConfigureBranch
}

// EXTEND RepoEventType_RemoveBranch with the RepoPSMEvent interface

// PSMIsSet is a helper for != nil, which does not work with generic parameters
func (msg *RepoEventType_RemoveBranch) PSMIsSet() bool {
	return msg != nil
}

func (*RepoEventType_RemoveBranch) PSMEventKey() RepoPSMEventKey {
	return RepoPSMEventRemoveBranch
}

func RepoPSMBuilder() *psm.StateMachineConfig[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
] {
	return &psm.StateMachineConfig[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
	]{}
}

func RepoPSMMutation[SE RepoPSMEvent](cb func(*RepoStateData, SE) error) psm.TransitionMutation[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
	SE,             // Specific event type for the transition
] {
	return psm.TransitionMutation[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
		SE,             // Specific event type for the transition
	](cb)
}

type RepoPSMHookBaton = psm.HookBaton[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
]

func RepoPSMLogicHook[SE RepoPSMEvent](cb func(context.Context, RepoPSMHookBaton, *RepoState, SE) error) psm.TransitionLogicHook[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
	SE,             // Specific event type for the transition
] {
	return psm.TransitionLogicHook[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
		SE,             // Specific event type for the transition
	](cb)
}
func RepoPSMDataHook[SE RepoPSMEvent](cb func(context.Context, sqrlx.Transaction, *RepoState, SE) error) psm.TransitionDataHook[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
	SE,             // Specific event type for the transition
] {
	return psm.TransitionDataHook[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
		SE,             // Specific event type for the transition
	](cb)
}
func RepoPSMLinkHook[SE RepoPSMEvent, DK psm.IKeyset, DIE psm.IInnerEvent](
	linkDestination psm.LinkDestination[DK, DIE],
	cb func(context.Context, *RepoState, SE, func(DK, DIE)) error,
) psm.LinkHook[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
	SE,             // Specific event type for the transition
	DK,             // Destination Keys
	DIE,            // Destination Inner Event
] {
	return psm.LinkHook[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
		SE,             // Specific event type for the transition
		DK,             // Destination Keys
		DIE,            // Destination Inner Event
	]{
		Derive:      cb,
		Destination: linkDestination,
	}
}
func RepoPSMGeneralLogicHook(cb func(context.Context, RepoPSMHookBaton, *RepoState, *RepoEvent) error) psm.GeneralLogicHook[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
] {
	return psm.GeneralLogicHook[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
	](cb)
}
func RepoPSMGeneralStateDataHook(cb func(context.Context, sqrlx.Transaction, *RepoState) error) psm.GeneralStateDataHook[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
] {
	return psm.GeneralStateDataHook[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
	](cb)
}
func RepoPSMGeneralEventDataHook(cb func(context.Context, sqrlx.Transaction, *RepoState, *RepoEvent) error) psm.GeneralEventDataHook[
	*RepoKeys,      // implements psm.IKeyset
	*RepoState,     // implements psm.IState
	RepoStatus,     // implements psm.IStatusEnum
	*RepoStateData, // implements psm.IStateData
	*RepoEvent,     // implements psm.IEvent
	RepoPSMEvent,   // implements psm.IInnerEvent
] {
	return psm.GeneralEventDataHook[
		*RepoKeys,      // implements psm.IKeyset
		*RepoState,     // implements psm.IState
		RepoStatus,     // implements psm.IStatusEnum
		*RepoStateData, // implements psm.IStateData
		*RepoEvent,     // implements psm.IEvent
		RepoPSMEvent,   // implements psm.IInnerEvent
	](cb)
}
