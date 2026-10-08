const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? '';

export class ApiError extends Error {
	constructor(
		message: string,
		public readonly status: number,
	) {
		super(message);
		this.name = 'ApiError';
	}
}

export async function apiClient<T>(
	endpoint: string,
	options: RequestInit = {},
): Promise<T> {
	let response: Response;

	try {
		response = await fetch(`${API_BASE_URL}${endpoint}`, {
			...options,
			headers: {
				Accept: 'application/json',
				'Content-Type': 'application/json',
				...options.headers,
			},
		});
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Unknown network error';
		throw new ApiError(`Network request failed: ${message}`, 0);
	}

	if (!response.ok) {
		throw new ApiError(await errorMessage(response), response.status);
	}

	if (response.status === 204) {
		return undefined as T;
	}

	return response.json() as Promise<T>;
}

/**
 * Reads the { code, message } body the API returns on error responses,
 * falling back to a generic message when there is no JSON body.
 */
async function errorMessage(response: Response): Promise<string> {
	try {
		const body = (await response.json()) as { message?: unknown };
		if (typeof body?.message === 'string' && body.message.trim() !== '') {
			return body.message;
		}
	} catch {
		// not JSON; keep the generic message
	}
	return `API request failed with status ${response.status}`;
}
