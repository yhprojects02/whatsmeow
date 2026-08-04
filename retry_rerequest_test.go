package whatsmeow

import (
	"fmt"
	"sync"
	"testing"

	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func newRerequestTestClient() *Client {
	return &Client{
		Log:                  waLog.Noop,
		phoneRerequestCounts: make(map[types.MessageID]int),
	}
}

func TestClaimPhoneRerequestStopsAtLimit(t *testing.T) {
	cli := newRerequestTestClient()
	const id = types.MessageID("3EB0DEADBEEF")

	for i := 0; i < MaxPhoneRerequests; i++ {
		if !cli.claimPhoneRerequest(id) {
			t.Fatalf("request %d should have been allowed", i+1)
		}
	}
	for i := 0; i < 100; i++ {
		if cli.claimPhoneRerequest(id) {
			t.Fatalf("request %d should have been blocked", MaxPhoneRerequests+i+1)
		}
	}
	if got := cli.phoneRerequestCounts[id]; got != MaxPhoneRerequests {
		t.Fatalf("counter = %d, want %d", got, MaxPhoneRerequests)
	}
}

func TestClaimPhoneRerequestPerMessageBudget(t *testing.T) {
	cli := newRerequestTestClient()
	const a = types.MessageID("AAA")
	const b = types.MessageID("BBB")

	for i := 0; i < MaxPhoneRerequests; i++ {
		cli.claimPhoneRerequest(a)
	}
	if cli.claimPhoneRerequest(a) {
		t.Fatal("exhausted message should be blocked")
	}
	if !cli.claimPhoneRerequest(b) {
		t.Fatal("other message must keep its own budget")
	}
}

func TestClaimPhoneRerequestRejectsEmptyID(t *testing.T) {
	cli := newRerequestTestClient()
	if cli.claimPhoneRerequest("") {
		t.Fatal("empty message ID must not be requested from the phone")
	}
	if len(cli.phoneRerequestCounts) != 0 {
		t.Fatalf("empty ID must not be tracked, map has %d entries", len(cli.phoneRerequestCounts))
	}
}

func TestClaimPhoneRerequestLazyInit(t *testing.T) {
	cli := &Client{Log: waLog.Noop}
	if !cli.claimPhoneRerequest("XYZ") {
		t.Fatal("first request on a zero-value client should be allowed")
	}
}

func TestClaimPhoneRerequestBoundsMemory(t *testing.T) {
	cli := newRerequestTestClient()
	for i := 0; i < maxTrackedPhoneRerequests*2; i++ {
		cli.claimPhoneRerequest(types.MessageID(fmt.Sprintf("ID%d", i)))
	}
	if len(cli.phoneRerequestCounts) > maxTrackedPhoneRerequests {
		t.Fatalf("tracked %d messages, want <= %d", len(cli.phoneRerequestCounts), maxTrackedPhoneRerequests)
	}
}

func TestClaimPhoneRerequestConcurrent(t *testing.T) {
	cli := newRerequestTestClient()
	const id = types.MessageID("CONCURRENT")

	var wg sync.WaitGroup
	allowed := make(chan struct{}, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cli.claimPhoneRerequest(id) {
				allowed <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(allowed)
	if got := len(allowed); got != MaxPhoneRerequests {
		t.Fatalf("%d requests allowed, want %d", got, MaxPhoneRerequests)
	}
}
