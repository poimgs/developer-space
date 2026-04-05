import { useQuery } from '@tanstack/react-query';
import { api } from '../api/client';

export function useChannels() {
  return useQuery({
    queryKey: ['channels'],
    queryFn: async () => {
      const res = await api.listChannels();
      return res.data;
    },
  });
}
