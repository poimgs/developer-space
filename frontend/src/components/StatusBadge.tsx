const styles: Record<string, string> = {
  scheduled:
    'bg-status-scheduled-bg dark:bg-status-scheduled-bg-dark text-status-scheduled',
  shifted:
    'bg-status-shifted-bg dark:bg-status-shifted-bg-dark text-status-shifted',
  canceled:
    'bg-status-canceled-bg dark:bg-status-canceled-bg-dark text-status-canceled',
};

const labels: Record<string, string> = {
  scheduled: 'Scheduled',
  shifted: 'Rescheduled',
  canceled: 'Canceled',
};

interface StatusBadgeProps {
  status: string;
}

export default function StatusBadge({ status }: StatusBadgeProps) {
  return (
    <span
      className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${styles[status] ?? ''}`}
    >
      {labels[status] ?? status}
    </span>
  );
}
