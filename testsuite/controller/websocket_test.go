package controller

// func TestWebsocketBroadcast(t *testing.T) {
// 	server := testsuite.InitServer(t)

// 	ts := httptest.NewServer(server.Router())
// 	defer ts.Close() // nolint:errcheck

// 	u, _ := url.Parse(ts.URL)
// 	u.Scheme = "ws"
// 	u.Path = "/api/v1/ws"

// 	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	defer resp.Body.Close() // nolint:errcheck
// 	require.NoError(t, err, "failed to connect to websocket")
// 	require.Equal(t, 101, resp.StatusCode, "expected websocket upgrade success")

// 	svc := server.Service(service.Websocket).(*websocketSvc.Service)
// 	svc.Broadcast(websocketSvc.TaskCreated, "moo")

// 	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
// 	_, received, err := conn.ReadMessage()
// 	require.NoError(t, err, "failed to read message")
// 	require.Equal(t, `{"payload":"moo","subject":"task:created"}`, strings.Trim(string(received), "\n"))
// }
