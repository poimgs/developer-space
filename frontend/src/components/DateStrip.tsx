import { useEffect, useRef } from 'react';

export interface DateChip {
  date: string;
  sessionCount: number;
}

interface DateStripProps {
  dates: DateChip[];
  selected: string;
  onSelect: (date: string) => void;
  monthLabel: string;
  onPrevMonth: () => void;
  onNextMonth: () => void;
  prevDisabled: boolean;
  nextDisabled: boolean;
}

function formatChipDay(dateStr: string): string {
  const d = new Date(dateStr + 'T00:00:00');
  return d.toLocaleDateString('en-US', { weekday: 'short' });
}

function formatChipDate(dateStr: string): number {
  const d = new Date(dateStr + 'T00:00:00');
  return d.getDate();
}

export default function DateStrip({ dates, selected, onSelect, monthLabel, onPrevMonth, onNextMonth, prevDisabled, nextDisabled }: DateStripProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const selectedRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (selectedRef.current && containerRef.current) {
      const container = containerRef.current;
      const chip = selectedRef.current;
      const scrollLeft = chip.offsetLeft - container.clientWidth / 2 + chip.clientWidth / 2;
      container.scrollTo?.({ left: scrollLeft, behavior: 'smooth' });
    }
  }, [selected]);

  return (
    <div>
      {/* Month header */}
      <div className="flex items-center justify-center gap-3 py-2">
        <button
          onClick={onPrevMonth}
          disabled={prevDisabled}
          aria-label="Previous month"
          className="rounded p-1 text-stone-500 transition-colors hover:text-stone-800 disabled:opacity-30 disabled:cursor-not-allowed dark:text-stone-400 dark:hover:text-stone-200"
        >
          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <span className="text-sm font-semibold text-stone-700 dark:text-stone-200 min-w-[10rem] text-center">
          {monthLabel}
        </span>
        <button
          onClick={onNextMonth}
          disabled={nextDisabled}
          aria-label="Next month"
          className="rounded p-1 text-stone-500 transition-colors hover:text-stone-800 disabled:opacity-30 disabled:cursor-not-allowed dark:text-stone-400 dark:hover:text-stone-200"
        >
          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      {/* Date chips or empty message */}
      {dates.length === 0 ? (
        <p className="py-4 text-center text-sm text-stone-400 dark:text-stone-500">
          No sessions this month
        </p>
      ) : (
        <div
          ref={containerRef}
          className="flex justify-center gap-2 overflow-x-auto py-2 px-1 scrollbar-hide"
          role="tablist"
          aria-label="Session dates"
        >
          {dates.map((chip) => {
            const isSelected = chip.date === selected;
            return (
              <button
                key={chip.date}
                ref={isSelected ? selectedRef : undefined}
                role="tab"
                aria-selected={isSelected}
                onClick={() => onSelect(chip.date)}
                className={`flex flex-col items-center px-3 py-2 rounded-xl text-sm cursor-pointer min-w-[3.5rem] transition-colors ${
                  isSelected
                    ? 'bg-amber-500 text-white shadow-md dark:bg-amber-500 dark:text-white'
                    : 'bg-stone-100 text-stone-600 hover:bg-stone-200 dark:bg-stone-800 dark:text-stone-400 dark:hover:bg-stone-700'
                }`}
              >
                <span className="text-xs font-medium uppercase">{formatChipDay(chip.date)}</span>
                <span className="text-lg font-bold">{formatChipDate(chip.date)}</span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
