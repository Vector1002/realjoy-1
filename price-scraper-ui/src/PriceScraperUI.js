import React, { useState, useEffect } from "react";
import { format } from "date-fns";
import "bootstrap/dist/css/bootstrap.min.css";

const Button = ({ onClick, loading }) => (
  <button
    className="btn btn-primary mt-3"
    onClick={onClick}
    disabled={loading}
  >
    {loading ? "Scraping..." : "Start Scraping"}
  </button>
);

const Calendar = ({ selected, onSelect }) => (
  <input
    type="date"
    value={format(selected, "yyyy-MM-dd")}
    onChange={(e) => onSelect(new Date(e.target.value))}
    className="form-control"
  />
);

// Make URLs clickable in new tab
const Table = ({ data }) => (
  <table className="table table-bordered mt-4">
    <thead>
      <tr>
        <th>URL</th>
        <th>Price</th>
      </tr>
    </thead>
    <tbody>
      {data.map((item, idx) => (
        <tr key={idx}>
          <td>
            <a href={item.url} target="_blank" rel="noopener noreferrer">
              {item.url}
            </a>
          </td>
          <td>{item.price}</td>
        </tr>
      ))}
    </tbody>
  </table>
);

export default function PriceScraperUI() {
  const [startDate, setStartDate] = useState(new Date());
  const [endDate, setEndDate] = useState(new Date());
  const [data, setData] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleScrape = async () => {
    setLoading(true);
    setError(null);
    setData([]); // Clear table immediately when scraping starts
    try {
      const res = await fetch("http://realjoy-1-3.onrender.com/scrape", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          arrivalDate: format(startDate, "yyyy-MM-dd"),
          departureDate: format(endDate, "yyyy-MM-dd"),
        }),
      });
      if (!res.ok) throw new Error("Network response was not ok");
      const result = await res.json();
      setData(result);
    } catch (err) {
      console.error("Scrape failed:", err);
      setError("Failed to fetch data. Please try again later.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mt-5">
      <h2>🏨 Hotel Price Scraper</h2>
      <div className="row mt-4">
        <div className="col">
          <label>Arrival Date</label>
          <Calendar selected={startDate} onSelect={setStartDate} />
        </div>
        <div className="col">
          <label>Departure Date</label>
          <Calendar selected={endDate} onSelect={setEndDate} />
        </div>
      </div>
      <Button onClick={handleScrape} loading={loading} />
      {loading && (
      <div className="mt-3 text-center">
        <div className="spinner-border text-primary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    )}

      {error && <div className="alert alert-danger mt-3">{error}</div>}
      {!loading && data.length > 0 && <Table data={data}/>}
    </div>
  );
}
