import { useEffect, useRef } from 'react';

export interface DateChip {
  date: string;
  sessionCount: number;
}

interface DateStripProps {
  dates: DateChip[];
  selected: string;
  onSelect: (date: string) => void;
}

function formatChipDay(dateStr: string): string {
  const d = new Date(dateStr + 'T00:00:00');
  return d.toLocaleDateString('en-US', { weekday: 'short' });
}

function formatChipDate(dateStr: string): number {
  const d = new Date(dateStr + 'T00:00:00');
  return d.getDate();
}

export default function DateStrip({ dates, selected, onSelect }: DateStripProps) {
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

  if (dates.length === 0) return null;

  return (
    <div
      ref={containerRef}
      className="flex gap-2 overflow-x-auto py-2 px-1 scrollbar-hide"
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
  );
}
