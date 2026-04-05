import { Link, useParams } from 'react-router-dom';
import { useChannels } from '../../hooks/useChannels';
import { useAuth } from '../../context/AuthContext';
import type { Channel } from '../../types';

const HashIcon = (
  <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 8.25h15m-16.5 7.5h15m-1.8-13.5-3.9 19.5m-2.1-19.5-3.9 19.5" />
  </svg>
);

const CalendarChatIcon = (
  <svg className="h-4 w-4 shrink-0" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
    <path strokeLinecap="round" strokeLinejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
  </svg>
);

function ChannelItem({ channel }: { channel: Channel }) {
  const { channelId } = useParams();
  const active = channelId === channel.id;

  return (
    <Link
      to={`/chat/${channel.id}`}
      className={`flex items-center gap-2 rounded-md px-3 py-1.5 text-sm transition-colors ${
        active
          ? 'bg-amber-50 text-amber-700 font-medium dark:bg-amber-900/30 dark:text-amber-300'
          : 'text-stone-600 hover:bg-stone-100 dark:text-stone-400 dark:hover:bg-stone-800'
      }`}
    >
      {channel.type === 'session' ? CalendarChatIcon : HashIcon}
      <span className="truncate">{channel.name}</span>
    </Link>
  );
}

interface ChannelListProps {
  onCreateClick?: () => void;
}

export default function ChannelList({ onCreateClick }: ChannelListProps) {
  const { data: channels, isLoading } = useChannels();
  const { user } = useAuth();

  if (isLoading) {
    return (
      <div className="p-4 text-sm text-stone-500 dark:text-stone-400">Loading channels...</div>
    );
  }

  const general = channels?.filter((c) => c.type === 'general') ?? [];
  const session = channels?.filter((c) => c.type === 'session') ?? [];

  return (
    <div className="flex flex-col gap-4">
      <div>
        <div className="flex items-center justify-between px-3 pb-1">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-stone-400 dark:text-stone-500">
            Channels
          </h3>
          {user?.is_admin && onCreateClick && (
            <button
              onClick={onCreateClick}
              className="text-stone-400 hover:text-stone-600 dark:text-stone-500 dark:hover:text-stone-300"
              title="Create channel"
            >
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
              </svg>
            </button>
          )}
        </div>
        <div className="space-y-0.5">
          {general.map((c) => (
            <ChannelItem key={c.id} channel={c} />
          ))}
        </div>
      </div>

      {session.length > 0 && (
        <div>
          <h3 className="px-3 pb-1 text-xs font-semibold uppercase tracking-wider text-stone-400 dark:text-stone-500">
            Sessions
          </h3>
          <div className="space-y-0.5">
            {session.map((c) => (
              <ChannelItem key={c.id} channel={c} />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
