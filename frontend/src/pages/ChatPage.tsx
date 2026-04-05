import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useChannels } from '../hooks/useChannels';
import ChannelList from '../components/chat/ChannelList';
import MessageList from '../components/chat/MessageList';
import MessageInput from '../components/chat/MessageInput';
import CreateChannelModal from '../components/chat/CreateChannelModal';

export default function ChatPage() {
  const { channelId } = useParams();
  const navigate = useNavigate();
  const { data: channels } = useChannels();
  const [showCreate, setShowCreate] = useState(false);

  // Auto-select first channel if none selected (desktop)
  useEffect(() => {
    if (!channelId && channels && channels.length > 0) {
      navigate(`/chat/${channels[0].id}`, { replace: true });
    }
  }, [channelId, channels, navigate]);

  const selectedChannel = channels?.find((c) => c.id === channelId);

  return (
    <div className="-mx-4 -my-6 flex h-[calc(100vh-3.5rem)]">
      {/* Sidebar — hidden on mobile when a channel is selected */}
      <div
        className={`w-full shrink-0 border-r border-stone-200 bg-stone-50 py-4 dark:border-stone-700 dark:bg-stone-900/50 md:block md:w-64 ${
          channelId ? 'hidden' : 'block'
        }`}
      >
        <div className="mb-3 px-4">
          <h2 className="text-lg font-semibold text-stone-900 dark:text-stone-100">Chat</h2>
        </div>
        <ChannelList onCreateClick={() => setShowCreate(true)} />
      </div>

      {/* Message area */}
      <div
        className={`flex min-w-0 flex-1 flex-col ${
          channelId ? 'flex' : 'hidden md:flex'
        }`}
      >
        {selectedChannel ? (
          <>
            <div className="flex items-center gap-3 border-b border-stone-200 px-4 py-3 dark:border-stone-700">
              {/* Back button on mobile */}
              <button
                onClick={() => navigate('/chat')}
                className="rounded-md p-1 text-stone-500 hover:bg-stone-100 md:hidden dark:text-stone-400 dark:hover:bg-stone-800"
              >
                <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
                </svg>
              </button>
              <div>
                <h3 className="text-sm font-semibold text-stone-900 dark:text-stone-100">
                  {selectedChannel.type === 'general' ? '#' : ''}{selectedChannel.name}
                </h3>
                <p className="text-xs text-stone-500 dark:text-stone-400">
                  {selectedChannel.type === 'session' ? 'Session chat' : 'General channel'}
                </p>
              </div>
            </div>
            <MessageList channelId={channelId!} />
            <MessageInput channelId={channelId!} />
          </>
        ) : (
          <div className="flex flex-1 items-center justify-center text-stone-400 dark:text-stone-500">
            <div className="text-center">
              <svg className="mx-auto mb-3 h-12 w-12" fill="none" viewBox="0 0 24 24" strokeWidth={1} stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" d="M8.625 12a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H8.25m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0H12m4.125 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 0 1-2.555-.337A5.972 5.972 0 0 1 5.41 20.97a5.969 5.969 0 0 1-.474-.065 4.48 4.48 0 0 0 .978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25Z" />
              </svg>
              <p className="text-sm">Select a channel to start chatting</p>
            </div>
          </div>
        )}
      </div>

      <CreateChannelModal open={showCreate} onClose={() => setShowCreate(false)} />
    </div>
  );
}
