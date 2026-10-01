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
	const response = await fetch(`${API_BASE_URL}${endpoint}`, {
		...options,
		headers: {
			Accept: 'application/json',
			'Content-Type': 'application/json',
			...options.headers,
		},
	});

	if (!response.ok) {
		throw new ApiError(
			`API request failed with status ${response.status}`,
			response.status,
		);
	}

	if (response.status === 204) {
		return undefined as T;
	}

	return response.json() as Promise<T>;
}
