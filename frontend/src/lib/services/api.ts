import type { LoginCredentials, LoginResponse, AuthTokens, User } from '$lib/types/auth';

const API_BASE_URL = 'http://localhost:8080/api';

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

// Generate a UUID for idempotency keys
function generateIdempotencyKey(): string {
    return crypto.randomUUID();
}

async function fetchApi<T>(
    endpoint: string,
    options: RequestInit = {}
): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    
    // Generate idempotency key for non-GET requests
    const needsIdempotencyKey = options.method && options.method !== 'GET';
    const idempotencyKey = needsIdempotencyKey ? generateIdempotencyKey() : undefined;
    
    const response = await fetch(url, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...(idempotencyKey && { 'Idempotency-Key': idempotencyKey }),
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

    async register(credentials: LoginCredentials): Promise<LoginResponse> {
        return fetchApi<LoginResponse>('/register', {
            method: 'POST',
            body: JSON.stringify(credentials),
        });
    },

    async renewToken(authTokens : AuthTokens): Promise<{accessToken: string}> {
        return fetchApi<{accessToken : string }>('/renew-token', {
            method: 'POST',
            body: JSON.stringify(authTokens),
        });
    },

    async logout(): Promise<void> {
        // Clear local storage and let the server-side handle cookie cleanup
        if (typeof window !== 'undefined') {
            sessionStorage.removeItem('access_token');
            sessionStorage.removeItem('access_token');
            sessionStorage.removeItem("refresh_token");
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