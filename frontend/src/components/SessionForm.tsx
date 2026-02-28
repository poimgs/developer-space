import { useState, type FormEvent } from 'react';
import type { CreateSessionRequest, SpaceSession, UpdateSessionRequest } from '../types';

interface SessionFormProps {
  session?: SpaceSession;
  onSubmit: (data: CreateSessionRequest | UpdateSessionRequest) => Promise<void>;
  loading?: boolean;
}

interface FormErrors {
  title?: string;
  date?: string;
  start_time?: string;
  end_time?: string;
  capacity?: string;
  repeat_weekly?: string;
}

export default function SessionForm({ session, onSubmit, loading }: SessionFormProps) {
  const isEdit = !!session;

  const [title, setTitle] = useState(session?.title ?? '');
  const [description, setDescription] = useState(session?.description ?? '');
  const [date, setDate] = useState(session?.date ?? '');
  const [startTime, setStartTime] = useState(session?.start_time ?? '');
  const [endTime, setEndTime] = useState(session?.end_time ?? '');
  const [capacity, setCapacity] = useState(session?.capacity?.toString() ?? '');
  const [repeatWeekly, setRepeatWeekly] = useState(false);
  const [repeatCount, setRepeatCount] = useState(1);
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
    const cap = parseInt(capacity, 10);
    if (!capacity || isNaN(cap) || cap < 1) errs.capacity = 'Capacity must be at least 1';
    if (!isEdit && repeatWeekly && (repeatCount < 1 || repeatCount > 12)) {
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
      const cap = parseInt(capacity, 10);
      if (cap !== session.capacity) data.capacity = cap;
      await onSubmit(data);
    } else {
      const data: CreateSessionRequest = {
        title: title.trim(),
        date,
        start_time: startTime,
        end_time: endTime,
        capacity: parseInt(capacity, 10),
      };
      if (description.trim()) data.description = description.trim();
      if (repeatWeekly) data.repeat_weekly = repeatCount;
      await onSubmit(data);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label htmlFor="title" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Title <span className="text-red-500">*</span>
        </label>
        <input
          id="title"
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
        />
        {errors.title && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.title}</p>}
      </div>

      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Description
        </label>
        <textarea
          id="description"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
        />
      </div>

      <div>
        <label htmlFor="date" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Date <span className="text-red-500">*</span>
        </label>
        <input
          id="date"
          type="date"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
        />
        {errors.date && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.date}</p>}
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label htmlFor="start_time" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            Start time <span className="text-red-500">*</span>
          </label>
          <input
            id="start_time"
            type="time"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
          />
          {errors.start_time && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.start_time}</p>}
        </div>
        <div>
          <label htmlFor="end_time" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
            End time <span className="text-red-500">*</span>
          </label>
          <input
            id="end_time"
            type="time"
            value={endTime}
            onChange={(e) => setEndTime(e.target.value)}
            className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
          />
          {errors.end_time && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.end_time}</p>}
        </div>
      </div>

      <div>
        <label htmlFor="capacity" className="block text-sm font-medium text-gray-700 dark:text-gray-300">
          Capacity <span className="text-red-500">*</span>
        </label>
        <input
          id="capacity"
          type="number"
          min="1"
          value={capacity}
          onChange={(e) => setCapacity(e.target.value)}
          className="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
        />
        {errors.capacity && <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.capacity}</p>}
      </div>

      {!isEdit && (
        <div className="space-y-2">
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={repeatWeekly}
              onChange={(e) => setRepeatWeekly(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Repeat weekly</span>
          </label>
          {repeatWeekly && (
            <div className="ml-6 flex items-center gap-2">
              <span className="text-sm text-gray-600 dark:text-gray-400">for</span>
              <input
                type="number"
                min="1"
                max="12"
                value={repeatCount}
                onChange={(e) => setRepeatCount(parseInt(e.target.value, 10) || 1)}
                className="w-16 rounded-md border border-gray-300 px-2 py-1 text-sm text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100"
              />
              <span className="text-sm text-gray-600 dark:text-gray-400">weeks</span>
              {errors.repeat_weekly && <p className="text-xs text-red-600 dark:text-red-400">{errors.repeat_weekly}</p>}
            </div>
          )}
        </div>
      )}

      <div className="flex justify-end pt-2">
        <button
          type="submit"
          disabled={loading}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {loading ? 'Saving...' : isEdit ? 'Save Changes' : 'Create Session'}
        </button>
      </div>
    </form>
  );
}
