import React, { useState, useEffect } from 'react'
import { Film } from 'lucide-react'
import MovieGrid from './components/MovieGrid'
import './App.css'

function App() {
  const [movies, setMovies] = useState([])
  const [selectedMovie, setSelectedMovie] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // Fetch movies from our Go backend on component mount
  useEffect(() => {
    fetch('/movies')
      .then((res) => {
        if (!res.ok) {
          throw new Error('Failed to load screenings')
        }
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
    <div className="min-h-screen bg-slate-950 text-slate-100 flex flex-col font-sans">
      {/* Premium Header */}
      <header className="border-b border-slate-900 bg-slate-950/80 backdrop-blur-md sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2.5 bg-indigo-600/10 border border-indigo-500/20 rounded-xl text-indigo-400">
              <Film className="w-6 h-6" />
            </div>
            <div>
              <span className="text-xl font-bold tracking-tight bg-gradient-to-r from-indigo-400 to-violet-400 bg-clip-text text-transparent">
                Cinema Booking
              </span>
              <span className="ml-2 text-xs font-semibold px-2 py-0.5 rounded-full bg-indigo-500/10 border border-indigo-500/25 text-indigo-300">
                Fullstack
              </span>
            </div>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-slate-400">Secure Ticket Desk</span>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-12 flex flex-col justify-center items-center">
        {selectedMovie ? (
          <div className="text-center">
            <h2 className="text-2xl font-bold text-slate-50">{selectedMovie.title}</h2>
            <button 
              onClick={() => setSelectedMovie(null)}
              className="mt-6 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 rounded-xl text-sm font-medium transition"
            >
              Back to Movies
            </button>
            <p className="mt-4 text-slate-400">Seat map for this movie will load here in Step 4.</p>
          </div>
        ) : (
          <MovieGrid 
            movies={movies} 
            onSelectMovie={setSelectedMovie} 
            loading={loading} 
            error={error} 
          />
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 py-6 text-center text-xs text-slate-500">
        &copy; {new Date().getFullYear()} Concurrent Cinema Booking. Built with Go, Redis, and React.
      </footer>
    </div>
  )
}

export default App
