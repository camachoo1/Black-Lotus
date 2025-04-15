'use client';

import { motion } from 'framer-motion';

export function Hero() {
  return (
    // MAIN CONTAINER
    <div className='flex flex-col py-24 px-4 justify-center items-center'>
      {/* INITIAL CONTAINER FOR INTRO */}
      <div className='text-center max-w-4xl'>
        <motion.h1
          className='text-3xl custom-text font-bold p-2'
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 1.5 }}
        >
          That trip you always talked about with your friends but
          never happened just got easier to plan
        </motion.h1>
        <motion.h3
          className='text-gray-500 custom-text text-sm p-2 mt-2'
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1, duration: 1 }}
        >
          Create detailed itineraries, manage all your bookings, find
          great events in the areas your visiting, and share your trip
          with your friends from this one app!
        </motion.h3>

        <motion.button
          className='hover:cursor-pointer custom-text text-lg  bg-cyan-400 hover:bg-cyan-500 rounded-full  px-8 py-3.5 me-2 m-6 font-bold text-white'
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 1.5, duration: 1 }}
        >
          Start Planning
        </motion.button>
      </div>
    </div>
  );
}
