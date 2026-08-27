import React from 'react'
import { Tv, Grid } from 'lucide-react'

function MovieGrid({ movies, onSelectMovie, loading, error }) {
  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-20 gap-4">
        <div className="w-12 h-12 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin"></div>
        <p className="text-slate-400 font-medium animate-pulse">Loading movies...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-20">
        <p className="text-red-400 font-semibold bg-red-500/10 border border-red-500/20 px-6 py-3 rounded-xl inline-block">
          Error: {error}
        </p>
      </div>
    )
  }

  // Helper to generate a unique colorful gradient for mock posters
  const getGradient = (id) => {
    const gradients = [
      'from-blue-600 to-indigo-900',
      'from-purple-600 to-violet-900',
      'from-rose-600 to-pink-900',
    ]
    const idx = id.charCodeAt(id.length - 1) % gradients.length
    return gradients[idx]
  }

  return (
    <div className="w-full">
      <div className="text-center max-w-xl mx-auto mb-12">
        <h2 className="text-3xl font-extrabold tracking-tight sm:text-4xl text-slate-50">
          Select a Movie
        </h2>
        <p className="mt-4 text-base text-slate-400">
          Browse active screenings and choose a film to select your seats. Holds are kept safe for 2 minutes.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl mx-auto">
        {movies.map((movie) => (
          <div
            key={movie.id}
            className="group relative bg-slate-900/40 border border-slate-900 hover:border-indigo-500/40 rounded-2xl overflow-hidden transition-all duration-300 hover:shadow-xl hover:shadow-indigo-500/5 flex flex-col cursor-pointer"
            onClick={() => onSelectMovie(movie)}
          >
            {/* Visual Poster Placeholder */}
            <div className={`h-48 bg-gradient-to-br ${getGradient(movie.id)} flex items-center justify-center relative overflow-hidden`}>
              <div className="absolute inset-0 bg-slate-950/20 backdrop-blur-[1px]" />
              <Tv className="w-16 h-16 text-white/20 group-hover:scale-110 transition-transform duration-300 relative z-10" />
            </div>

            {/* Info Block */}
            <div className="p-6 flex-1 flex flex-col justify-between">
              <div>
                <h3 className="text-xl font-bold text-slate-50 group-hover:text-indigo-400 transition-colors">
                  {movie.title}
                </h3>
                <div className="flex items-center gap-2 mt-3 text-slate-400 text-sm">
                  <Grid className="w-4 h-4 text-slate-500" />
                  <span>
                    {movie.rows} Rows &times; {movie.seats_per_row} Seats
                  </span>
                </div>
              </div>

              <button className="mt-6 w-full py-2.5 bg-slate-800 group-hover:bg-indigo-600 text-slate-200 group-hover:text-white rounded-xl text-sm font-semibold transition-all duration-300">
                Book Tickets
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default MovieGrid
