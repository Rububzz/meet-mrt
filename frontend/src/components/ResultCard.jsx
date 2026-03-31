import useWindowWidth from "../hooks/useWindowWidth";

export default function ResultCard({ result, rank, stationA, stationB }) {
  const { station, timeFromA, timeFromB, difference } = result;

  function formatTime(seconds) {
    const minutes = Math.round(seconds / 60);
    if (minutes < 60) return `${minutes} min`;
    const hours = Math.floor(minutes / 60);
    const remaining = minutes % 60;
    return `${hours}h ${remaining}m`;
  }

  function fairnessScore() {
    const maxDiff = 30 * 60; // 30 minutes in seconds
    const score = Math.max(0, 100 - Math.round((difference / maxDiff) * 100));
    return score;
  }

  const width = useWindowWidth();

  return (
    <div
      style={{
        border: rank === 0 ? "2px solid #16a34a" : "1px solid #ddd",
        borderRadius: "8px",
        padding: "16px",
        marginBottom: "12px",
        background: "white",
      }}
    >
      {rank === 0 && (
        <span
          style={{
            background: "green",
            color: "white",
            fontSize: "11px",
            padding: "2px 8px",
            borderRadius: "4px",
            marginBottom: "8px",
            display: "inline-block",
          }}
        >
          Best Match
        </span>
      )}

      <h3 style={{ margin: "8px 0", color: "#1a1a1a" }}>
        #{rank + 1} {station.name}
      </h3>

      <p style={{ color: "#888", fontSize: "13px", margin: "0 0 12px" }}>
        {station.code} — {station.lines.join(", ")} line
      </p>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: width > 600 ? "1fr 1fr 1fr" : "1fr 1fr",
          gap: "12px",
          marginBottom: "12px",
        }}
      >
        <div
          style={{
            flex: 1,
            background: "#f5f5f5",
            padding: "10px",
            borderRadius: "6px",
          }}
        >
          <div style={{ fontSize: "11px", color: "#888", marginBottom: "4px" }}>
            From {stationA?.name}
          </div>
          <div
            style={{ fontSize: "20px", fontWeight: "bold", color: "#1a1a1a" }}
          >
            {formatTime(timeFromA)}
          </div>
        </div>

        <div
          style={{
            flex: 1,
            background: "#f5f5f5",
            padding: "10px",
            borderRadius: "6px",
          }}
        >
          <div style={{ fontSize: "11px", color: "#888", marginBottom: "4px" }}>
            From {stationB?.name}
          </div>
          <div
            style={{ fontSize: "20px", fontWeight: "bold", color: "#1a1a1a" }}
          >
            {formatTime(timeFromB)}
          </div>
        </div>

        <div
          style={{
            flex: 1,
            background: "#f5f5f5",
            padding: "10px",
            borderRadius: "6px",
          }}
        >
          <div style={{ fontSize: "11px", color: "#888", marginBottom: "4px" }}>
            Fairness
          </div>
          <div
            style={{
              fontSize: "20px",
              fontWeight: "bold",
              color:
                fairnessScore() >= 80
                  ? "#16a34a"
                  : fairnessScore() >= 50
                    ? "#d97706"
                    : "#dc2626",
            }}
          >
            {fairnessScore()}%
          </div>
        </div>
      </div>

      <p style={{ fontSize: "12px", color: "#aaa", margin: 0 }}>
        Difference: {formatTime(difference)}
      </p>
    </div>
  );
}
