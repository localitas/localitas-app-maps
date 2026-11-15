package maps

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	client "github.com/localitas/localitas-go"
)

func strictDecode(t *testing.T, data []byte, into any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(into); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}
}

func assertCronJob(t *testing.T, got, want client.CronJob) {
	t.Helper()
	if got.ID != want.ID || got.Path != want.Path || got.Method != want.Method || got.Schedule != want.Schedule ||
		got.Description != want.Description || got.Timeout != want.Timeout || got.NotifyLevel != want.NotifyLevel ||
		got.ExecutionStrategy != want.ExecutionStrategy || len(got.Headers) != len(want.Headers) || len(got.Body) != len(want.Body) {
		t.Fatalf("job %+v, want %+v", got, want)
	}
	if (got.Retry == nil) != (want.Retry == nil) || (got.Retry != nil && *got.Retry != *want.Retry) {
		t.Fatalf("job %s retry %+v, want %+v", want.ID, got.Retry, want.Retry)
	}
	if (got.Auth == nil) != (want.Auth == nil) || (got.Auth != nil && *got.Auth != *want.Auth) {
		t.Fatalf("job %s auth %+v, want %+v", want.ID, got.Auth, want.Auth)
	}
}

func TestHandleCron_TypedJobs(t *testing.T) {
	rec := httptest.NewRecorder()
	HandleCron(rec, httptest.NewRequest("GET", "/cron.json", nil))
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("cron.json: %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	var spec client.CronSpec
	strictDecode(t, rec.Body.Bytes(), &spec)
	if len(spec.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(spec.Jobs))
	}
	assertCronJob(t, spec.Jobs[0], client.CronJob{
		ID:          "cron:maps:poi-import",
		Path:        "/api/poi/import",
		Method:      "POST",
		Schedule:    "0 3 * * 0",
		Description: "Imports points of interest from OpenStreetMap Overpass API",
		Timeout:     "60s",
		Retry:       &client.CronRetry{MaxAttempts: 1},
	})
}
