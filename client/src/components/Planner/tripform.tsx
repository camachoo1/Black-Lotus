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
import { cn } from '@/lib/utils';

export default function TripForm() {
  const [range, setRange] = useState<DateRange | undefined>();
  const [destination, setDestination] = useState('');
  const [isFocused, setIsFocused] = useState(false); // for floating label behavior

  return (
    // Main wrapper for the form layout
    <div className='max-w-lg mx-auto mt-24 px-4 text-center space-y-6'>
      <h1 className='custom-text text-2xl font-bold text-gray-900'>
        Plan a new trip
      </h1>

      {/* Destination Input with Floating Label + Animated Placeholder */}
      <div className='relative w-full'>
        <input
          id='destination'
          type='text'
          value={destination}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          onChange={(e) => setDestination(e.target.value)}
          placeholder=' ' // required for peer-placeholder-shown to work
          className='custom-text peer w-full border rounded-lg px-4 py-3 pt-6 text-sm text-gray-900 placeholder-transparent font-medium'
        />

        {/* Floating label: animates to top left on focus or input */}
        <label
          htmlFor='destination'
          className={cn(
            'absolute left-4 text-sm font-medium text-gray-500 transition-all duration-200 pointer-events-none',
            isFocused || destination
              ? '-top-2 text-[0.7rem] bg-white px-1 left-3'
              : 'top-4'
          )}
        >
          Destination
        </label>

        {/* Inline placeholder: visible initially, animates/fades out on focus or input */}
        {!destination && (
          <span
            className={cn(
              'absolute left-[100px] top-4 text-sm text-gray-400 transition-all duration-200 pointer-events-none',
              isFocused
                ? 'opacity-0 -translate-x-2'
                : 'opacity-100 translate-x-0'
            )}
          >
            i.e. London, Barcelona, Tokyo
          </span>
        )}
      </div>

      {/* Date Picker using shadcn/ui Popover + Calendar */}
      <div className='w-full text-left'>
        <label className='text-xs font-semibold text-gray-500 mb-1 block'>
          Dates (Optional)
        </label>

        <Popover>
          {/* The clickable field that opens the calendar */}
          <PopoverTrigger asChild>
            <Button
              variant='ghost'
              className='w-full h-auto px-4 py-3 justify-start border rounded-lg hover:bg-transparent text-left text-gray-600'
            >
              <div className='flex items-center gap-6 w-full'>
                {/* Start date display */}
                <div className='flex items-center gap-2'>
                  <FaCalendarAlt className='text-gray-500' />
                  <span className='custom-text font-medium text-gray-900'>
                    {range?.from
                      ? format(range.from, 'MMMM d')
                      : 'Start date'}
                  </span>
                </div>

                {/* Separator */}
                <div className='text-gray-300'>|</div>

                {/* End date display */}
                <div className='flex items-center gap-2'>
                  <FaCalendarAlt className='text-gray-500' />
                  <span className='custom-text font-medium text-gray-900'>
                    {range?.to
                      ? format(range.to, 'MMMM d')
                      : 'End date'}
                  </span>
                </div>
              </div>
            </Button>
          </PopoverTrigger>

          {/* Calendar popover content */}
          <PopoverContent
            className='w-full p-0'
            align='start'
            sideOffset={8}
          >
            <Calendar
              initialFocus
              mode='range'
              selected={range}
              onSelect={setRange}
              numberOfMonths={2}
              disabled={(date) => date <= new Date()} // block past dates
            />

            {/* Clear selection option */}
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

      {/* Invite option (additional CTA) */}
      <div className='flex justify-between text-sm text-gray-500'>
        <button className='custom-text flex items-center gap-1 cursor-pointer'>
          <span>+</span>Invite Friends To Your Trip
        </button>
      </div>
    </div>
  );
}
