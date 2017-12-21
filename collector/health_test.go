package collector

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClusterHealth(t *testing.T) {
	s := `{"cluster_name":"es","status":"green","timed_out":false,"number_of_nodes":65,"number_of_data_nodes":62,"active_primary_shards":11331,"active_shards":21882,"relocating_shards":2,"initializing_shards":0,"unassigned_shards":0,"delayed_unassigned_shards":0,"number_of_pending_tasks":0,"number_of_in_flight_fetch":0,"task_max_waiting_in_queue_millis":0,"active_shards_percent_as_number":100.0}`

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(s))
	}))

	defer testServer.Close()
	u, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}
	c := NewClusterHealth(http.DefaultClient, u)
	chr, err := c.getMetrics()
	if err != nil {
		t.Fatalf("Failed to parse cluster info %s", err)
	}
	So(chr.ClusterName, ShouldEqual, "es")
	So(chr.Status, ShouldEqual, "green")
	So(chr.TimedOut, ShouldBeTrue)
	So(chr.NumberOfNodes, ShouldBeGreaterThan, 0)
}
