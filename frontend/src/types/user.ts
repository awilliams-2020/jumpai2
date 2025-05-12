export interface User {
  id: number;
  email: string;
  name: string;
  profile_picture: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  google_id?: string;
  hubspot_id?: string | null;
  last_login_at: string;
} 