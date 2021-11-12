package cdctest

import (
	"context"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"gocloud.dev/pubsub"
)

// MockPubsubSink is the Webhook sink used in tests.
type MockPubsubSink struct {
	sub *pubsub.Subscription
	ctx context.Context
	groupCtx ctxgroup.Group
	errChan chan error
	url string
	shutdown func()
	mu                 struct {
		syncutil.Mutex
		rows             []string
	}
}

// MakeMockPubsubSink returns a MockPubsubSink object initialized with the given url and context
func MakeMockPubsubSink(url string, ctx context.Context) (*MockPubsubSink, error){
	ctx, shutdown := context.WithCancel(ctx)
	groupCtx := ctxgroup.WithContext(ctx)
	p := &MockPubsubSink{
		ctx: ctx, errChan: make(chan error, 1), url: url, shutdown: shutdown, groupCtx: groupCtx,
	}
	return p, nil
}

// Close shuts down the subscriber object and closes the channels used
func (p *MockPubsubSink) Close() {
	if p.sub != nil {
		_ = p.sub.Shutdown(p.ctx)
	}
	p.shutdown()
	_ = p.groupCtx.Wait()
	close(p.errChan)
}

// Dial opens a subscriber using the url of the MockPubsubSink
func (p *MockPubsubSink) Dial() error{
	var err error
	p.sub, err = pubsub.OpenSubscription(p.ctx, p.url)
	if err != nil {
		return err
	}
	p.groupCtx.GoCtx(func(ctx context.Context) error {
		p.receive()
		return nil
	})
	return nil
}

// receive loops to read in messages
func (p *MockPubsubSink) receive() {
	for {
		msg, err := p.sub.Receive(p.ctx)
		if err != nil {
			select {
			case <-p.ctx.Done():
			case p.errChan <-err:
			default:
			}
			return
		}
		msg.Ack()
		msgBody := string(msg.Body)

		select {
		case <-p.ctx.Done():
			return
		default:
			p.push(msgBody)
		}
	}
}

// push adds a pubsub message to the end of the rows string slice
func (p *MockPubsubSink) push(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mu.rows = append(p.mu.rows, msg)
}

// Pop removes and returns the first string in the rows string slice
func (p *MockPubsubSink) Pop() *string{
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.mu.rows) > 0 {
		oldest := p.mu.rows[0]
		p.mu.rows = p.mu.rows[1:]
		return &oldest
	}
	return nil
}

// CheckSinkError checks the errChan for any errors and returns it
func (p *MockPubsubSink)CheckSinkError() error{
	select {
		case err := <-p.errChan:
			return err
		default:
	}
	return nil
}
