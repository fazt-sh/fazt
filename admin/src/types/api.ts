// API Response types matching the standardized format

export interface APIResponse<T> {
  data: T;
}

export interface APIListResponse<T> {
  data: T[];
  meta?: {
    total: number;
    limit: number;
    offset: number;
  };
}

export interface APIError {
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export class APIErrorClass extends Error {
  code: string;
  details?: Record<string, unknown>;

  constructor(error: APIError['error']) {
    super(error.message);
    this.code = error.code;
    this.details = error.details;
  }
}
