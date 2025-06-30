export interface User {
    id: string;
    email: string;
}

export interface LoginCredentials {
    email: string;
    password: string;
}

export interface AuthTokens {
    accessToken: string;
    refreshToken: string;
}

export interface LoginResponse {
    userId: string;
    email: string;
    accessToken: string; 
    refreshToken: string;
    fingerprint: string; 
    accessTokenDuration: number; // in seconds
    refreshTokenDuration: number; // in seconds
}

export interface RenewAccessTokenResponse {
    accessToken: string;
}

export interface ApiError {
    message: string;
    code?: string;
}


