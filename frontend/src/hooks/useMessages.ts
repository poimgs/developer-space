import { useCallback, useEffect } from 'react';
import { useInfiniteQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { useWebSocket } from '../context/WebSocketContext';
import type { ChatMessage, MessagePage, WSMessage } from '../types';

export function useMessages(channelId: string | undefined) {
  const queryClient = useQueryClient();
  const { subscribe } = useWebSocket();

  const query = useInfiniteQuery({
    queryKey: ['messages', channelId],
    queryFn: async ({ pageParam }) => {
      if (!channelId) throw new Error('No channel selected');
      const res = await api.listMessages(channelId, pageParam);
      return res.data;
    },
    getNextPageParam: (lastPage: MessagePage) => lastPage.cursor ?? undefined,
    initialPageParam: undefined as string | undefined,
    enabled: !!channelId,
  });

  const handleWSMessage = useCallback(
    (wsMsg: WSMessage) => {
      if (wsMsg.type !== 'message' || !channelId) return;
      const msg = wsMsg.payload as ChatMessage;
      if (msg.channel_id !== channelId) return;

      queryClient.setQueryData(
        ['messages', channelId],
        (old: typeof query.data) => {
          if (!old) return old;
          const firstPage = old.pages[0];
          if (!firstPage) return old;
          // Avoid duplicates
          if (firstPage.messages.some((m) => m.id === msg.id)) return old;
          return {
            ...old,
            pages: [
              { ...firstPage, messages: [msg, ...firstPage.messages] },
              ...old.pages.slice(1),
            ],
          };
        },
      );
    },
    [channelId, queryClient, query.data],
  );

  useEffect(() => {
    return subscribe(handleWSMessage);
  }, [subscribe, handleWSMessage]);

  return query;
}
