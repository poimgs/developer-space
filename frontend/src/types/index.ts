export interface Member {
  id: string;
  email: string;
  name: string;
  telegram_handle: string | null;
  is_admin: boolean;
  is_active: boolean;
  bio: string | null;
  skills: string[];
  linkedin_url: string | null;
  instagram_handle: string | null;
  github_username: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateMemberRequest {
  email: string;
  name: string;
  telegram_handle?: string | null;
  is_admin: boolean;
  send_invite: boolean;
}

export interface UpdateMemberRequest {
  name?: string;
  telegram_handle?: string | null;
  is_admin?: boolean;
  is_active?: boolean;
}

export interface PublicMember {
  id: string;
  name: string;
  telegram_handle: string | null;
  bio: string | null;
  skills: string[];
  linkedin_url: string | null;
  instagram_handle: string | null;
  github_username: string | null;
}

export interface SpaceSession {
  id: string;
  title: string;
  description: string | null;
  date: string;
  start_time: string;
  end_time: string;
  capacity: number;
  status: 'scheduled' | 'shifted' | 'canceled';
  image_url: string | null;
  location: string | null;
  series_id: string | null;
  created_by: string;
  created_at: string;
  updated_at: string;
  rsvp_count: number;
  user_rsvped: boolean;
}

export interface CreateSessionRequest {
  title: string;
  description?: string | null;
  date: string;
  start_time: string;
  end_time: string;
  capacity: number;
  location?: string | null;
  repeat_weekly?: number;
  repeat_forever?: boolean;
  day_of_week?: number;
  every_n_weeks?: number;
}

export interface UpdateSessionRequest {
  title?: string;
  description?: string | null;
  date?: string;
  start_time?: string;
  end_time?: string;
  capacity?: number;
  location?: string | null;
}

export interface UpdateProfileRequest {
  name?: string;
  telegram_handle?: string;
  bio?: string;
  skills?: string[];
  linkedin_url?: string;
  instagram_handle?: string;
  github_username?: string;
}

export interface RSVP {
  id: string;
  session_id: string;
  member_id: string;
  created_at: string;
}

export interface RSVPWithMember {
  id: string;
  session_id: string;
  member: RSVPMember;
  created_at: string;
}

export interface RSVPMember {
  id: string;
  name: string;
  telegram_handle: string | null;
}

export interface APIResponse<T> {
  data: T;
}

export interface APIError {
  error: string;
  details?: Record<string, string>;
}
