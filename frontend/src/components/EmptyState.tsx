import type { ReactNode } from 'react';

interface EmptyStateProps {
  icon?: ReactNode;
  heading: string;
  subtext?: string;
  action?: { label: string; onClick: () => void };
}

export default function EmptyState({ icon, heading, subtext, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      {icon && <div className="mb-4 text-stone-400 dark:text-stone-500">{icon}</div>}
      <h3 className="text-lg font-semibold text-stone-900 dark:text-stone-100">{heading}</h3>
      {subtext && <p className="mt-1 text-sm text-stone-500 dark:text-stone-400">{subtext}</p>}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 rounded-md bg-amber-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-amber-700 focus:ring-2 focus:ring-amber-500 focus:outline-none"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
