export function Hero() {
  return (
    // MAIN CONTAINER
    <div className='flex py-24 px-4 justify-center items-center'>
      {/* INITIAL CONTAINER FOR INTRO */}
      <div className='text-center max-w-4xl'>
        <h1 className='text-3xl custom-text font-bold p-2'>
          That trip you always talked about with your friends but
          never happened just got easier to plan
        </h1>
        <h3 className='text-gray-500 custom-text text-sm p-2'>
          Create detailed itineraries, manage all your bookings, find
          great events in the areas your visiting, and share your trip
          with your friends from this one app!
        </h3>

        <button className='custom-text text-lg focus:outline-none bg-cyan-400 hover:bg-cyan-500 rounded-full focus:ring-4 px-8 py-3.5 me-2 mb-2 font-bold text-white'>
          Start Planning
        </button>
      </div>
    </div>
  );
}
