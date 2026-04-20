import React, { useState, useEffect, useRef } from "react";

const BACKEND_URL = "https://endterm3project.onrender.com";
const WS_URL = "wss://endterm3project.onrender.com/ws";

export default function App() {
  const [data, setData] = useState({ servers: [], logs: [] });
  const [connected, setConnected] = useState(false);
  const logsEndRef = useRef(null);

  useEffect(() => {
    const connect = () => {
      const ws = new WebSocket(WS_URL);

      ws.onopen = () => setConnected(true);

      ws.onmessage = (event) => {
        try {
          setData(JSON.parse(event.data));
        } catch (err) {
          console.error("Invalid WS data", err);
        }
      };

      ws.onclose = () => {
        setConnected(false);
        setTimeout(connect, 2000);
      };

      ws.onerror = () => {
        ws.close();
      };
    };

    connect();
  }, []);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [data.logs]);

  const triggerChaos = async () => {
    try {
      await fetch(`${BACKEND_URL}/chaos`);
    } catch (err) {
      console.error("Chaos trigger failed", err);
    }
  };

  const stats = {
    total: data.servers.length,
    online: data.servers.filter((s) => s.status !== "DEAD").length,
    critical: data.servers.filter((s) => s.status === "CRITICAL").length,
  };

  return (
    <div style={{ padding: 20, background: "#0a0f1e", color: "#e2e8f0", minHeight: "100vh", fontFamily: "sans-serif" }}>
      
      {/* HEADER */}
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 20 }}>
        <h2>🚀 METRIC_VOYAGER</h2>

        <div style={{ display: "flex", gap: 10 }}>
          <button onClick={triggerChaos} style={{ padding: 10, background: "#ef4444", color: "white", border: "none", borderRadius: 6 }}>
            INJECT CHAOS
          </button>

          <div style={{ padding: 10, border: "1px solid gray", borderRadius: 6 }}>
            {connected ? "🟢 LIVE" : "🔴 OFFLINE"}
          </div>
        </div>
      </div>

      {/* STATS */}
      <div style={{ display: "flex", gap: 20, marginBottom: 20 }}>
        <div>Total: {stats.total}</div>
        <div style={{ color: "lightgreen" }}>Online: {stats.online}</div>
        <div style={{ color: "red" }}>Critical: {stats.critical}</div>
      </div>

      {/* GRID */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))", gap: 10 }}>
        {data.servers.map((srv) => (
          <div key={srv.id} style={{ padding: 15, background: "#1e293b", borderRadius: 10 }}>
            <div style={{ display: "flex", justifyContent: "space-between" }}>
              <b>{srv.id}</b>
              <span>{srv.status}</span>
            </div>

            <div>CPU: {Math.round(srv.cpu)}%</div>
            <div>MEM: {Math.round(srv.memory)}%</div>
            <div>LAT: {Math.round(srv.latency)}</div>
            <div>ERR: {srv.errorRate?.toFixed(1)}%</div>

            {srv.status === "DEAD" && (
              <div style={{ color: "red", marginTop: 10 }}>☠ NODE DEAD</div>
            )}
          </div>
        ))}
      </div>

      {/* LOGS */}
      <div style={{ marginTop: 30 }}>
        <h3>Logs</h3>
        <div style={{ maxHeight: 200, overflowY: "auto", background: "#111827", padding: 10 }}>
          {data.logs.map((log, i) => (
            <div key={i}>{log}</div>
          ))}
          <div ref={logsEndRef} />
        </div>
      </div>
    </div>
  );
}