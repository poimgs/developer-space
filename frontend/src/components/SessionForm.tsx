import { useState, type FormEvent } from 'react';
import type { CreateSessionRequest, SpaceSession, UpdateSessionRequest } from '../types';

interface SessionFormProps {
  session?: SpaceSession;
  onSubmit: (data: CreateSessionRequest | UpdateSessionRequest) => Promise<void>;
  loading?: boolean;
  hideDate?: boolean;
  children?: React.ReactNode;
}

interface FormErrors {
  title?: string;
  date?: string;
  start_time?: string;
  end_time?: string;
  repeat_weekly?: string;
}

export default function SessionForm({ session, onSubmit, loading, hideDate, children }: SessionFormProps) {
  const isEdit = !!session;

  const [title, setTitle] = useState(session?.title ?? '');
  const [description, setDescription] = useState(session?.description ?? '');
  const [date, setDate] = useState(session?.date ?? '');
  const [startTime, setStartTime] = useState(session?.start_time ?? '');
  const [endTime, setEndTime] = useState(session?.end_time ?? '');
  const [location, setLocation] = useState(session?.location ?? '');
  const [repeatMode, setRepeatMode] = useState<'none' | 'weekly' | 'forever'>('none');
  const [repeatCount, setRepeatCount] = useState(1);
  const [dayOfWeek, setDayOfWeek] = useState<number | null>(null);
  const [everyNWeeks, setEveryNWeeks] = useState(1);
  const [errors, setErrors] = useState<FormErrors>({});

  function validate(): FormErrors {
    const errs: FormErrors = {};
    if (!title.trim()) errs.title = 'Title is required';
    if (!date) errs.date = 'Date is required';
    if (!startTime) errs.start_time = 'Start time is required';
    if (!endTime) errs.end_time = 'End time is required';
    if (startTime && endTime && endTime <= startTime) {
      errs.end_time = 'End time must be after start time';
    }
    if (!isEdit && repeatMode === 'weekly' && (repeatCount < 1 || repeatCount > 12)) {
      errs.repeat_weekly = 'Repeat count must be between 1 and 12';
    }
    return errs;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const errs = validate();
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }
    setErrors({});

    if (isEdit) {
      const data: UpdateSessionRequest = {};
      if (title !== session.title) data.title = title.trim();
      if ((description || '') !== (session.description || '')) data.description = description.trim() || null;
      if (date !== session.date) data.date = date;
      if (startTime !== session.start_time) data.start_time = startTime;
      if (endTime !== session.end_time) data.end_time = endTime;
      if ((location || '') !== (session.location || '')) data.location = location.trim() || null;
      await onSubmit(data);
    } else {
      const data: CreateSessionRequest = {
        title: title.trim(),
        date,
        start_time: startTime,
        end_time: endTime,
      };
      if (description.trim()) data.description = description.trim();
      if (location.trim()) data.location = location.trim();
      if (repeatMode === 'weekly') data.repeat_weekly = repeatCount;
      if (repeatMode === 'forever') data.repeat_forever = true;
      if (repeatMode !== 'none') {
        if (dayOfWeek !== null) data.day_of_week = dayOfWeek;
        if (everyNWeeks > 1) data.every_n_weeks = everyNWeeks;
      }
      await onSubmit(data);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="title" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
          Title <span className="text-red-500">*</span>
        </label>
        <input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
        />
        {errors.title && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.title}</p>}
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
          Description
        </label>
        <textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          className="mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
        />
      </div>

      <div>
        <label htmlFor="location" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
          Location
        </label>
        <div className="relative mt-1">
          <svg className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-stone-400 dark:text-stone-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
          </svg>
          <input
            id="location"
            type="text"
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="e.g. 123 Main St, Suite 4B"
            className="block w-full rounded-md border border-stone-300 pl-9 pr-3 py-2 text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
          />
        </div>
      </div>

      {!hideDate && (
        <div>
          <label htmlFor="date" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
            Date <span className="text-red-500">*</span>
          </label>
          <input
            id="date"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            className="mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
          />
          {errors.date && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.date}</p>}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label htmlFor="start_time" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
            Start time <span className="text-red-500">*</span>
          </label>
          <input
            id="start_time"
            type="time"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            className="mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
          />
          {errors.start_time && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.start_time}</p>}
        </div>
        <div>
          <label htmlFor="end_time" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
            End time <span className="text-red-500">*</span>
          </label>
          <input
            id="end_time"
            type="time"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            className="mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
          />
          {errors.end_time && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.end_time}</p>}
        </div>
      </div>

      {!isEdit && (
        <div className="space-y-2">
          <span className="block text-sm font-medium text-stone-700 dark:text-stone-300">Repeat</span>
          <div className="space-y-1">
            <label className="flex items-center gap-2">
              <input
                type="radio"
                name="repeat"
                checked={repeatMode === 'none'}
                onChange={() => setRepeatMode('none')}
                className="h-4 w-4 border-stone-300 text-amber-600 focus:ring-amber-500"
              />
              <span className="text-sm text-stone-700 dark:text-stone-300">No repeat</span>
            </label>
            <label className="flex items-center gap-2">
              <input
                type="radio"
                name="repeat"
                checked={repeatMode === 'weekly'}
                onChange={() => setRepeatMode('weekly')}
                className="h-4 w-4 border-stone-300 text-amber-600 focus:ring-amber-500"
              />
              <span className="text-sm text-stone-700 dark:text-stone-300">Repeat weekly for</span>
            </label>
            {repeatMode === 'weekly' && (
              <div className="ml-6 flex items-center gap-2">
                <input
                  type="number"
                  min="1"
                  max="12"
                  value={repeatCount}
                  onChange={(e) => setRepeatCount(parseInt(e.target.value, 10) || 1)}
                  className="w-16 rounded-md border border-stone-300 px-2 py-1 text-sm text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
                />
                <span className="text-sm text-stone-600 dark:text-stone-400">weeks</span>
                {errors.repeat_weekly && <p className="text-xs text-red-600 dark:text-red-400">{errors.repeat_weekly}</p>}
              </div>
            )}
            <label className="flex items-center gap-2">
              <input
                type="radio"
                name="repeat"
                checked={repeatMode === 'forever'}
                onChange={() => setRepeatMode('forever')}
                className="h-4 w-4 border-stone-300 text-amber-600 focus:ring-amber-500"
              />
              <span className="text-sm text-stone-700 dark:text-stone-300">Repeat forever</span>
            </label>
          </div>

          {repeatMode !== 'none' && (
            <div className="ml-6 space-y-3 pt-1">
              <div>
                <label htmlFor="day_of_week" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                  Recur on
                </label>
                <select
                  id="day_of_week"
                  value={dayOfWeek === null ? '' : dayOfWeek}
                  onChange={(e) => setDayOfWeek(e.target.value === '' ? null : parseInt(e.target.value, 10))}
                  className="mt-1 block w-full rounded-md border border-stone-300 px-3 py-2 text-sm text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
                >
                  <option value="">Same day as start date</option>
                  <option value="0">Sunday</option>
                  <option value="1">Monday</option>
                  <option value="2">Tuesday</option>
                  <option value="3">Wednesday</option>
                  <option value="4">Thursday</option>
                  <option value="5">Friday</option>
                  <option value="6">Saturday</option>
                </select>
              </div>

              <div>
                <label htmlFor="every_n_weeks" className="block text-sm font-medium text-stone-700 dark:text-stone-300">
                  Every
                </label>
                <div className="mt-1 flex items-center gap-2">
                  <select
                    id="every_n_weeks"
                    value={everyNWeeks}
                    onChange={(e) => setEveryNWeeks(parseInt(e.target.value, 10))}
                    className="block w-20 rounded-md border border-stone-300 px-3 py-2 text-sm text-stone-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-700 dark:text-stone-100"
                  >
                    <option value="1">1</option>
                    <option value="2">2</option>
                    <option value="3">3</option>
                    <option value="4">4</option>
                  </select>
                  <span className="text-sm text-stone-600 dark:text-stone-400">
                    {everyNWeeks === 1 ? 'week' : 'weeks'}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {children}

      <div className="flex justify-end pt-2">
        <button
          type="submit"
          disabled={loading}
          className="rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amber-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Saving...' : isEdit ? 'Save Changes' : 'Create Session'}
        </button>
      </div>
    </form>
  );
}
