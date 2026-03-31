import { useState, useEffect, useRef } from "react";
import stations from "../data/stations.json";

export default function StationPicker({ label, value, onChange }) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState([]);
  const [open, setOpen] = useState(false);
  const ref = useRef(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClick(e) {
      if (ref.current && !ref.current.contains(e.target)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  // Filter stations as user types
  function handleChange(e) {
    const q = e.target.value;
    setQuery(q);
    onChange(null); // clear selected station when typing

    if (q.length === 0) {
      setResults([]);
      setOpen(false);
      return;
    }

    const filtered = stations.filter(
      (s) =>
        s.name.toLowerCase().includes(q.toLowerCase()) ||
        s.code.toLowerCase().includes(q.toLowerCase()),
    );
    setResults(filtered.slice(0, 6));
    setOpen(true);
  }

  // When user clicks a station in the dropdown
  function handleSelect(station) {
    setQuery(station.name);
    onChange(station);
    setOpen(false);
    setResults([]);
  }

  return (
    <div ref={ref} style={{ position: "relative" }}>
      <label
        style={{ display: "block", marginBottom: "6px", fontWeight: "bold" }}
      >
        {label}
      </label>

      <input
        value={query}
        onChange={handleChange}
        onFocus={() => results.length > 0 && setOpen(true)}
        placeholder="Search station..."
        style={{ width: "100%", padding: "10px", fontSize: "16px" }}
      />

      {open && results.length > 0 && (
        <ul
          style={{
            position: "absolute",
            top: "100%",
            left: 0,
            right: 0,
            background: "white",
            border: "1px solid #ccc",
            borderRadius: "4px",
            listStyle: "none",
            margin: 0,
            padding: 0,
            zIndex: 100,
          }}
        >
          {results.map((station) => (
            <li
              key={station.code}
              onClick={() => handleSelect(station)}
              style={{
                padding: "10px",
                cursor: "pointer",
                borderBottom: "1px solid #eee",
                background: "white",
              }}
              onMouseOver={(e) =>
                (e.currentTarget.style.background = "#f5f5f5")
              }
              onMouseOut={(e) => (e.currentTarget.style.background = "white")}
            >
              <span style={{ fontWeight: "bold", color: "#1a1a1a" }}>
                {station.name}
              </span>
              <span
                style={{ marginLeft: "8px", color: "#888", fontSize: "13px" }}
              >
                {station.code}
              </span>
            </li>
          ))}
        </ul>
      )}

      {value && (
        <p style={{ marginTop: "6px", color: "green", fontSize: "13px" }}>
          ✓ {value.name} selected
        </p>
      )}
    </div>
  );
}
