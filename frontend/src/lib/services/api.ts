import type { LoginCredentials, LoginResponse, AuthTokens, User } from '$lib/types/auth';
import type { Account } from '$lib/types/account';

const API_BASE_URL = 'http://localhost:8080/api';

const PROTECTED_ENDPOINTS = ['/profile', '/get-all-accounts', '/get-account', '/delete-account']; 
function isProtectedEndpoint(endpoint: string): boolean {
    return PROTECTED_ENDPOINTS.some(protectedEndpoint => endpoint.includes(protectedEndpoint));
}

// Token provider interface for dependency injection.
// This allows us to separate the token retrieval logic from the API service
interface TokenProvider {
    getAccessToken(): string | null;
}

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

    // Get access token for protected endpoints
    const accessToken = isProtectedEndpoint(endpoint) ? tokenProvider?.getAccessToken() : null;

    const response = await fetch(url, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            ...(idempotencyKey && { 'Idempotency-Key': idempotencyKey }),
            ...(accessToken && { 'Authorization': `Bearer ${accessToken}` }),
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

let tokenProvider: TokenProvider | null = null;
export const api = {
    tokenProvider,
    setTokenProvider(provider: TokenProvider) {
        tokenProvider = provider;
    },

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

    async renewToken(authTokens: AuthTokens): Promise<{ accessToken: string }> {
        return fetchApi<{ accessToken: string }>('/renew-token', {
            method: 'POST',
            body: JSON.stringify(authTokens),
        });
    },

    async logout(): Promise<void> {
        // Clear local storage and let the server-side handle cookie cleanup
        if (typeof window !== 'undefined') {
            sessionStorage.removeItem('accessToken');
            sessionStorage.removeItem("refreshToken");
        }
    },

    async getProfile(): Promise<User> {
        return fetchApi<User>('/profile', {
            method: 'GET',
        })
    },

    async getAccounts(): Promise<Account[]> {
        return fetchApi<Account[]>('/get-all-accounts', {
            method: 'GET',
        })
    },

    async getAccount(accountNumber: number): Promise<Account[]> {
        return fetchApi<Account[]>(`/get-account?accountNumber=${encodeURIComponent(accountNumber)}`, {
            method: 'GET',
        })
    },

    async deleteAccount(accountNumber: number): Promise< {success : boolean }> {
        return fetchApi<{ success : boolean }>('/delete-account', {
            method: 'DELETE',
            body: JSON.stringify({ 'accountNumber': accountNumber })
        })
    },
};

export { ApiError };
