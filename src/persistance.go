package main

// import (
// 	"github.com/asynkron/protoactor-go/persistence"
// 	"google.golang.org/protobuf/reflect/protoreflect"
// )

// type Provider struct {
// 	providerState persistence.ProviderState
// }

// func NewProvider(snapshotInterval int) *Provider {
// 	return &Provider{
// 		providerState: persistence.NewInMemoryProvider(snapshotInterval),
// 	}
// }

// func (p *Provider) InitState(actorName string, eventNum, eventIndexAfterSnapshot int) {
// 	p.providerState.PersistSnapshot(
// 		actorName,
// 		eventIndexAfterSnapshot,
// 		&Snapshot{protoMsg: protoMsg{state: "state" + strconv.Itoa(eventIndexAfterSnapshot-1)}},
// 	)
// }

// func (p *Provider) GetState() persistence.ProviderState {
// 	return p.providerState
// }

// type protoMsg struct {
// 	state string
// 	set   bool
// 	value string
// }

// func (p *protoMsg) Reset()         {}
// func (p *protoMsg) String() string { return p.state }
// func (p *protoMsg) ProtoMessage()  {}

// type (
// 	Message  struct{ protoMsg }
// 	Snapshot struct{ protoMsg }
// )

// func (m *protoMsg) ProtoReflect() protoreflect.Message { return (*message)(m) }

// type message protoMsg
