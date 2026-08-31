import React, { useState, useEffect } from 'react'
import './App.css'

function App() {
  const [movies, setMovies] = useState([])
  const [selectedMovie, setSelectedMovie] = useState(null)
  const [seatStatuses, setSeatStatuses] = useState([])
  const [activeSession, setActiveSession] = useState(null)
  const [timerSeconds, setTimerSeconds] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Generate a persistent 12-character hex User ID (unique per tab using sessionStorage)
  const [userID] = useState(() => {
    const saved = sessionStorage.getItem('cinema_userID')
    if (saved) return saved
    const newID = Math.random().toString(16).substring(2, 14)
    sessionStorage.setItem('cinema_userID', newID)
    return newID
  })

  // Fetch movies on mount
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

  // Connect EventSource to stream seat updates in real-time
  useEffect(() => {
    if (!selectedMovie) {
      setSeatStatuses([])
      return
    }

    // Open persistent SSE stream connection
    const eventSource = new EventSource(`/movies/${selectedMovie.id}/seats/stream`)

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        setSeatStatuses(data || [])
      } catch (err) {
        console.error('Failed to parse stream event data:', err)
      }
    }

    eventSource.onerror = (err) => {
      console.error('SSE Connection failed:', err)
    }

    return () => {
      // Close stream connection on unmount/movie selection change
      eventSource.close()
    }
  }, [selectedMovie])

  // Handle hold countdown timer
  useEffect(() => {
    if (!activeSession) {
      setTimerSeconds(0)
      return
    }

    const updateTimer = () => {
      const remainingMs = activeSession.expiresAt.getTime() - Date.now()
      const remainingSec = Math.max(0, Math.floor(remainingMs / 1000))
      
      setTimerSeconds(remainingSec)

      // Auto-expire hold on client side if timer reaches 0
      if (remainingSec <= 0) {
        setActiveSession(null)
        alert('Your hold session has expired!')
      }
    }

    updateTimer()
    const timerInterval = setInterval(updateTimer, 1000)
    return () => clearInterval(timerInterval)
  }, [activeSession])

  // Helper to format total seconds into MM:SS
  const formatTime = (totalSeconds) => {
    const minutes = Math.floor(totalSeconds / 60)
    const seconds = totalSeconds % 60
    return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  }

  // Handle available seat clicks to request a hold from Go API
  const handleSeatClick = (seatID, status) => {
    if (status !== 'available') return

    if (activeSession) {
      alert('You are already holding a seat! Release it first to hold another.')
      return
    }

    fetch(`/movies/${selectedMovie.id}/seats/${seatID}/hold`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ user_id: userID }),
    })
      .then((res) => {
        if (!res.ok) {
          return res.json().then((data) => {
            throw new Error(data.error || 'Failed to hold seat')
          })
        }
        return res.json()
      })
      .then((data) => {
        setActiveSession({
          sessionID: data.session_id,
          movieID: data.movie_id,
          seatID: data.seat_id,
          expiresAt: new Date(data.expires_at),
        })
      })
      .catch((err) => {
        alert(err.message)
      })
  }

  // Confirm the seat booking permanently (calls PUT /sessions/:id/confirm?movie_id=...)
  const handleConfirm = () => {
    if (!activeSession) return

    fetch(`/sessions/${activeSession.sessionID}/confirm?movie_id=${activeSession.movieID}`, {
      method: 'PUT',
    })
      .then((res) => {
        if (!res.ok) {
          return res.json().then((data) => {
            throw new Error(data.error || 'Failed to confirm booking')
          })
        }
        setActiveSession(null)
        alert('Booking confirmed successfully!')
      })
      .catch((err) => {
        alert(err.message)
      })
  }

  // Release the seat hold immediately (calls DELETE /sessions/:id?movie_id=...)
  const handleRelease = () => {
    if (!activeSession) return

    fetch(`/sessions/${activeSession.sessionID}?movie_id=${activeSession.movieID}`, {
      method: 'DELETE',
    })
      .then((res) => {
        if (!res.ok) {
          return res.json().then((data) => {
            throw new Error(data.error || 'Failed to release hold')
          })
        }
        setActiveSession(null)
      })
      .catch((err) => {
        alert(err.message)
      })
  }

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
                  onClick={() => {
                    setSelectedMovie(movie)
                    setActiveSession(null)
                  }}
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

        {/* Seat Grid Layout */}
        {selectedMovie && (
          <div className="w-full max-w-2xl mx-auto flex flex-col items-center mt-6">
            {/* Screen Line */}
            <div className="w-full text-center text-xs text-slate-500 uppercase tracking-widest mb-1.5 font-semibold">
              Screen
            </div>
            <div className="w-full h-[2px] bg-gradient-to-r from-transparent via-sky-500 to-transparent rounded-full shadow-[0_0_12px_rgba(14,165,233,0.6)] mb-10" />

            {/* Dynamic Grid */}
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
                        const seatID = `${rowLetter}${seatNum}`

                        // Determine seat status based on backend data
                        const seatData = seatStatuses.find(s => s.seat_id === seatID)
                        let status = 'available'
                        if (seatData) {
                          if (seatData.confirmed) {
                            status = 'confirmed'
                          } else if (seatData.booked) {
                            status = seatData.user_id === userID ? 'held' : 'other_hold'
                          }
                        }

                        const isDisabled = status === 'confirmed' || status === 'other_hold'

                        return (
                          <button
                            key={seatID}
                            disabled={isDisabled}
                            onClick={() => handleSeatClick(seatID, status)}
                            className={`w-9 h-9 rounded-lg flex items-center justify-center text-xs font-semibold transition-all ${
                              status === 'available'
                                ? 'bg-slate-900 text-slate-500 hover:text-sky-400 border border-slate-800/40 hover:border-sky-500/30'
                                : status === 'held'
                                ? 'bg-yellow-500 text-slate-950 font-bold'
                                : status === 'other_hold'
                                ? 'bg-orange-500 text-white font-bold cursor-not-allowed opacity-80'
                                : 'bg-red-500 text-white font-bold cursor-not-allowed'
                            }`}
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

            {/* Checkout Panel (Confirm and Release actions wired up) */}
            {activeSession && (
              <div className="w-full max-w-md bg-slate-900 border border-slate-800 p-5 rounded-xl flex items-center justify-between mt-10 shadow-xl">
                <div>
                  <div className="text-sm font-semibold text-slate-300">
                    Holding Seat <span className="text-sky-400 font-bold">{activeSession.seatID}</span>
                  </div>
                  <div className="text-xs text-slate-500 mt-1.5">
                    Time remaining: <span className="text-yellow-500 font-bold">{formatTime(timerSeconds)}</span>
                  </div>
                </div>
                <div className="flex gap-3">
                  <button
                    onClick={handleConfirm}
                    className="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-slate-50 text-xs font-bold rounded-lg transition"
                  >
                    Confirm
                  </button>
                  <button
                    onClick={handleRelease}
                    className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-bold rounded-lg transition"
                  >
                    Release
                  </button>
                </div>
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  )
}

export default App
