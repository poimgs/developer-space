import { useEffect, useRef } from 'react';

interface CancelScopeModalProps {
  open: boolean;
  onThisOnly: () => void;
  onAllSessions: () => void;
  onCancel: () => void;
}

export default function CancelScopeModal({ open, onThisOnly, onAllSessions, onCancel }: CancelScopeModalProps) {
  const thisOnlyRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (open) thisOnlyRef.current?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onCancel();
    };
    document.addEventListener('keydown', handleKey);
    return () => document.removeEventListener('keydown', handleKey);
  }, [open, onCancel]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={onCancel}
      role="dialog"
      aria-modal="true"
      aria-label="Cancel scope"
    >
      <div
        className="mx-4 w-full max-w-md rounded-xl bg-white p-6 shadow-xl dark:bg-stone-800"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Cancel recurring session</h2>
        <p className="mt-2 text-sm text-stone-600 dark:text-stone-400">
          This session is part of a recurring series. What would you like to cancel?
        </p>
        <div className="mt-6 flex flex-col gap-3">
          <button
            ref={thisOnlyRef}
            onClick={onThisOnly}
            className="w-full rounded-md border border-stone-300 px-4 py-2 text-sm font-medium text-stone-700 transition-colors hover:bg-stone-50 dark:border-stone-600 dark:text-stone-300 dark:hover:bg-stone-700"
          >
            Cancel this session only
          </button>
          <button
            onClick={onAllSessions}
            className="w-full rounded-md bg-red-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-red-700"
          >
            Cancel all future sessions
          </button>
          <button
            onClick={onCancel}
            className="w-full rounded-md px-4 py-2 text-sm font-medium text-stone-500 transition-colors hover:text-stone-700 dark:text-stone-400 dark:hover:text-stone-200"
          >
            Go back
          </button>
        </div>
      </div>
    </div>
  );
}
