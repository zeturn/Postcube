export interface User {
  id: number;
  basalt_id: string;
  email: string;
  name: string;
  slug: string;
  box_title: string;
  created_at: string;
  updated_at: string;
}

export interface Question {
  id: number;
  anonymous_name: string;
  content: string;
  answer: string;
  status: 'pending' | 'answered';
  background_color: string;
  created_at: string;
  answered_at: string | null;
}

export interface PublicBoxResponse {
  owner: {
    name: string;
    slug: string;
    box_title: string;
  };
  answered: Question[];
  unanswered: Question[];
}

export interface MyBoxResponse {
  user: User;
  stats: {
    total: number;
    answered: number;
    unanswered: number;
  };
}

async function parseError(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string };
    return data.error || `Request failed: ${res.status}`;
  } catch {
    return `Request failed: ${res.status}`;
  }
}

async function requestJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    credentials: 'include',
    ...init,
  });

  if (!response.ok) {
    throw new Error(await parseError(response));
  }

  return response.json() as Promise<T>;
}

export function loginUrl() {
  return '/api/auth/login';
}

export async function fetchMe(): Promise<User> {
  return requestJSON<User>('/api/auth/me');
}

export async function logout(): Promise<void> {
  await requestJSON<{ message: string }>('/api/auth/logout', {
    method: 'POST',
  });
}

export async function fetchPublicBox(slug: string): Promise<PublicBoxResponse> {
  return requestJSON<PublicBoxResponse>(`/api/public/box/${encodeURIComponent(slug)}`);
}

export async function submitQuestion(
  slug: string,
  payload: { content: string; anonymous_name?: string },
): Promise<Question> {
  return requestJSON<Question>(`/api/public/box/${encodeURIComponent(slug)}/questions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

export async function fetchInbox(): Promise<Question[]> {
  return requestJSON<Question[]>('/api/inbox');
}

export async function updateInboxQuestion(
  id: number,
  payload: { answer?: string; background_color?: string },
): Promise<Question> {
  return requestJSON<Question>(`/api/inbox/questions/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}

export async function deleteInboxQuestion(id: number): Promise<void> {
  await requestJSON<{ message: string }>(`/api/inbox/questions/${id}`, {
    method: 'DELETE',
  });
}

export async function fetchMyBox(): Promise<MyBoxResponse> {
  return requestJSON<MyBoxResponse>('/api/box/me');
}

export async function updateMyBox(payload: { box_title: string }): Promise<User> {
  return requestJSON<User>('/api/box/me', {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });
}
