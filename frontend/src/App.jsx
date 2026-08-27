import React, { useState } from 'react'
import { Film } from 'lucide-react'
import './App.css'

function App() {
  const [selectedMovie, setSelectedMovie] = useState(null)

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
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8 flex flex-col justify-center items-center">
        {selectedMovie ? (
          <div className="text-center">
            <button 
              onClick={() => setSelectedMovie(null)}
              className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 rounded-lg text-sm font-medium transition"
            >
              Back to Movies
            </button>
            <p className="mt-4">Movie details will load here in Step 3.</p>
          </div>
        ) : (
          <div className="text-center max-w-lg">
            <h2 className="text-3xl font-extrabold tracking-tight sm:text-4xl text-slate-50">
              Book Your Experience
            </h2>
            <p className="mt-4 text-lg text-slate-400">
              Choose a film to begin your seat reservation. Temporary holds are valid for 2 minutes.
            </p>
            <button 
              onClick={() => setSelectedMovie({ id: 'movie-1', title: 'The Matrix' })}
              className="mt-6 px-6 py-3 bg-indigo-600 hover:bg-indigo-700 rounded-xl font-semibold shadow-lg shadow-indigo-500/25 transition-all"
            >
              Select Movie (Step 2 Demo)
            </button>
          </div>
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
