import { useToast } from '../context/ToastContext';

const typeStyles = {
  success: 'border-green-500 bg-green-50 dark:bg-green-900/20 text-green-800 dark:text-green-200',
  error: 'border-red-500 bg-red-50 dark:bg-red-900/20 text-red-800 dark:text-red-200',
  info: 'border-amber-500 bg-amber-50 dark:bg-amber-900/20 text-amber-800 dark:text-amber-200',
};

export default function Toast() {
  const { toasts, removeToast } = useToast();

  if (toasts.length === 0) return null;

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2" role="status" aria-live="polite">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className={`flex items-center gap-2 rounded-md border-l-4 px-4 py-3 shadow-md ${typeStyles[toast.type]}`}
        >
          <span className="flex-1 text-sm font-medium">{toast.message}</span>
          <button
            onClick={() => removeToast(toast.id)}
            className="text-current opacity-50 hover:opacity-100"
            aria-label="Dismiss"
          >
            &times;
          </button>
        </div>
      ))}
    </div>
  );
}
