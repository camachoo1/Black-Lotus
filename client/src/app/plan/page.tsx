'use client';

import { useState } from 'react';

import { FaCalendarAlt } from 'react-icons/fa';
import { format } from 'date-fns';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Calendar } from '@/components/ui/calendar';
import { DateRange } from 'react-day-picker';

export default function TripForm() {
  const [range, setRange] = useState<DateRange | undefined>();
  const [destination, setDestination] = useState('');

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
        value={destination}
        onChange={(e) => setDestination(e.target.value)}
        className='w-full border rounded-lg px-4 py-3 text-sm placeholder-gray-400 font-medium'
      />
      {!destination && (
        <p className='text-xs text-red-500 -mt-4'>
          Choose a destination to start planning
        </p>
      )}

      {/* DATE PICKER */}
      <div className='flex gap-2 justify-between border rounded-lg px-4 py-3 text-left cursor-pointer'>
        <div className='flex flex-col gap-1 text-sm text-gray-700'>
          <label className='text-xs font-semibold text-gray-500 mb-1 block'>
            Dates (Optional)
          </label>
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant='ghost'
                className='px-0 py-0 h-auto text-left text-gray-600 hover:bg-transparent hover:cursor-pointer'
              >
                <div className='flex items-center gap-6'>
                  {/* Start Date */}
                  <div className='flex items-center gap-2'>
                    <FaCalendarAlt className='text-gray-500' />
                    <span className='font-medium text-gray-900'>
                      {range?.from
                        ? format(range.from, 'MMMM d')
                        : 'Start date'}
                    </span>
                  </div>

                  {/* Separator */}
                  <div className='text-gray-300'>|</div>

                  {/* End Date */}
                  <div className='flex items-center gap-2'>
                    <FaCalendarAlt className='text-gray-500' />
                    <span className='font-medium text-gray-900'>
                      {range?.to
                        ? format(range.to, 'MMMM d')
                        : 'End date'}
                    </span>
                  </div>
                </div>
              </Button>
            </PopoverTrigger>
            <PopoverContent
              className='w-auto p-0'
              align='center'
              sideOffset={8}
            >
              <Calendar
                initialFocus
                mode='range'
                selected={range}
                onSelect={setRange}
                numberOfMonths={2}
                disabled={(date) => date <= new Date()}
              />
              <div className='p-2 text-right'>
                <button
                  className='text-xs text-gray-500 hover:text-red-500 hover:underline'
                  onClick={() => setRange(undefined)}
                >
                  Clear Dates
                </button>
              </div>
            </PopoverContent>
          </Popover>
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
