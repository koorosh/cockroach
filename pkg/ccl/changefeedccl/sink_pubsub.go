package changefeedccl

import (
	"context"
	"encoding/json"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"golang.org/x/oauth2/google"
	"hash/crc32"
	"net/url"
	"path"
	"regexp"
	"strings"

	_ "cloud.google.com/go/pubsub"
	gcpClient "cloud.google.com/go/pubsub/apiv1"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/errors"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/gcppubsub"
	_ "gocloud.dev/pubsub/mempubsub"
	pbapi "google.golang.org/genproto/googleapis/pubsub/v1"
	"google.golang.org/grpc"
)

const googleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
const credentialsParam = "CREDENTIALS"
const authParam = "AUTH"
const authSpecified = "specified"
const authImplicit = "implicit"
const authDefault = "default"
const gcpScheme = "gcppubsub"
const memScheme = "mem"
const gcpScope = "https://www.googleapis.com/auth/pubsub"
const gcpTopicFullPAth = "projects/%s/topics/%s"

var (
	fullTopicPathRE  = regexp.MustCompile("^projects/[^/]+/topics/[^/]+$")
	shortTopicPathRE = regexp.MustCompile("^[^/]+/[^/]+$")
)

// isPubsubSInk returns true if url contains scheme with valid pubsub sink
func isPubsubSink(u *url.URL) bool {
	switch u.Scheme {
	case gcpScheme, memScheme:
		return true
	default:
		return false
	}
}

// payload struct is sent to the sink
type payload struct {
	Key   json.RawMessage `json:"key"`
	Value json.RawMessage `json:"value"`
	Topic string `json:"topic"`
}

// pubsubMessage is sent to worker channels for workers to consume
type pubsubMessage struct {
	alloc   kvevent.Alloc
	message payload
	isFlush bool
	topicId descpb.ID
}

type gcpPubsubSink struct {
	client     *gcpClient.PublisherClient
	cleanup    func()
	pubsubSink *pubsubSink
	creds      *google.Credentials
	conn       *grpc.ClientConn
}

type memPubsubSink struct {
	pubsubSink *pubsubSink
}

type topicStruct struct {
	topicName string
	pathName string
	topicClient *pubsub.Topic
}

type pubsubSink struct {
	topics			map[descpb.ID]*topicStruct
	url        sinkURL
	numWorkers int

	workerCtx   context.Context
	workerGroup ctxgroup.Group

	exitWorkers func()               // Signaled to shut down all workers.
	eventsChans []chan pubsubMessage //channel where messages are consumed and sent out

	// flushDone channel signaled when flushing completes.
	flushDone chan struct{}

	// errChan is written to indicate an error while sending message.
	errChan chan error
}

// getGCPCredentials returns gcp credentials parsed out from url
func getGCPCredentials(u sinkURL, ctx context.Context) (*google.Credentials, error) {
	var credsJSON []byte
	var creds *google.Credentials
	authOption := u.consumeParam(authParam)

	// implemented according to https://github.com/cockroachdb/cockroach/pull/64737
	switch authOption {
	case authImplicit:
		creds, err := google.FindDefaultCredentials(ctx, gcpScope)
		if err != nil {
			return nil, err
		}
		return creds, nil
	case authSpecified:
		fallthrough
	case authDefault:
		fallthrough
	default:
		err := u.decodeBase64(credentialsParam, &credsJSON)
		if err != nil {
			return nil, errors.Wrap(err, "decoding credentials json")
		}
		creds, err = google.CredentialsFromJSON(ctx, credsJSON, gcpScope)
		if err != nil {
			return nil, errors.Wrap(err, "creating credentials")
		}
		return creds, nil
	}
}

// parseGCPURL returns fullpath of url if properly formatted
func parseGCPURL(u sinkURL) (string, error) {
	fullPath := path.Join(u.Host, u.Path)
	if fullTopicPathRE.MatchString(fullPath) {
		parts := strings.SplitN(fullPath, "/", 4)
		if len(parts) < 4 {
			return "", errors.Errorf("unexpected number of components in %s", fullPath)
		}
		return fullPath, nil
	} else if shortTopicPathRE.MatchString(fullPath) {
		return path.Join("projects", u.Host, "topics", u.Path), nil
	}
	return "", errors.Errorf("could not parse project and topic from %s", fullPath)
}

func createGCPURL(u sinkURL, topicName string) (string, error){
	// TODO: look into topic name validation https://cloud.google.com/pubsub/docs/admin#resource_names
	return path.Join("projects", u.Host, "topics", topicName), nil
}

// MakePubsubSink returns the corresponding pubsub sink based on the url given
func MakePubsubSink(ctx context.Context, u *url.URL, opts map[string]string, targets jobspb.ChangefeedTargets,
	settings *cluster.Settings) (Sink, error) {
	log.Info(ctx, "\x1b[33m starting pubsub \x1b[0m")

	sinkURL  := sinkURL{u, u.Query()}
	pubsubTopicName := sinkURL.consumeParam(changefeedbase.SinkParamTopicName)

	switch changefeedbase.FormatType(opts[changefeedbase.OptFormat]) {
	case changefeedbase.OptFormatJSON:
	default:
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptFormat, opts[changefeedbase.OptFormat])
	}

	switch changefeedbase.EnvelopeType(opts[changefeedbase.OptEnvelope]) {
	case changefeedbase.OptEnvelopeWrapped:
	default:
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptEnvelope, opts[changefeedbase.OptEnvelope])
	}

	//if _, ok := opts[changefeedbase.OptKeyInValue]; !ok {
	//	return nil, errors.Errorf(`this sink requires the WITH %s option`, changefeedbase.OptKeyInValue)
	//}

	ctx, cancel := context.WithCancel(ctx)
	// currently just hardcoding numWorkers to 100, it will be a config option later down the road
	p := &pubsubSink {
		workerCtx: ctx, url: sinkURL, numWorkers: 100,
		exitWorkers: cancel, topics: make(map[descpb.ID]*topicStruct),
	}

	//creates a topic for each target
	for id, topic := range targets {
		var topicName string
		if pubsubTopicName == ""{
			topicName = topic.StatementTimeName
		} else {
			topicName = pubsubTopicName
		}
		p.topics[id] = &topicStruct{topicName: topicName}
	}
	p.setupWorkers()

	// creates custom pubsub object based on scheme
	switch u.Scheme {
	case gcpScheme:
		creds, err := getGCPCredentials(p.url, ctx)
		if err != nil {
			_ = p.close()
			return nil, err
		}
		g := &gcpPubsubSink{creds: creds, pubsubSink: p}
		return g, nil
	case memScheme:
		m := &memPubsubSink{pubsubSink: p}
		return m, nil
	default:
		_ = p.close()
		return nil, errors.Errorf("unknown scheme: %s", u.Scheme)
	}
}

// getWorkerCtx returns workerCtx
func (p *pubsubSink) getWorkerCtx() context.Context {
	return p.workerCtx
}

// getUrl returns url
func (p *pubsubSink) getUrl() sinkURL {
	return p.url
}

// EmitRow pushes a message to event channel where it is consumed by workers
func (p *pubsubSink) emitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	_ hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	m := pubsubMessage{alloc: alloc, isFlush: false, topicId: topic.GetID(), message: payload{
		Key:   key,
		Value: value,
		Topic: p.topics[topic.GetID()].topicName,
	}}
	log.Info(ctx, "\x1b[33m sending message \x1b[0m")
	// calculate index by hashing key
	i := p.workerIndex(key)
	select {
	// check the sink context in case workers have been terminated
	case <-p.workerCtx.Done():
		// check again for error in case it triggered since last check
		// will return more verbose error instead of "context canceled"
		return errors.CombineErrors(p.workerCtx.Err(), p.sinkError())
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.errChan:
		// check if there are any errors with sink
		return err
	case p.eventsChans[i] <- m:
		// send message to event channel
	}
	return nil
}

// EmitResolvedTimestamp sends resolved timestamp message
func (p *pubsubSink) emitResolvedTimestamp(ctx context.Context, encoder Encoder, resolved hlc.Timestamp) error {
	payload, err := encoder.EncodeResolvedTimestamp(ctx, "", resolved)
	if err != nil {
		return errors.Wrap(err, "encoding resolved timestamp")
	}

	for topicId, _ := range p.topics {
		err = p.sendMessage(payload, topicId, "")
		if err != nil {
			return errors.Wrap(err, "emiting resolved timestamp")
		}
	}
	return nil
}

// Flush blocks until all messages in the event channels are sent
func (p *pubsubSink) flush(ctx context.Context) error {
	log.Info(p.workerCtx, "\x1b[33m flush mes \x1b[0m")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.errChan:
		return err
	default:
		err := p.flushWorkers()
		if err != nil {
			return err
		}
	}
	log.Info(p.workerCtx, "\x1b[33m flush mes 2 \x1b[0m")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.errChan:
		return err
	case <-p.flushDone:
		return p.sinkError()
	}

}

// Close closes all the channels and shutdowns the topic
func (p *pubsubSink) close() error {
	p.exitWorkers()
	_ = p.workerGroup.Wait()
	for _, topic := range p.topics {
		if topic.topicClient != nil {
			err := topic.topicClient.Shutdown(p.getWorkerCtx())
			if err != nil {
				return errors.Wrap(err, "closing pubsub topic")
			}
		}
	}
	close(p.errChan)
	close(p.flushDone)
	for i := 0; i < p.numWorkers; i++ {
		close(p.eventsChans[i])
	}
	return nil
}

// sendMessage sends a message to the topic
func (p *pubsubSink) sendMessage(m []byte, topicId descpb.ID, key string) error {
	var c *gcpClient.PublisherClient
	var err error
	if p.topics[topicId].topicClient.As(&c){
		gcpMessage := &pbapi.PublishRequest{Topic: p.topics[topicId].pathName}
		gcpMessage.Messages = append(gcpMessage.Messages, &pbapi.PubsubMessage{Data: m, OrderingKey: key})
		_, err = c.Publish(context.Background(), gcpMessage)
	} else {
		// this is for mempubsub since As() on mempubsub returns an unexported topic type that we cannot interact with
		// https://github.com/google/go-cloud/blob/master/pubsub/mempubsub/mem.go#L185-L188
		err = p.topics[topicId].topicClient.Send(p.workerCtx, &pubsub.Message{
			Body: m,
		})
	}
	if err != nil {
		return err
	}
	return nil
}

// setupWorkers sets up the channels used by the sink and starts a goroutine for every worker
func (p *pubsubSink) setupWorkers() {
	// setup events channels to send to workers and the worker group
	p.eventsChans = make([]chan pubsubMessage, p.numWorkers)
	p.workerGroup = ctxgroup.WithContext(p.workerCtx)

	// an error channel with buffer for the first error.
	p.errChan = make(chan error, 1)

	// flushDone notified when flush completes.
	p.flushDone = make(chan struct{}, 1)

	for i := 0; i < p.numWorkers; i++ {
		//initialize worker goroutine and channel for worker
		p.eventsChans[i] = make(chan pubsubMessage)
		j := i
		p.workerGroup.GoCtx(func(ctx context.Context) error {
			p.workerLoop(j)
			return nil
		})
	}
}

// workerLoop consumes any message sent to the channel corresponding to the worker index
func (p *pubsubSink) workerLoop(workerIndex int) {
	for {
		log.Info(p.workerCtx, "\x1b[33m workerselect \x1b[0m")
		select {
		case <-p.workerCtx.Done():
			log.Info(p.workerCtx, "\x1b[33m workerselect DONE\x1b[0m")
			return
		case msg := <-p.eventsChans[workerIndex]:
			if msg.isFlush {
				// Signals a flush request, makes sure that the messages in eventsChans are finished sending
				continue
			}

			m := msg.message
			b, err := json.Marshal(m)
			if err != nil {
				p.exitWorkersWithError(err)
			}
			err = p.sendMessage(b, msg.topicId, string(msg.message.Key), )
			if err != nil {
				p.exitWorkersWithError(err)
			}
			msg.alloc.Release(p.workerCtx)
		}
	}
}

// exitWorkersWithError sends an error to the sink error channel
func (p *pubsubSink) exitWorkersWithError(err error) {
	// errChan has buffer size 1, first error will be saved to the buffer and
	// subsequent errors will be ignored
	select {
	case p.errChan <- err:
		p.exitWorkers()
	default:
	}
	log.Info(p.workerCtx, "\x1b[33m exiting workers \x1b[0m")
}

// sinkError checks if there is an error in the error channel
func (p *pubsubSink) sinkError() error {
	log.Info(p.workerCtx, "\x1b[33m select sinkError \x1b[0m")
	select {
	case err := <-p.errChan:
		return err
	default:
	}
	return nil
}

// workerIndex hashes key to return a worker index
func (p *pubsubSink) workerIndex(key []byte) uint32 {
	return crc32.ChecksumIEEE(key) % uint32(p.numWorkers)
}

// flushWorkers sends a flush message to every worker channel and then signals sink that flush is done
func (p *pubsubSink) flushWorkers() error {
	for i := 0; i < p.numWorkers; i++ {
		//flush message will be blocked until all the messages in the channel are processed
		log.Info(p.workerCtx, "\x1b[33m flush \x1b[0m")
		select {
		case <-p.workerCtx.Done():
			return p.workerCtx.Err()
		case p.eventsChans[i] <- pubsubMessage{isFlush: true}:
		}
	}
	log.Info(p.workerCtx, "\x1b[33m flush2 \x1b[0m")
	select {
	// signals sink that flush is complete
	case <-p.workerCtx.Done():
		return p.workerCtx.Err()
	case p.flushDone <- struct{}{}:
		return nil
	}
}

// Dial connects to gcp client and opens a topic
func (p *gcpPubsubSink) Dial() error {
	// Open a gRPC connection to the GCP Pub/Sub API.
	conn, cleanup, err := gcppubsub.Dial(p.pubsubSink.getWorkerCtx(), p.creds.TokenSource)
	if err != nil {
		return errors.Wrap(err, "establishing gcp connection")
	}
	p.conn = conn

	p.cleanup = cleanup

	// Construct a PublisherClient using the connection.
	pubClient, err := gcppubsub.PublisherClient(p.pubsubSink.getWorkerCtx(), conn)
	if err != nil {
		return errors.Wrap(err, "creating publisher client")
	}
	p.client = pubClient

	// Construct a *pubsub.Topic. if user  spcifies specific topic
	//fullPath, err := parseGCPURL(p.pubsubSink.getUrl())
	//if err != nil {
	//	return errors.Wrap(err, "parsing url")
	//}

	for _, topic := range p.pubsubSink.topics {
		topicPath, err := createGCPURL(p.pubsubSink.url, topic.topicName)
		if err != nil {
			return errors.Wrap(err, "invalid topic name")
		}
		// TODO: implement topic config https://pkg.go.dev/cloud.google.com/go/pubsub#TopicConfig
		topic.topicClient, err = gcppubsub.OpenTopicByPath(pubClient, topicPath, nil)
		if err != nil {
			return errors.Wrap(err, "opening topic")
		}
		topic.pathName = topicPath
	}
	//
	//
	//topic, err := gcppubsub.OpenTopicByPath(pubClient, fullPath, nil)
	//if err != nil {
	//	return errors.Wrap(err, "opening topic")
	//}
	//
	//p.pubsubSink.setTopic(topic)
	return nil
}

// EmitRow calls the pubsubDesc EmitRow
func (p *gcpPubsubSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	err := p.pubsubSink.emitRow(ctx, topic, key, value, updated, alloc)
	if err != nil {
		return err
	}
	return nil
}

// EmitResolvedTimestamp calls the pubsubDesc EmitResolvedTimestamp
func (p *gcpPubsubSink) EmitResolvedTimestamp(ctx context.Context, encoder Encoder, resolved hlc.Timestamp) error {
	return p.pubsubSink.emitResolvedTimestamp(ctx, encoder, resolved)
}

// Flush calls the pubsubDesc Flush
func (p *gcpPubsubSink) Flush(ctx context.Context) error {
	return p.pubsubSink.flush(ctx)
}

// Close calls the pubsubDesc Close and closes the client and connection
func (p *gcpPubsubSink) Close() error {
	log.Info(p.pubsubSink.workerCtx, "\x1b[33m closing pubsub \x1b[0m")
	if p.pubsubSink != nil {
		err := p.pubsubSink.close()
		if err != nil {
			return err
		}
	}
	if p.client != nil {
		p.client.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	if p.cleanup != nil {
		p.cleanup()
	}
	return nil
}

// Dial opens topic using url
func (p *memPubsubSink) Dial() error {
	//topic, err := pubsub.OpenTopic(p.pubsubSink.getWorkerCtx(), p.pubsubSink.url.String())
	//if err != nil {
	//	return err
	//}
	//
	//p.pubsubSink.setTopic(topic)
	var err error
	for _, topic := range p.pubsubSink.topics {
		//TODO: implement topic config https://pkg.go.dev/cloud.google.com/go/pubsub#TopicConfig
		// implement odering key option for gcp https://cloud.google.com/pubsub/docs/publisher#using_ordering_keys
		topic.topicClient, err = pubsub.OpenTopic(p.pubsubSink.getWorkerCtx(), p.pubsubSink.url.String())
		if err != nil {
			return errors.Wrap(err, "opening topic")
		}
	}
	return nil
}

// EmitRow calls the pubsubDesc EmitRow
func (p *memPubsubSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	err := p.pubsubSink.emitRow(ctx, topic, key, value, updated, alloc)
	if err != nil {
		return err
	}
	return nil
}

// EmitResolvedTimestamp calls the pubsubDesc EmitResolvedTimestamp
func (p *memPubsubSink) EmitResolvedTimestamp(ctx context.Context, encoder Encoder, resolved hlc.Timestamp) error {
	return p.pubsubSink.emitResolvedTimestamp(ctx, encoder, resolved)
}

// Flush calls the pubsubDesc Flush
func (p *memPubsubSink) Flush(ctx context.Context) error {
	return p.pubsubSink.flush(ctx)
}

// Close calls the pubsubDesc Close
func (p *memPubsubSink) Close() error {
	p.pubsubSink.exitWorkers()
	_ = p.pubsubSink.workerGroup.Wait()
	close(p.pubsubSink.errChan)
	close(p.pubsubSink.flushDone)
	for i := 0; i < p.pubsubSink.numWorkers; i++ {
		close(p.pubsubSink.eventsChans[i])
	}
	return nil
}
