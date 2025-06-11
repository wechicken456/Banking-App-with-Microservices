import type { LoginCredentials, RegisterCredentials, LoginResponse, User } from '$lib/types/auth';

const API_BASE_URL = '/api';

class ApiError extends Error {
    constructor(
        message: string,
        public status: number,
        public code?: string
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

async function fetchApi<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {

    const url = `${API_BASE_URL}${endpoint}`;
  
    const response = await fetch(url, {
        ...options,
        headers: {
        'Content-Type': 'application/json',
        ...options.headers,
        },
        credentials: 'include', // Important for cookies
    });

    if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new ApiError(
            errorData.message || 'An error occurred',
            response.status,
            errorData.code
        );
    }

    return response.json();
}

export const api = {
    async login(credentials: LoginCredentials): Promise<LoginResponse> {
        return fetchApi<LoginResponse>('/login', {
            method: 'POST',
            body: JSON.stringify(credentials),
        });
    },

    async register(credentials: RegisterCredentials): Promise<LoginResponse> {
        return fetchApi<LoginResponse>('/register', {
            method: 'POST',
            body: JSON.stringify(credentials),
        });
    },

    async renewToken(): Promise<{ accessToken: string }> {
        return fetchApi<{ accessToken: string }>('/renew-token', {
            method: 'POST',
        });
    },

    async logout(): Promise<void> {
        // Clear local storage and let the server-side handle cookie cleanup
        if (typeof window !== 'undefined') {
            localStorage.removeItem('access_token');
            sessionStorage.removeItem('access_token');
            localStorage.removeItem("refresh_token");
            sessionStorage.removeItem("refresh_token");
        }
    },

    async getProfile(): Promise<User> {
        return fetchApi<User>('/profile', {
            method: 'GET',
        })
    }
};

export { ApiError };