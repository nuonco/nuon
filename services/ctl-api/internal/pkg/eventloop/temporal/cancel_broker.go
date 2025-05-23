package temporal

import (
	// "slices"

	"go.temporal.io/sdk/workflow"
)

const CancelChannelName = "cancel-signal"

type CancelBroker[Message any] struct {
	pctx     workflow.Context
	done     bool
	msgs     []Message
	children []child[Message]
}

type child[Message any] struct {
	ctx      workflow.Context
	match    func(Message) bool
	cancelfn func()
}

func NewCancelBroker[Message any](pctx workflow.Context) *CancelBroker[Message] {
	b := &CancelBroker[Message]{
		pctx: pctx,
	}

	workflow.Go(pctx, func(ctx workflow.Context) {
		sel := workflow.NewSelector(ctx)
		sel.AddReceive(ctx.Done(), func(c workflow.ReceiveChannel, _ bool) {
			b.done = true
		})
		sel.AddReceive(workflow.GetSignalChannel(ctx, CancelChannelName), func(c workflow.ReceiveChannel, _ bool) {
			var msg Message
			c.Receive(ctx, &msg)
			b.newMessage(msg)
		})
		for !b.done {
			sel.Select(ctx)
		}
	})
	return b
}

func (b *CancelBroker[Message]) newMessage(msg Message) {
	var remove []int

	// TODO(sdboyer) commented out until we have test showing this is reliable
	// defer func() {
	// 	// TODO(sdboyer) will this actually clean up children? As in, does temporal auto-cancel contexts in a way this code can observe?
	// 	for i := len(remove); i >= 0; i-- {
	// 		pos := remove[i]
	// 		b.children = slices.Delete(b.children, pos, pos+1)
	// 	}
	// }()

	// new message, check it against all existing children
	for i, child := range b.children {
		if child.ctx.Err() != nil {
			// child context is done, remove from list
			remove = append(remove, i)
			continue
		}

		if child.match(msg) {
			child.cancelfn()
			// b.children = slices.Delete(b.children, i, i+1)
			return
		}
	}

	b.msgs = append(b.msgs, msg)
}

func (b *CancelBroker[Message]) newChild(ch child[Message]) {
	// new child, check it against all pending messages
	for _, msg := range b.msgs {
		if ch.match(msg) {
			ch.cancelfn()
			// b.msgs = slices.Delete(b.msgs, i, i+1)
			return
		}
	}

	b.children = append(b.children, ch)
}

func (b *CancelBroker[Message]) Join(pctx workflow.Context, match func(Message) bool) workflow.Context {
	ctx, cancel := workflow.WithCancel(pctx)
	if b.done {
		cancel()
		return ctx
	}

	b.newChild(child[Message]{
		ctx:      ctx,
		match:    match,
		cancelfn: cancel,
	})

	return ctx
}
