import { useRef, useState, type KeyboardEvent } from 'react';

interface TagInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  max?: number;
  placeholder?: string;
  suggestions?: string[];
}

export default function TagInput({ value, onChange, max = 10, placeholder = 'Add a tag…', suggestions = [] }: TagInputProps) {
  const [input, setInput] = useState('');
  const [highlightIndex, setHighlightIndex] = useState(-1);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const listboxRef = useRef<HTMLUListElement>(null);

  const filtered = input.trim()
    ? suggestions.filter(
        (s) => s.toLowerCase().includes(input.trim().toLowerCase()) && !value.includes(s.toLowerCase()),
      )
    : [];

  const dropdownVisible = showSuggestions && filtered.length > 0;

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
    setHighlightIndex(-1);
    setShowSuggestions(false);
  }

  function removeTag(index: number) {
    onChange(value.filter((_, i) => i !== index));
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Escape' && dropdownVisible) {
      e.preventDefault();
      setShowSuggestions(false);
      setHighlightIndex(-1);
      return;
    }

    if (e.key === 'ArrowDown' && dropdownVisible) {
      e.preventDefault();
      setHighlightIndex((prev) => (prev < filtered.length - 1 ? prev + 1 : 0));
      return;
    }

    if (e.key === 'ArrowUp' && dropdownVisible) {
      e.preventDefault();
      setHighlightIndex((prev) => (prev > 0 ? prev - 1 : filtered.length - 1));
      return;
    }

    if (e.key === 'Tab' && dropdownVisible) {
      e.preventDefault();
      const idx = highlightIndex >= 0 ? highlightIndex : 0;
      addTag(filtered[idx]);
      return;
    }

    if (e.key === 'Enter') {
      e.preventDefault();
      if (dropdownVisible && highlightIndex >= 0) {
        addTag(filtered[highlightIndex]);
      } else {
        addTag(input);
      }
      return;
    }

    if (e.key === ',') {
      e.preventDefault();
      addTag(input);
      return;
    }

    if (e.key === 'Backspace' && !input && value.length > 0) {
      removeTag(value.length - 1);
    }
  }

  const activeDescendant = dropdownVisible && highlightIndex >= 0 ? `tag-suggestion-${highlightIndex}` : undefined;
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
        <div className="relative">
          <input
            type="text"
            role="combobox"
            aria-expanded={dropdownVisible}
            aria-controls="tag-suggestions-listbox"
            aria-activedescendant={activeDescendant}
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              setHighlightIndex(-1);
              setShowSuggestions(true);
            }}
            onFocus={() => setShowSuggestions(true)}
            onBlur={() => {
              setShowSuggestions(false);
              setHighlightIndex(-1);
            }}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            className="mt-2 block w-full rounded-md border border-stone-300 px-3 py-2 text-sm shadow-sm focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none dark:border-stone-600 dark:bg-stone-800 dark:text-stone-100"
          />
          {dropdownVisible && (
            <ul
              ref={listboxRef}
              id="tag-suggestions-listbox"
              role="listbox"
              className="absolute z-10 mt-1 max-h-48 w-full overflow-auto rounded-md border border-stone-200 bg-white shadow-lg dark:border-stone-600 dark:bg-stone-800"
            >
              {filtered.map((s, i) => (
                <li
                  key={s}
                  id={`tag-suggestion-${i}`}
                  role="option"
                  aria-selected={i === highlightIndex}
                  onMouseDown={(e) => {
                    e.preventDefault();
                    addTag(s);
                  }}
                  className={`cursor-pointer px-3 py-2 text-sm ${
                    i === highlightIndex
                      ? 'bg-amber-100 text-amber-900 dark:bg-amber-900/40 dark:text-amber-200'
                      : 'text-stone-700 hover:bg-stone-100 dark:text-stone-300 dark:hover:bg-stone-700'
                  }`}
                >
                  {s}
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
      {atMax && (
        <p className="mt-1 text-xs text-stone-500 dark:text-stone-400">
          Maximum of {max} tags reached
        </p>
      )}
    </div>
  );
}
