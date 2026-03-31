import { useEffect, useState } from "react";
import StationPicker from "./components/StationPicker";
import ResultCard from "./components/ResultCard";
import useWindowWidth from "./hooks/useWindowWidth";

const API_URL = "http://localhost:8080";

export default function App() {
  const [stationA, setStationA] = useState(null);
  const [stationB, setStationB] = useState(null);
  const [results, setResults] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [remaining, setRemaining] = useState(null);

  useEffect(() => {
    fetch(`${API_URL}/api/remaining`)
      .then((res) => res.json())
      .then((data) => setRemaining(data.remaining));
  }, []);

  const width = useWindowWidth();

  const canSearch =
    stationA && stationB && stationA.name !== stationB.name && remaining > 0;

  async function handleFind() {
    setLoading(true);
    setError(null);
    setResults(null);

    try {
      const res = await fetch(`${API_URL}/api/find-meetup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          stationA: stationA.name,
          stationB: stationB.name,
        }),
      });

      const data = await res.json();

      if (data.error) {
        setError(data.error);
      } else {
        setResults(data);
      }
    } catch (e) {
      setError("Could not reach the server. Is the backend running?");
    } finally {
      setLoading(false);
      fetch(`${API_URL}/api/remaining`)
        .then((res) => res.json())
        .then((data) => setRemaining(data.remaining));
    }
  }

  return (
    <div
      style={{
        maxWidth: "600px",
        margin: "0 auto",
        padding: width > 600 ? "40px 20px" : "20px 16px",
      }}
    >
      <h1 style={{ marginBottom: "8px" }}>MeetMRT</h1>
      <p style={{ color: "#888", marginBottom: "32px" }}>
        Find the fairest MRT meeting point for two people.
      </p>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: width > 600 ? "1fr 1fr" : "1fr",
          gap: "20px",
          marginBottom: "24px",
        }}
      >
        <StationPicker
          label="Your Station"
          value={stationA}
          onChange={setStationA}
        />
        <StationPicker
          label="Their Station"
          value={stationB}
          onChange={setStationB}
        />
      </div>
      {remaining !== null && (
        <p
          style={{
            fontSize: "12px",
            color: remaining > 10 ? "#888" : remaining > 0 ? "orange" : "red",
            marginBottom: "8px",
            textAlign: "right",
          }}
        >
          {remaining > 0
            ? `${remaining} searches remaining this month`
            : "Monthly search limit reached"}
        </p>
      )}

      <button
        onClick={handleFind}
        disabled={!canSearch || loading}
        style={{
          width: "100%",
          padding: "14px",
          fontSize: "16px",
          background: canSearch ? "#1a1a1a" : "#ccc",
          color: "white",
          border: "none",
          borderRadius: "8px",
          cursor: canSearch ? "pointer" : "not-allowed",
          marginBottom: "32px",
        }}
      >
        {loading
          ? "Finding..."
          : remaining === 0
            ? "Limit Reached"
            : "Find Meeting Point"}
      </button>

      {error && (
        <div
          style={{
            background: "#fff0f0",
            border: "1px solid #ffcccc",
            borderRadius: "8px",
            padding: "16px",
            color: "red",
            marginBottom: "24px",
          }}
        >
          {error}
        </div>
      )}

      {results && (
        <div>
          <h2 style={{ marginBottom: "16px" }}>
            Top meeting points between{" "}
            <span style={{ color: "green" }}>{results.stationA.name}</span> and{" "}
            <span style={{ color: "blue" }}>{results.stationB.name}</span>
          </h2>

          {results.candidates.map((result, index) => (
            <ResultCard
              key={result.station.code}
              result={result}
              rank={index}
              stationA={results.stationA}
              stationB={results.stationB}
            />
          ))}
        </div>
      )}
    </div>
  );
}
