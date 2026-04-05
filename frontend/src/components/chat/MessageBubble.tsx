import type { ChatMessage } from '../../types';

function formatTime(iso: string): string {
  const date = new Date(iso);
  const now = new Date();
  const isToday = date.toDateString() === now.toDateString();
  if (isToday) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' }) +
    ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

interface MessageBubbleProps {
  message: ChatMessage;
  showAuthor: boolean;
  isOwn: boolean;
}

export default function MessageBubble({ message, showAuthor, isOwn }: MessageBubbleProps) {
  return (
    <div className={`group flex flex-col ${isOwn ? 'items-end' : 'items-start'}`}>
      {showAuthor && (
        <span className="mb-0.5 px-1 text-xs font-medium text-stone-500 dark:text-stone-400">
          {message.author_name}
        </span>
      )}
      <div
        className={`max-w-[75%] rounded-xl px-3 py-2 text-sm ${
          isOwn
            ? 'bg-amber-500 text-white dark:bg-amber-600'
            : 'bg-stone-100 text-stone-900 dark:bg-stone-800 dark:text-stone-100'
        }`}
      >
        <p className="whitespace-pre-wrap break-words">{message.content}</p>
      </div>
      <span className="mt-0.5 px-1 text-[10px] text-stone-400 opacity-0 transition-opacity group-hover:opacity-100 dark:text-stone-500">
        {formatTime(message.created_at)}
      </span>
    </div>
  );
}
