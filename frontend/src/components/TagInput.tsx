import { useState, type KeyboardEvent } from 'react';

interface TagInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  max?: number;
  placeholder?: string;
}

export default function TagInput({ value, onChange, max = 10, placeholder = 'Add a tag…' }: TagInputProps) {
  const [input, setInput] = useState('');

  function addTag(raw: string) {
    const tag = raw.trim().toLowerCase();
    if (!tag) return;
    if (value.includes(tag)) {
      setInput('');
      return;
    }
    if (value.length >= max) return;
    onChange([...value, tag]);
    setInput('');
  }

  function removeTag(index: number) {
    onChange(value.filter((_, i) => i !== index));
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      addTag(input);
    }
    if (e.key === 'Backspace' && !input && value.length > 0) {
      removeTag(value.length - 1);
    }
  }

  const atMax = value.length >= max;

  return (
    <div>
      <div className="flex flex-wrap gap-2">
        {value.map((tag, i) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 rounded-full bg-amber-100 px-3 py-1 text-sm font-medium text-amber-800 dark:bg-amber-900/30 dark:text-amber-300"
          >
            {tag}
            <button
              type="button"
              onClick={() => removeTag(i)}
              className="ml-0.5 inline-flex h-4 w-4 items-center justify-center rounded-full text-amber-600 hover:bg-amber-200 hover:text-amber-800 dark:text-amber-400 dark:hover:bg-amber-800 dark:hover:text-amber-200"
              aria-label={`Remove ${tag}`}
            >
              ×
            </button>
          </span>
        ))}
      </div>
      {!atMax && (
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="mt-2 block w-full rounded-md border border-stone-300 px-3 py-2 text-sm shadow-sm focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-800 dark:text-stone-100"
        />
      )}
      {atMax && (
        <p className="mt-1 text-xs text-stone-500 dark:text-stone-400">
          Maximum of {max} tags reached
        </p>
      )}
    </div>
  );
}
