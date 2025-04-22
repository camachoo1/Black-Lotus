export default function Home() {
  return (
    // MAIN WRAPPER
    <div className='max-w-lg mx-auto mt-24 px-4 text-center space-y-6'>
      <h1 className='custom-text text-2xl font-bold text-gray-900'>
        Plan a new trip
      </h1>

      {/* DESTINATION PICKER */}
      <input
        type='text'
        placeholder='London, Barcelona, Tokyo'
        className='w-full border rounded-lg px-4 py-3 text-sm placeholder-gray-400 font-medium'
      />

      {/* DATE PICKER */}
      <div className='flex gap-2 items-center justify-between border rounded-lg px-4 py-3'>
        <div className='text-left'>
          <label className='text-xs font-semibold text-gray-500 mb-1 block'>
            Dates (optional)
          </label>
          <div className='flex gap-2 text-sm text-gray-600'>
            <input
              type='date'
              className='text-sm border-none focus:ring-0'
            />
            <span className='text-gray-400'>to</span>
            <input
              type='date'
              className='text-sm border-none focus:ring-0'
            />
          </div>
        </div>
      </div>

      {/* EXTRA OPTIONS */}
      <div className='flex justify-between text-sm text-gray-500'>
        <button className='flex items-center gap-1 hover:underline'>
          <span>+</span>Invite Friends To Your Trip
        </button>
      </div>
    </div>
  );
}
