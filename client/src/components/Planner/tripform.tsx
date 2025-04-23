'use client';

import { useState } from 'react';
import { FaCalendarAlt } from 'react-icons/fa';
import { format } from 'date-fns';
import { useForm, Controller, FormProvider } from 'react-hook-form';

import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Calendar } from '@/components/ui/calendar';
import { DateRange } from 'react-day-picker';
import { cn } from '@/lib/utils';

// Define the form data type
interface TripFormData {
  destination: string;
  dateRange: DateRange | undefined;
}

export default function TripForm() {
  const [isFocused, setIsFocused] = useState(false);

  const methods = useForm<TripFormData>({
    defaultValues: {
      destination: '',
      dateRange: undefined,
    },
    mode: 'onSubmit', // Validate on form submission
  });

  const {
    register,
    control,
    handleSubmit,
    watch,
    // setValue, // Will need once backend is setup with external API's
    formState: { errors, isSubmitting },
  } = methods;

  // Watch the destination field for conditional rendering
  const destination = watch('destination');

  // Form submission handler
  const onSubmit = (data: TripFormData) => {
    console.log('Form submitted:', data);
    // Add data later on
  };

  return (
    <div className='max-w-lg mx-auto mt-24 px-4 text-center space-y-6'>
      <h1 className='custom-text text-2xl font-bold text-gray-900'>
        Plan a new trip
      </h1>

      <FormProvider {...methods}>
        <form onSubmit={handleSubmit(onSubmit)} className='space-y-6'>
          {/* Destination Input with Floating Label */}
          <div className='relative w-full'>
            <input
              id='destination'
              type='text'
              placeholder=' '
              {...register('destination', {
                required: 'Please enter a destination',
              })}
              onFocus={() => setIsFocused(true)}
              onBlur={() => setIsFocused(false)}
              className={cn(
                'custom-text peer w-full border rounded-lg px-4 py-3 pt-6 text-sm text-gray-900 placeholder-transparent font-medium',
                errors.destination ? 'border-red-500' : ''
              )}
            />

            {/* Floating label */}
            <label
              htmlFor='destination'
              className={cn(
                'absolute left-4 text-sm font-medium transition-all duration-200 pointer-events-none',
                isFocused || destination
                  ? '-top-2 text-[0.7rem] bg-white px-1 left-3'
                  : 'top-4',
                errors.destination ? 'text-red-500' : 'text-gray-500'
              )}
            >
              Destination
            </label>

            {/* Inline placeholder */}
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

            {/* Error message for destination */}
            {errors.destination && (
              <p className='text-red-500 text-xs mt-1 text-left'>
                {errors.destination.message}
              </p>
            )}
          </div>

          {/* Date Picker using Controller to integrate with react-hook-form */}
          <div className='w-full text-left'>
            <label className='text-xs font-semibold text-gray-500 mb-1 block'>
              Dates
            </label>

            <Controller
              name='dateRange'
              control={control}
              rules={{
                required: 'Please select your trip dates',
                validate: (value) => {
                  // Properly typed validation function
                  if (!value) return 'Please select your trip dates';
                  if (!value.from || !value.to)
                    return 'Please select both start and end dates';
                  return true; // Valid
                },
              }}
              render={({ field }) => (
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      type='button'
                      variant='ghost'
                      className={cn(
                        'w-full h-auto px-4 py-3 justify-start border rounded-lg hover:bg-transparent text-left text-gray-600',
                        errors.dateRange ? 'border-red-500' : ''
                      )}
                    >
                      <div className='flex items-center gap-6 w-full'>
                        {/* Start date display */}
                        <div className='flex items-center gap-2'>
                          <FaCalendarAlt
                            className={cn(
                              errors.dateRange
                                ? 'text-red-500'
                                : 'text-gray-500'
                            )}
                          />
                          <span
                            className={cn(
                              'custom-text font-medium',
                              errors.dateRange
                                ? 'text-red-500'
                                : 'text-gray-900'
                            )}
                          >
                            {field.value?.from
                              ? format(field.value.from, 'MMMM d')
                              : 'Start date'}
                          </span>
                        </div>

                        {/* Separator */}
                        <div
                          className={cn(
                            errors.dateRange
                              ? 'text-red-300'
                              : 'text-gray-300'
                          )}
                        >
                          |
                        </div>

                        {/* End date display */}
                        <div className='flex items-center gap-2'>
                          <FaCalendarAlt
                            className={cn(
                              errors.dateRange
                                ? 'text-red-500'
                                : 'text-gray-500'
                            )}
                          />
                          <span
                            className={cn(
                              'custom-text font-medium',
                              errors.dateRange
                                ? 'text-red-500'
                                : 'text-gray-900'
                            )}
                          >
                            {field.value?.to
                              ? format(field.value.to, 'MMMM d')
                              : 'End date'}
                          </span>
                        </div>
                      </div>
                    </Button>
                  </PopoverTrigger>

                  <PopoverContent
                    className='w-full p-0'
                    align='start'
                    sideOffset={8}
                  >
                    <Calendar
                      initialFocus
                      mode='range'
                      selected={field.value}
                      onSelect={field.onChange}
                      numberOfMonths={2}
                      disabled={(date) => date <= new Date()}
                    />

                    {/* Clear selection option */}
                    <div className='p-2 text-right'>
                      <button
                        type='button'
                        className='text-xs text-gray-500 hover:text-red-500 hover:underline'
                        onClick={() => field.onChange(undefined)}
                      >
                        Clear Dates
                      </button>
                    </div>
                  </PopoverContent>
                </Popover>
              )}
            />

            {/* Error message for date range */}
            {errors.dateRange && (
              <p className='text-red-500 text-xs mt-1'>
                {errors.dateRange.message}
              </p>
            )}
          </div>

          {/* Invite option */}
          <div className='flex justify-between text-sm text-gray-500'>
            <button
              type='button'
              className='custom-text flex items-center gap-1 cursor-pointer'
            >
              <span>+</span>Invite Friends To Your Trip
            </button>
          </div>

          {/* Submit button */}
          <Button
            type='submit'
            className='w-full bg-cyan-400 hover:bg-cyan-500 text-white py-3 rounded-lg cursor-pointer'
            disabled={isSubmitting}
          >
            {isSubmitting
              ? 'Planning Your Trip...'
              : 'Plan Your Trip'}
          </Button>
        </form>
      </FormProvider>
    </div>
  );
}
