import { useCallback, useEffect, useRef } from 'react';
import { useMessages } from '../../hooks/useMessages';
import { useAuth } from '../../context/AuthContext';
import MessageBubble from './MessageBubble';
import type { ChatMessage } from '../../types';

interface MessageListProps {
  channelId: string;
}

export default function MessageList({ channelId }: MessageListProps) {
  const { user } = useAuth();
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } =
    useMessages(channelId);
  const containerRef = useRef<HTMLDivElement>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const prevMessageCountRef = useRef(0);

  // Flatten messages from all pages (newest first in each page, but we reverse for display)
  const allMessages: ChatMessage[] = [];
  if (data) {
    for (let i = data.pages.length - 1; i >= 0; i--) {
      const page = data.pages[i];
      // Each page has messages newest-first, reverse them for chronological order
      const reversed = [...page.messages].reverse();
      allMessages.push(...reversed);
    }
  }

  // Auto-scroll to bottom on new messages (only if already near bottom)
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const messageCount = allMessages.length;
    const isNewMessage = messageCount > prevMessageCountRef.current;
    prevMessageCountRef.current = messageCount;

    if (isNewMessage) {
      const isNearBottom =
        container.scrollHeight - container.scrollTop - container.clientHeight < 100;
      if (isNearBottom) {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
      }
    }
  }, [allMessages.length]);

  // Scroll to bottom on channel change
  useEffect(() => {
    prevMessageCountRef.current = 0;
    bottomRef.current?.scrollIntoView();
  }, [channelId]);

  // Infinite scroll: load older messages when scrolling to top
  const handleScroll = useCallback(() => {
    const container = containerRef.current;
    if (!container || !hasNextPage || isFetchingNextPage) return;
    if (container.scrollTop < 100) {
      fetchNextPage();
    }
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  if (isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center text-sm text-stone-500 dark:text-stone-400">
        Loading messages...
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="flex flex-1 flex-col gap-1 overflow-y-auto px-4 py-3"
    >
      {isFetchingNextPage && (
        <div className="py-2 text-center text-xs text-stone-400 dark:text-stone-500">
          Loading older messages...
        </div>
      )}
      {hasNextPage && !isFetchingNextPage && (
        <button
          onClick={() => fetchNextPage()}
          className="py-2 text-center text-xs text-amber-600 hover:text-amber-700 dark:text-amber-400"
        >
          Load older messages
        </button>
      )}
      {allMessages.length === 0 ? (
        <div className="flex flex-1 items-center justify-center text-sm text-stone-400 dark:text-stone-500">
          No messages yet. Start the conversation!
        </div>
      ) : (
        allMessages.map((msg, i) => {
          const prevMsg = i > 0 ? allMessages[i - 1] : null;
          const showAuthor = !prevMsg || prevMsg.member_id !== msg.member_id;
          return (
            <MessageBubble
              key={msg.id}
              message={msg}
              showAuthor={showAuthor}
              isOwn={msg.member_id === user?.id}
            />
          );
        })
      )}
      <div ref={bottomRef} />
    </div>
  );
}
