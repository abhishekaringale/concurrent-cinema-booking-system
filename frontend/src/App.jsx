import React, { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [movies, setMovies] = useState([])
  const [selectedMovie, setSelectedMovie] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // 1. Generate a persistent 12-character hex User ID (just like "7af73b13543c")
  const [userID] = useState(() => {
    const saved = localStorage.getItem('cinema_userID')
    if (saved) return saved
    const newID = Math.random().toString(16).substring(2, 14)
    localStorage.setItem('cinema_userID', newID)
    return newID
  })

  // 2. Fetch movies from the Go backend on mount
  useEffect(() => {
    fetch('/movies')
      .then((res) => {
        if (!res.ok) throw new Error('Failed to load screenings')
        return res.json()
      })
      .then((data) => {
        setMovies(data)
        setLoading(false)
      })
      .catch((err) => {
        setError(err.message)
        setLoading(false)
      })
  }, [])

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-mono p-8 select-none">
      {/* Header */}
      <header className="max-w-4xl w-full mx-auto flex items-center justify-between mb-12 border-b border-slate-900 pb-6">
        <h1 className="text-3xl font-bold tracking-tight text-slate-100">
          Cinema Booking
        </h1>
        <div className="px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-800 text-xs text-slate-400">
          user: <span className="text-slate-200 font-bold">{userID}</span>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="max-w-4xl w-full mx-auto flex-1 flex flex-col items-center">
        {loading ? (
          <div className="text-slate-500 py-12">Loading screenings...</div>
        ) : error ? (
          <div className="text-red-400 py-12">Error: {error}</div>
        ) : (
          /* Horizontal Movie Selector */
          <div className="flex flex-wrap gap-6 justify-center mb-12">
            {movies.map((movie) => {
              const isSelected = selectedMovie?.id === movie.id
              return (
                <div
                  key={movie.id}
                  onClick={() => setSelectedMovie(movie)}
                  className={`cursor-pointer px-6 py-5 rounded-xl border-2 w-64 transition-all ${
                    isSelected
                      ? 'border-sky-500 bg-sky-950/10 shadow-lg shadow-sky-500/10'
                      : 'border-slate-900 bg-slate-900/30 hover:border-slate-800'
                  }`}
                >
                  <h3 className="text-lg font-bold text-slate-100">{movie.title}</h3>
                  <p className="text-slate-500 text-xs mt-1.5">
                    {movie.rows} rows &times; {movie.seats_per_row} seats
                  </p>
                </div>
              )
            })}
          </div>
        )}

        {/* Seat Grid Layout (Placeholder UI for Step 3 selection preview) */}
        {selectedMovie && (
          <div className="w-full max-w-2xl mx-auto flex flex-col items-center mt-6">
            {/* Screen Line */}
            <div className="w-full text-center text-xs text-slate-500 uppercase tracking-widest mb-1.5 font-semibold">
              Screen
            </div>
            <div className="w-full h-[2px] bg-gradient-to-r from-transparent via-sky-500 to-transparent rounded-full shadow-[0_0_12px_rgba(14,165,233,0.6)] mb-10" />

            {/* Static Grid (Step 3 Layout Preview) */}
            <div className="flex flex-col gap-3 mb-8">
              {Array.from({ length: selectedMovie.rows }).map((_, rowIndex) => {
                const rowLetter = String.fromCharCode(65 + rowIndex) // A, B, C...
                return (
                  <div key={rowLetter} className="flex items-center gap-4">
                    {/* Left Row Label */}
                    <span className="w-5 text-slate-600 font-bold text-center text-xs">{rowLetter}</span>

                    {/* Seat Row */}
                    <div className="flex gap-2.5">
                      {Array.from({ length: selectedMovie.seats_per_row }).map((_, seatIndex) => {
                        const seatNum = seatIndex + 1
                        return (
                          <button
                            key={seatIndex}
                            disabled
                            className="w-9 h-9 rounded-lg flex items-center justify-center text-xs font-semibold bg-slate-900 text-slate-600 border border-slate-800/40 cursor-not-allowed"
                          >
                            {seatNum}
                          </button>
                        )
                      })}
                    </div>

                    {/* Right Row Label */}
                    <span className="w-5 text-slate-600 font-bold text-center text-xs">{rowLetter}</span>
                  </div>
                )
              })}
            </div>

            {/* Legend */}
            <div className="flex flex-wrap gap-6 justify-center text-xs font-semibold text-slate-500 mt-4">
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 rounded bg-slate-900 border border-slate-850" />
                <span>Available</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 rounded bg-yellow-500" />
                <span>Your hold</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 rounded bg-orange-500" />
                <span>Other hold</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="w-4 h-4 rounded bg-red-500" />
                <span>Confirmed</span>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  )
}

export default App
