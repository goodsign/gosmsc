package gosmsc

import (
	. "github.com/goodsign/gosmsc/contract"
	"testing"
	"time"
)

func TestFaultySend(t *testing.T) {
	impl, err := newTestSenderCheckerImpl(&smscTestClientOptions{false, true, 0}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = impl.Send("+7 921 123 45 67", "test", false)
	if err == nil {
		t.Fatalf("Expected to get error. Got: nil.")
	}
}

func TestInvalidPasswordSend(t *testing.T) {
	impl, err := newTestSenderCheckerImpl(&smscTestClientOptions{true, false, 0}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = impl.Send("+7 921 123 45 67", "test", false)
	if err == nil {
		t.Fatal(err)
	}
}

func TestSuccessfulSend(t *testing.T) {
	impl, err := newTestSenderCheckerImpl(&smscTestClientOptions{false, false, 0}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = impl.Send("+7 921 123 45 67", "test", false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSuccessfulSendWithTracking(t *testing.T) {
	expectedCode := MessageStatusCode(555)
	impl, err := newTestSenderCheckerImpl(&smscTestClientOptions{false, false, expectedCode}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id, err := impl.Send("+7 921 123 45 67", "test", true)
	if err != nil {
		t.Fatal(err)
	}
	mstatus, err := impl.GetActualStatus(id)
	if err != nil {
		t.Fatal(err)
	}
	if mstatus.StatusCode != MessageStatusCodeUnknown {
		t.Fatalf("Expected code = '%d'. Got '%d'", MessageStatusCodeUnknown, mstatus.StatusCode)
	}
	impl.tracker.tickerForTest <- false
	mstatus, err = impl.GetActualStatus(id)
	if err != nil {
		t.Fatal(err)
	}
	if mstatus.StatusCode != expectedCode {
		t.Fatalf("Expected code = '%d'. Got '%d'", expectedCode, mstatus.StatusCode)
	}
}

func expectPanicOnClosedChannel(c chan bool, t *testing.T, tracker *MessageTracker) {
	defer func() {
		recover()
	}()
	c <- false
	t.Fatal("Expected channel to be closed!")
}

func TestStoppingTracking(t *testing.T) {
	expectedCode := MessageStatusCode(555)
	impl, err := newTestSenderCheckerImpl(&smscTestClientOptions{false, false, expectedCode}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	id, err := impl.Send("+7 921 123 45 67", "test", true)
	if err != nil {
		t.Fatal(err)
	}
	mstatus, err := impl.GetActualStatus(id)
	if err != nil {
		t.Fatal(err)
	}
	if mstatus.StatusCode != MessageStatusCodeUnknown {
		t.Fatalf("Expected code = '%d'. Got '%d'", MessageStatusCodeUnknown, mstatus.StatusCode)
	}
	err = impl.tracker.Stop()
	if err != nil {
		t.Fatal(err)
	}
	if !impl.tracker.IsStopped() {
		t.Fatal("Tracker didn't stop")
	}
	expectPanicOnClosedChannel(impl.tracker.tickerForTest, t, impl.tracker)
	expectPanicOnClosedChannel(impl.tracker.stopChannel, t, impl.tracker)
	mstatus, err = impl.GetActualStatus(id)
	if err != nil {
		t.Fatal(err)
	}
	if mstatus.StatusCode != MessageStatusCodeUnknown {
		t.Fatalf("Expected code = '%d'. Got '%d'", MessageStatusCodeUnknown, mstatus.StatusCode)
	}
}
